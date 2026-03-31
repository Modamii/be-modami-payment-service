package gateway_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modami/be-payment-service/module/core/dto"
	"github.com/modami/be-payment-service/module/core/transformers"
	"github.com/modami/be-payment-service/module/core/usecases"
)

// SubscriptionHandler handles subscription and package endpoints.
type SubscriptionHandler struct {
	subUC *usecases.SubscriptionUsecase
}

// NewSubscriptionHandler creates a SubscriptionHandler.
func NewSubscriptionHandler(subUC *usecases.SubscriptionUsecase) *SubscriptionHandler {
	return &SubscriptionHandler{subUC: subUC}
}

// ListPackages handles GET /api/v1/packages
// @Summary List available packages
// @Tags Packages
// @Produce json
// @Success 200 {object} map[string]interface{} "Wrapped []dto.PackageResponse"
// @Failure 500 {object} map[string]interface{}
// @Router /packages [get]
func (h *SubscriptionHandler) ListPackages(c *gin.Context) {
	pkgs, err := h.subUC.ListPackages(c.Request.Context())
	if err != nil {
		respondErr(c, err)
		return
	}
	dtos := make([]*dto.PackageResponse, 0, len(pkgs))
	for _, p := range pkgs {
		dtos = append(dtos, transformers.ToPackageResponse(p))
	}
	respond(c, dtos)
}

// GetPackage handles GET /api/v1/packages/:code
// @Summary Get package by code
// @Tags Packages
// @Produce json
// @Param code path string true "Package code"
// @Success 200 {object} map[string]interface{} "Wrapped dto.PackageResponse"
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /packages/{code} [get]
func (h *SubscriptionHandler) GetPackage(c *gin.Context) {
	pkg, err := h.subUC.GetPackageByCode(c.Request.Context(), c.Param("code"))
	if err != nil {
		respondErr(c, err)
		return
	}
	respond(c, transformers.ToPackageResponse(pkg))
}

// Subscribe handles POST /api/v1/subscriptions
// @Summary Subscribe to a package (creates pending subscription + payment URL)
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string false "Idempotency key"
// @Param request body dto.SubscribeRequest true "Subscribe request"
// @Success 201 {object} map[string]interface{} "Wrapped dto.SubscriptionResponse (with payment_url)"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /subscriptions [post]
func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	userID := mustUserID(c)
	var req dto.SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	sub, paymentURL, err := h.subUC.Subscribe(c.Request.Context(), userID, req.PackageCode, req.BillingCycle, req.PaymentMethod, c.ClientIP())
	if err != nil {
		respondErr(c, err)
		return
	}

	resp := transformers.ToSubscriptionResponse(sub)
	resp.PaymentURL = paymentURL

	c.JSON(http.StatusCreated, gin.H{
		"success":    true,
		"data":       resp,
		"request_id": c.GetString("request_id"),
	})
}

// GetCurrent handles GET /api/v1/subscriptions/current
// @Summary Get current subscription for current user
// @Tags Subscriptions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Wrapped dto.SubscriptionResponse"
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /subscriptions/current [get]
func (h *SubscriptionHandler) GetCurrent(c *gin.Context) {
	userID := mustUserID(c)
	sub, err := h.subUC.GetCurrentSubscription(c.Request.Context(), userID)
	if err != nil {
		respondErr(c, err)
		return
	}
	respond(c, transformers.ToSubscriptionResponse(sub))
}

// Cancel handles PUT /api/v1/subscriptions/current/cancel
// @Summary Disable auto-renew for current subscription
// @Tags Subscriptions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Wrapped dto.CancelSubscriptionResponse"
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /subscriptions/current/cancel [put]
func (h *SubscriptionHandler) Cancel(c *gin.Context) {
	userID := mustUserID(c)
	if err := h.subUC.CancelSubscription(c.Request.Context(), userID); err != nil {
		respondErr(c, err)
		return
	}
	respond(c, &dto.CancelSubscriptionResponse{Message: "Auto-renewal disabled. Subscription will expire at end of billing period."})
}

// Renew handles PUT /api/v1/subscriptions/current/renew
// @Summary Re-enable auto-renew for current subscription
// @Tags Subscriptions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /subscriptions/current/renew [put]
func (h *SubscriptionHandler) Renew(c *gin.Context) {
	userID := mustUserID(c)
	if err := h.subUC.ReenableAutoRenew(c.Request.Context(), userID); err != nil {
		respondErr(c, err)
		return
	}
	respond(c, gin.H{"message": "auto-renewal re-enabled"})
}

// Upgrade handles POST /api/v1/subscriptions/upgrade
// @Summary Upgrade subscription (creates new pending subscription + payment URL)
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string false "Idempotency key"
// @Param request body dto.UpgradeSubscriptionRequest true "Upgrade request"
// @Success 201 {object} map[string]interface{} "Wrapped dto.SubscriptionResponse (with payment_url)"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /subscriptions/upgrade [post]
func (h *SubscriptionHandler) Upgrade(c *gin.Context) {
	userID := mustUserID(c)
	var req dto.UpgradeSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	sub, paymentURL, err := h.subUC.UpgradeSubscription(c.Request.Context(), userID, req.PackageCode, req.BillingCycle, req.PaymentMethod, c.ClientIP())
	if err != nil {
		respondErr(c, err)
		return
	}

	resp := transformers.ToSubscriptionResponse(sub)
	resp.PaymentURL = paymentURL

	c.JSON(http.StatusCreated, gin.H{
		"success":    true,
		"data":       resp,
		"request_id": c.GetString("request_id"),
	})
}

// GetHistory handles GET /api/v1/subscriptions/history
// @Summary List subscription history for current user
// @Tags Subscriptions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Wrapped []dto.SubscriptionResponse"
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /subscriptions/history [get]
func (h *SubscriptionHandler) GetHistory(c *gin.Context) {
	userID := mustUserID(c)
	subs, err := h.subUC.GetHistory(c.Request.Context(), userID)
	if err != nil {
		respondErr(c, err)
		return
	}
	dtos := make([]*dto.SubscriptionResponse, 0, len(subs))
	for _, s := range subs {
		dtos = append(dtos, transformers.ToSubscriptionResponse(s))
	}
	respond(c, dtos)
}
