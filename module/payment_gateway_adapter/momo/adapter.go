package momo

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	gw "github.com/modami/be-payment-service/module/payment_gateway_adapter"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
)

// Config holds MoMo credentials and endpoints.
type Config struct {
	PartnerCode string
	AccessKey   string
	SecretKey   string
	APIEndpoint string
	ReturnURL   string
	NotifyURL   string
}

// Adapter implements PaymentGateway for MoMo.
type Adapter struct {
	cfg Config
}

// New creates a new MoMo Adapter.
func New(cfg Config) *Adapter {
	return &Adapter{cfg: cfg}
}

func (a *Adapter) Name() string { return "momo" }

// CreatePaymentURL creates a MoMo QR/deeplink payment request.
func (a *Adapter) CreatePaymentURL(ctx context.Context, req gw.CreatePaymentRequest) (*gw.PaymentURL, error) {
	requestID := fmt.Sprintf("%d", time.Now().UnixMilli())
	orderInfo := req.Description
	if orderInfo == "" {
		orderInfo = "Payment " + req.OrderRef
	}

	rawSig := fmt.Sprintf("accessKey=%s&amount=%d&extraData=&ipnUrl=%s&orderId=%s&orderInfo=%s&partnerCode=%s&redirectUrl=%s&requestId=%s&requestType=payWithMethod",
		a.cfg.AccessKey,
		req.Amount,
		a.cfg.NotifyURL,
		req.OrderRef,
		orderInfo,
		a.cfg.PartnerCode,
		a.cfg.ReturnURL,
		requestID,
	)
	signature := hmacSHA256(a.cfg.SecretKey, rawSig)

	payload := map[string]interface{}{
		"partnerCode": a.cfg.PartnerCode,
		"accessKey":   a.cfg.AccessKey,
		"requestId":   requestID,
		"amount":      req.Amount,
		"orderId":     req.OrderRef,
		"orderInfo":   orderInfo,
		"redirectUrl": a.cfg.ReturnURL,
		"ipnUrl":      a.cfg.NotifyURL,
		"extraData":   "",
		"requestType": "payWithMethod",
		"signature":   signature,
		"lang":        "vi",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(a.cfg.APIEndpoint+"/create", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.ErrGatewayUnavailable
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, apperrors.ErrGatewayUnavailable
	}

	payURL, _ := result["payUrl"].(string)
	if payURL == "" {
		msg, _ := result["message"].(string)
		return nil, fmt.Errorf("momo error: %s", msg)
	}

	return &gw.PaymentURL{
		URL:      payURL,
		ExpireAt: time.Now().Add(15 * time.Minute).Unix(),
	}, nil
}

// VerifyCallback verifies the MoMo IPN callback signature.
func (a *Adapter) VerifyCallback(ctx context.Context, params map[string]string) (*gw.PaymentResult, error) {
	signature := params["signature"]
	if signature == "" {
		return nil, apperrors.ErrInvalidSignature
	}

	rawSig := fmt.Sprintf("accessKey=%s&amount=%s&extraData=%s&message=%s&orderId=%s&orderInfo=%s&orderType=%s&partnerCode=%s&payType=%s&requestId=%s&responseTime=%s&resultCode=%s&transId=%s",
		a.cfg.AccessKey,
		params["amount"],
		params["extraData"],
		params["message"],
		params["orderId"],
		params["orderInfo"],
		params["orderType"],
		params["partnerCode"],
		params["payType"],
		params["requestId"],
		params["responseTime"],
		params["resultCode"],
		params["transId"],
	)

	expected := hmacSHA256(a.cfg.SecretKey, rawSig)
	if expected != signature {
		return nil, apperrors.ErrInvalidSignature
	}

	resultCode := params["resultCode"]
	success := resultCode == "0"

	result := &gw.PaymentResult{
		Success:     success,
		GatewayTxID: params["transId"],
		RawResponse: params,
	}
	if !success {
		result.FailureReason = "MoMo result code: " + resultCode
	}
	return result, nil
}

// QueryTransaction queries a MoMo payment status.
func (a *Adapter) QueryTransaction(ctx context.Context, txRef string) (*gw.TransactionStatus, error) {
	requestID := fmt.Sprintf("%d", time.Now().UnixMilli())

	rawSig := fmt.Sprintf("accessKey=%s&orderId=%s&partnerCode=%s&requestId=%s",
		a.cfg.AccessKey, txRef, a.cfg.PartnerCode, requestID)
	signature := hmacSHA256(a.cfg.SecretKey, rawSig)

	payload := map[string]interface{}{
		"partnerCode": a.cfg.PartnerCode,
		"requestId":   requestID,
		"orderId":     txRef,
		"signature":   signature,
		"lang":        "vi",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(a.cfg.APIEndpoint+"/query", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.ErrGatewayUnavailable
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	_ = json.Unmarshal(data, &result)

	status := "pending"
	if rc, ok := result["resultCode"].(float64); ok && rc == 0 {
		status = "success"
	}

	return &gw.TransactionStatus{
		Status:      status,
		GatewayTxID: stringVal(result["transId"]),
	}, nil
}

// Refund initiates a MoMo refund.
func (a *Adapter) Refund(ctx context.Context, req gw.RefundRequest) (*gw.RefundResult, error) {
	requestID := fmt.Sprintf("%d", time.Now().UnixMilli())

	rawSig := fmt.Sprintf("accessKey=%s&amount=%d&description=%s&orderId=%s&partnerCode=%s&requestId=%s&transId=%s",
		a.cfg.AccessKey, req.Amount, req.Reason, req.OrderRef, a.cfg.PartnerCode, requestID, req.GatewayTxID)
	signature := hmacSHA256(a.cfg.SecretKey, rawSig)

	payload := map[string]interface{}{
		"partnerCode": a.cfg.PartnerCode,
		"orderId":     req.OrderRef,
		"requestId":   requestID,
		"amount":      req.Amount,
		"transId":     req.GatewayTxID,
		"lang":        "vi",
		"description": req.Reason,
		"signature":   signature,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(a.cfg.APIEndpoint+"/refund", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.ErrGatewayUnavailable
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	_ = json.Unmarshal(data, &result)

	success := false
	if rc, ok := result["resultCode"].(float64); ok && rc == 0 {
		success = true
	}

	return &gw.RefundResult{
		Success:  success,
		RefundID: stringVal(result["transId"]),
	}, nil
}

func hmacSHA256(secret, data string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func stringVal(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
