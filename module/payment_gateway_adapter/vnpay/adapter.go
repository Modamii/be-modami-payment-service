package vnpay

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	gw "github.com/modami/be-payment-service/module/payment_gateway_adapter"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
)

// Config holds VNPay credentials and endpoints.
type Config struct {
	TMNCode    string
	HashSecret string
	PaymentURL string
	ReturnURL  string
	IPNURL     string
	QueryURL   string
}

// Adapter implements PaymentGateway for VNPay.
type Adapter struct {
	cfg Config
}

// New creates a new VNPay Adapter.
func New(cfg Config) *Adapter {
	return &Adapter{cfg: cfg}
}

func (a *Adapter) Name() string { return "vnpay" }

// CreatePaymentURL builds and signs the VNPay redirect URL.
func (a *Adapter) CreatePaymentURL(ctx context.Context, req gw.CreatePaymentRequest) (*gw.PaymentURL, error) {
	now := time.Now().In(time.FixedZone("Asia/Ho_Chi_Minh", 7*3600))
	createDate := now.Format("20060102150405")
	expireDate := now.Add(15 * time.Minute).Format("20060102150405")

	params := url.Values{}
	params.Set("vnp_Version", "2.1.0")
	params.Set("vnp_Command", "pay")
	params.Set("vnp_TmnCode", a.cfg.TMNCode)
	params.Set("vnp_Amount", fmt.Sprintf("%d", req.Amount*100)) // VNPay uses amount * 100
	params.Set("vnp_CreateDate", createDate)
	params.Set("vnp_CurrCode", "VND")
	params.Set("vnp_IpAddr", req.IPAddress)
	params.Set("vnp_Locale", "vn")
	params.Set("vnp_OrderInfo", req.Description)
	params.Set("vnp_OrderType", "other")
	params.Set("vnp_ReturnUrl", a.cfg.ReturnURL)
	params.Set("vnp_TxnRef", req.OrderRef)
	params.Set("vnp_ExpireDate", expireDate)

	// Sort params and build query string for signing.
	signData := buildSignatureData(params)
	signature := hmacSHA512(a.cfg.HashSecret, signData)
	params.Set("vnp_SecureHash", signature)

	paymentURL := a.cfg.PaymentURL + "?" + params.Encode()
	expireTS := now.Add(15 * time.Minute).Unix()

	return &gw.PaymentURL{URL: paymentURL, ExpireAt: expireTS}, nil
}

// VerifyCallback verifies the vnp_SecureHash from VNPay callback params.
func (a *Adapter) VerifyCallback(ctx context.Context, params map[string]string) (*gw.PaymentResult, error) {
	secureHash := params["vnp_SecureHash"]
	if secureHash == "" {
		return nil, apperrors.ErrInvalidSignature
	}

	// Build params without vnp_SecureHash and vnp_SecureHashType.
	filtered := url.Values{}
	for k, v := range params {
		if k != "vnp_SecureHash" && k != "vnp_SecureHashType" {
			filtered.Set(k, v)
		}
	}

	signData := buildSignatureData(filtered)
	expected := hmacSHA512(a.cfg.HashSecret, signData)

	if !strings.EqualFold(expected, secureHash) {
		return nil, apperrors.ErrInvalidSignature
	}

	responseCode := params["vnp_ResponseCode"]
	success := responseCode == "00"

	result := &gw.PaymentResult{
		Success:     success,
		GatewayTxID: params["vnp_TransactionNo"],
		RawResponse: params,
	}
	if !success {
		result.FailureReason = "VNPay response code: " + responseCode
	}
	return result, nil
}

// QueryTransaction calls the VNPay query API for a transaction.
func (a *Adapter) QueryTransaction(ctx context.Context, txRef string) (*gw.TransactionStatus, error) {
	now := time.Now().In(time.FixedZone("Asia/Ho_Chi_Minh", 7*3600))
	createDate := now.Format("20060102150405")

	params := url.Values{}
	params.Set("vnp_RequestId", fmt.Sprintf("%d", now.UnixMilli()))
	params.Set("vnp_Version", "2.1.0")
	params.Set("vnp_Command", "querydr")
	params.Set("vnp_TmnCode", a.cfg.TMNCode)
	params.Set("vnp_TxnRef", txRef)
	params.Set("vnp_OrderInfo", "Query transaction "+txRef)
	params.Set("vnp_TransDate", createDate)
	params.Set("vnp_CreateDate", createDate)
	params.Set("vnp_IpAddr", "127.0.0.1")

	signData := buildSignatureData(params)
	params.Set("vnp_SecureHash", hmacSHA512(a.cfg.HashSecret, signData))

	body, _ := json.Marshal(urlValuesToMap(params))
	resp, err := http.Post(a.cfg.QueryURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.ErrGatewayUnavailable
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, apperrors.ErrGatewayUnavailable
	}

	status := "pending"
	if rc, ok := result["vnp_ResponseCode"].(string); ok && rc == "00" {
		status = "success"
	}

	return &gw.TransactionStatus{
		Status:      status,
		GatewayTxID: stringVal(result["vnp_TransactionNo"]),
	}, nil
}

// Refund initiates a refund via VNPay.
func (a *Adapter) Refund(ctx context.Context, req gw.RefundRequest) (*gw.RefundResult, error) {
	now := time.Now().In(time.FixedZone("Asia/Ho_Chi_Minh", 7*3600))
	createDate := now.Format("20060102150405")

	params := url.Values{}
	params.Set("vnp_RequestId", fmt.Sprintf("%d", now.UnixMilli()))
	params.Set("vnp_Version", "2.1.0")
	params.Set("vnp_Command", "refund")
	params.Set("vnp_TmnCode", a.cfg.TMNCode)
	params.Set("vnp_TransactionType", "02") // Full refund
	params.Set("vnp_TxnRef", req.OrderRef)
	params.Set("vnp_Amount", fmt.Sprintf("%d", req.Amount*100))
	params.Set("vnp_OrderInfo", req.Reason)
	params.Set("vnp_TransDate", createDate)
	params.Set("vnp_CreateDate", createDate)
	params.Set("vnp_CreateBy", "system")
	params.Set("vnp_IpAddr", "127.0.0.1")

	signData := buildSignatureData(params)
	params.Set("vnp_SecureHash", hmacSHA512(a.cfg.HashSecret, signData))

	body, _ := json.Marshal(urlValuesToMap(params))
	resp, err := http.Post(a.cfg.QueryURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.ErrGatewayUnavailable
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	_ = json.Unmarshal(data, &result)

	success := false
	if rc, ok := result["vnp_ResponseCode"].(string); ok && rc == "00" {
		success = true
	}

	return &gw.RefundResult{
		Success:  success,
		RefundID: stringVal(result["vnp_TransactionNo"]),
	}, nil
}

// buildSignatureData sorts url.Values by key and returns the query string for signing.
func buildSignatureData(params url.Values) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params.Get(k))
	}
	return strings.Join(parts, "&")
}

// hmacSHA512 computes the HMAC-SHA512 of data with the given secret.
func hmacSHA512(secret, data string) string {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func urlValuesToMap(v url.Values) map[string]string {
	m := make(map[string]string, len(v))
	for k := range v {
		m[k] = v.Get(k)
	}
	return m
}

func stringVal(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
