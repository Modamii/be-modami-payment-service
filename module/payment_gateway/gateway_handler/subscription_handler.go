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
func (h *SubscriptionHandler) GetPackage(c *gin.Context) {
	pkg, err := h.subUC.GetPackageByCode(c.Request.Context(), c.Param("code"))
	if err != nil {
		respondErr(c, err)
		return
	}
	respond(c, transformers.ToPackageResponse(pkg))
}

// Subscribe handles POST /api/v1/subscriptions
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
func (h *SubscriptionHandler) Cancel(c *gin.Context) {
	userID := mustUserID(c)
	if err := h.subUC.CancelSubscription(c.Request.Context(), userID); err != nil {
		respondErr(c, err)
		return
	}
	respond(c, &dto.CancelSubscriptionResponse{Message: "Auto-renewal disabled. Subscription will expire at end of billing period."})
}

// Renew handles PUT /api/v1/subscriptions/current/renew
func (h *SubscriptionHandler) Renew(c *gin.Context) {
	userID := mustUserID(c)
	if err := h.subUC.ReenableAutoRenew(c.Request.Context(), userID); err != nil {
		respondErr(c, err)
		return
	}
	respond(c, gin.H{"message": "auto-renewal re-enabled"})
}

// Upgrade handles POST /api/v1/subscriptions/upgrade
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
