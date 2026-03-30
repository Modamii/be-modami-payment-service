package gateway_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modami/be-payment-service/module/core/usecases"
)

// WebhookHandler handles payment gateway callbacks.
type WebhookHandler struct {
	paymentUC *usecases.PaymentUsecase
}

// NewWebhookHandler creates a WebhookHandler.
func NewWebhookHandler(paymentUC *usecases.PaymentUsecase) *WebhookHandler {
	return &WebhookHandler{paymentUC: paymentUC}
}

// VNPayIPN handles POST /api/v1/webhooks/vnpay
func (h *WebhookHandler) VNPayIPN(c *gin.Context) {
	params := collectParams(c)
	if err := h.paymentUC.HandleVNPayCallback(c.Request.Context(), params); err != nil {
		c.JSON(http.StatusOK, gin.H{"RspCode": "99", "Message": err.Error()})
		return
	}
	// VNPay requires exact response format.
	c.JSON(http.StatusOK, gin.H{"RspCode": "00", "Message": "Confirm Success"})
}

// VNPayReturn handles GET /api/v1/payments/return/vnpay
func (h *WebhookHandler) VNPayReturn(c *gin.Context) {
	params := make(map[string]string)
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	if err := h.paymentUC.HandleVNPayCallback(c.Request.Context(), params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Payment processed"})
}

// MoMoIPN handles POST /api/v1/webhooks/momo
func (h *WebhookHandler) MoMoIPN(c *gin.Context) {
	params := collectParams(c)
	if err := h.paymentUC.HandleMoMoCallback(c.Request.Context(), params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// MoMoReturn handles GET /api/v1/payments/return/momo
func (h *WebhookHandler) MoMoReturn(c *gin.Context) {
	params := make(map[string]string)
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	if err := h.paymentUC.HandleMoMoCallback(c.Request.Context(), params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ZaloPayCallback handles POST /api/v1/webhooks/zalopay
func (h *WebhookHandler) ZaloPayCallback(c *gin.Context) {
	params := collectParams(c)
	if err := h.paymentUC.HandleZaloPayCallback(c.Request.Context(), params); err != nil {
		c.JSON(http.StatusOK, gin.H{"return_code": 0, "return_message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"return_code": 1, "return_message": "success"})
}

// ZaloPayReturn handles GET /api/v1/payments/return/zalopay
func (h *WebhookHandler) ZaloPayReturn(c *gin.Context) {
	params := make(map[string]string)
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	if err := h.paymentUC.HandleZaloPayCallback(c.Request.Context(), params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// collectParams reads both JSON body and form values into a flat map.
func collectParams(c *gin.Context) map[string]string {
	params := make(map[string]string)

	// Try JSON body first.
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err == nil {
		for k, v := range body {
			params[k] = stringify(v)
		}
		return params
	}

	// Fall back to form data.
	_ = c.Request.ParseForm()
	for k, v := range c.Request.Form {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}
	return params
}

func stringify(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}
