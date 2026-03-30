package zalopay

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
	"net/url"
	"time"

	gw "github.com/modami/be-payment-service/module/payment_gateway_adapter"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
)

// Config holds ZaloPay credentials and endpoints.
type Config struct {
	AppID       string
	Key1        string
	Key2        string
	CreateURL   string
	QueryURL    string
	RefundURL   string
	CallbackURL string
	ReturnURL   string
}

// Adapter implements PaymentGateway for ZaloPay.
type Adapter struct {
	cfg Config
}

// New creates a new ZaloPay Adapter.
func New(cfg Config) *Adapter {
	return &Adapter{cfg: cfg}
}

func (a *Adapter) Name() string { return "zalopay" }

// CreatePaymentURL creates a ZaloPay order and returns the payment URL.
func (a *Adapter) CreatePaymentURL(ctx context.Context, req gw.CreatePaymentRequest) (*gw.PaymentURL, error) {
	now := time.Now()
	appTime := now.UnixMilli()
	appTransID := fmt.Sprintf("%s_%s", now.Format("060102"), req.OrderRef)

	items, _ := json.Marshal([]interface{}{})
	embedData, _ := json.Marshal(map[string]string{"redirecturl": a.cfg.ReturnURL})

	// ZaloPay signature: hmac_sha256(key1, appid|app_trans_id|appuser|amount|apptime|embeddata|item)
	rawSig := fmt.Sprintf("%s|%s|%s|%d|%d|%s|%s",
		a.cfg.AppID,
		appTransID,
		req.UserID,
		req.Amount,
		appTime,
		string(embedData),
		string(items),
	)
	mac := hmacSHA256(a.cfg.Key1, rawSig)

	payload := map[string]interface{}{
		"app_id":       a.cfg.AppID,
		"app_trans_id": appTransID,
		"app_user":     req.UserID,
		"app_time":     appTime,
		"amount":       req.Amount,
		"item":         string(items),
		"description":  req.Description,
		"embed_data":   string(embedData),
		"callback_url": a.cfg.CallbackURL,
		"mac":          mac,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(a.cfg.CreateURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.ErrGatewayUnavailable
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, apperrors.ErrGatewayUnavailable
	}

	returnCode, _ := result["return_code"].(float64)
	if returnCode != 1 {
		msg, _ := result["return_message"].(string)
		return nil, fmt.Errorf("zalopay error: %s", msg)
	}

	orderURL, _ := result["order_url"].(string)
	return &gw.PaymentURL{
		URL:      orderURL,
		ExpireAt: now.Add(15 * time.Minute).Unix(),
	}, nil
}

// VerifyCallback verifies the ZaloPay callback MAC.
func (a *Adapter) VerifyCallback(ctx context.Context, params map[string]string) (*gw.PaymentResult, error) {
	mac := params["mac"]
	data := params["data"]
	if mac == "" || data == "" {
		return nil, apperrors.ErrInvalidSignature
	}

	expected := hmacSHA256(a.cfg.Key2, data)
	if expected != mac {
		return nil, apperrors.ErrInvalidSignature
	}

	var callbackData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &callbackData); err != nil {
		return nil, apperrors.ErrInvalidSignature
	}

	status, _ := callbackData["return_code"].(float64)
	success := status == 1

	appTransID, _ := callbackData["app_trans_id"].(string)
	zpTransID := stringVal(callbackData["zp_trans_id"])

	result := &gw.PaymentResult{
		Success:     success,
		GatewayTxID: zpTransID,
		RawResponse: map[string]string{
			"app_trans_id": appTransID,
			"zp_trans_id":  zpTransID,
		},
	}
	if !success {
		result.FailureReason = "ZaloPay return_code != 1"
	}
	return result, nil
}

// QueryTransaction queries a ZaloPay transaction status.
func (a *Adapter) QueryTransaction(ctx context.Context, txRef string) (*gw.TransactionStatus, error) {
	now := time.Now()
	appTransID := fmt.Sprintf("%s_%s", now.Format("060102"), txRef)

	rawSig := fmt.Sprintf("%s|%s|%s", a.cfg.AppID, appTransID, a.cfg.Key1)
	mac := hmacSHA256(a.cfg.Key1, rawSig)

	formData := url.Values{}
	formData.Set("app_id", a.cfg.AppID)
	formData.Set("app_trans_id", appTransID)
	formData.Set("mac", mac)

	resp, err := http.PostForm(a.cfg.QueryURL, formData)
	if err != nil {
		return nil, apperrors.ErrGatewayUnavailable
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	_ = json.Unmarshal(data, &result)

	status := "pending"
	if rc, ok := result["return_code"].(float64); ok && rc == 1 {
		status = "success"
	}

	return &gw.TransactionStatus{
		Status:      status,
		GatewayTxID: stringVal(result["zp_trans_id"]),
	}, nil
}

// Refund initiates a ZaloPay refund.
func (a *Adapter) Refund(ctx context.Context, req gw.RefundRequest) (*gw.RefundResult, error) {
	now := time.Now()
	timestamp := now.UnixMilli()
	mRefundID := fmt.Sprintf("%s_%d_%s", now.Format("060102"), timestamp, req.OrderRef)

	rawSig := fmt.Sprintf("%s|%s|%s|%d|%s",
		a.cfg.AppID, mRefundID, req.GatewayTxID, req.Amount, req.Reason)
	mac := hmacSHA256(a.cfg.Key1, rawSig)

	payload := map[string]interface{}{
		"app_id":       a.cfg.AppID,
		"m_refund_id":  mRefundID,
		"timestamp":    timestamp,
		"zp_trans_id":  req.GatewayTxID,
		"amount":       req.Amount,
		"description":  req.Reason,
		"mac":          mac,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(a.cfg.RefundURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.ErrGatewayUnavailable
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	_ = json.Unmarshal(data, &result)

	success := false
	if rc, ok := result["return_code"].(float64); ok && rc == 1 {
		success = true
	}

	return &gw.RefundResult{
		Success:  success,
		RefundID: mRefundID,
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
