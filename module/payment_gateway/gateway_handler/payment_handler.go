package gateway_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/modami/be-payment-service/module/core/dto"
	"github.com/modami/be-payment-service/module/core/usecases"
)

// PaymentHandler handles payment transaction endpoints.
type PaymentHandler struct {
	paymentUC *usecases.PaymentUsecase
}

// NewPaymentHandler creates a PaymentHandler.
func NewPaymentHandler(paymentUC *usecases.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{paymentUC: paymentUC}
}

// CreatePayment handles POST /api/v1/payments
// @Summary Create a payment transaction
// @Tags Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string false "Idempotency key"
// @Param request body dto.CreatePaymentRequest true "Create payment request"
// @Success 201 {object} map[string]interface{} "Wrapped dto.PaymentResponse"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments [post]
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	userID := mustUserID(c)
	var req dto.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	var purposeRefID *uuid.UUID
	if req.PurposeRefID != "" {
		id, err := uuid.Parse(req.PurposeRefID)
		if err != nil {
			respondBadRequest(c, "invalid purpose_ref_id")
			return
		}
		purposeRefID = &id
	}

	pt, paymentURL, err := h.paymentUC.CreatePayment(
		c.Request.Context(), userID, req.Amount, req.Method, req.Purpose, purposeRefID, c.ClientIP(),
	)
	if err != nil {
		respondErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": &dto.PaymentResponse{
			ID:         pt.ID.String(),
			UserID:     pt.UserID.String(),
			OrderRef:   pt.OrderRef,
			Amount:     pt.Amount,
			Currency:   pt.Currency,
			Method:     pt.Method,
			Purpose:    pt.Purpose,
			Status:     pt.Status,
			PaymentURL: &paymentURL,
			ExpiresAt:  pt.ExpiresAt,
			CreatedAt:  pt.CreatedAt,
		},
		"request_id": c.GetString("request_id"),
	})
}

// GetPayment handles GET /api/v1/payments/:id
// @Summary Get payment by ID
// @Tags Payments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Payment ID (uuid)"
// @Success 200 {object} map[string]interface{} "Wrapped dto.PaymentResponse"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/{id} [get]
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid id")
		return
	}
	pt, err := h.paymentUC.GetPayment(c.Request.Context(), id)
	if err != nil {
		respondErr(c, err)
		return
	}
	respond(c, &dto.PaymentResponse{
		ID:        pt.ID.String(),
		UserID:    pt.UserID.String(),
		OrderRef:  pt.OrderRef,
		Amount:    pt.Amount,
		Currency:  pt.Currency,
		Method:    pt.Method,
		Purpose:   pt.Purpose,
		Status:    pt.Status,
		ExpiresAt: pt.ExpiresAt,
		PaidAt:    pt.PaidAt,
		CreatedAt: pt.CreatedAt,
	})
}

// GetPaymentStatus handles GET /api/v1/payments/:id/status
// @Summary Get payment status by ID
// @Tags Payments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Payment ID (uuid)"
// @Success 200 {object} map[string]interface{} "Wrapped dto.PaymentStatusResponse"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/{id}/status [get]
func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid id")
		return
	}
	status, err := h.paymentUC.GetPaymentStatus(c.Request.Context(), id)
	if err != nil {
		respondErr(c, err)
		return
	}
	respond(c, &dto.PaymentStatusResponse{ID: id.String(), Status: status})
}

// GetHistory handles GET /api/v1/payments/history
// @Summary List payment history for current user
// @Tags Payments
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Page size" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{} "Wrapped dto.PaymentHistoryResponse"
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /payments/history [get]
func (h *PaymentHandler) GetHistory(c *gin.Context) {
	userID := mustUserID(c)
	limit := queryInt(c, "limit", 20)
	offset := queryInt(c, "offset", 0)

	items, total, err := h.paymentUC.GetHistory(c.Request.Context(), userID, limit, offset)
	if err != nil {
		respondErr(c, err)
		return
	}

	dtos := make([]*dto.PaymentResponse, 0, len(items))
	for _, pt := range items {
		dtos = append(dtos, &dto.PaymentResponse{
			ID:        pt.ID.String(),
			UserID:    pt.UserID.String(),
			OrderRef:  pt.OrderRef,
			Amount:    pt.Amount,
			Currency:  pt.Currency,
			Method:    pt.Method,
			Purpose:   pt.Purpose,
			Status:    pt.Status,
			ExpiresAt: pt.ExpiresAt,
			PaidAt:    pt.PaidAt,
			CreatedAt: pt.CreatedAt,
		})
	}

	respond(c, &dto.PaymentHistoryResponse{Items: dtos, Total: total, Limit: limit, Offset: offset})
}
