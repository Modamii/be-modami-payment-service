package gateway_handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/modami/be-payment-service/module/core/dto"
	"github.com/modami/be-payment-service/module/core/usecases"
)

// UnlockHandler handles contact unlock endpoints.
type UnlockHandler struct {
	unlockUC *usecases.UnlockUsecase
}

// NewUnlockHandler creates an UnlockHandler.
func NewUnlockHandler(unlockUC *usecases.UnlockUsecase) *UnlockHandler {
	return &UnlockHandler{unlockUC: unlockUC}
}

// Unlock handles POST /api/v1/unlocks
// @Summary Unlock a product contact (deduct credits)
// @Tags Unlocks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string false "Idempotency key"
// @Param request body dto.UnlockContactRequest true "Unlock request"
// @Success 201 {object} map[string]interface{} "Wrapped dto.UnlockContactResponse"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /unlocks [post]
func (h *UnlockHandler) Unlock(c *gin.Context) {
	buyerID := mustUserID(c)

	var req dto.UnlockContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		respondBadRequest(c, "invalid product_id")
		return
	}
	sellerID, err := uuid.Parse(req.SellerID)
	if err != nil {
		respondBadRequest(c, "invalid seller_id")
		return
	}

	unlock, err := h.unlockUC.UnlockContact(c.Request.Context(), buyerID, productID, sellerID, req.IdempotencyKey)
	if err != nil {
		respondErr(c, err)
		return
	}

	resp := &dto.UnlockContactResponse{
		ID:        unlock.ID.String(),
		BuyerID:   unlock.BuyerID.String(),
		ProductID: unlock.ProductID.String(),
		SellerID:  unlock.SellerID.String(),
		CreatedAt: unlock.CreatedAt,
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":    true,
		"data":       resp,
		"request_id": c.GetString("request_id"),
	})
}

// ListUnlocks handles GET /api/v1/unlocks
// @Summary List unlocks for current user
// @Tags Unlocks
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Page size" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{} "Wrapped dto.UnlockListResponse"
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /unlocks [get]
func (h *UnlockHandler) ListUnlocks(c *gin.Context) {
	buyerID := mustUserID(c)
	limit := queryInt(c, "limit", 20)
	offset := queryInt(c, "offset", 0)

	items, total, err := h.unlockUC.ListUnlocks(c.Request.Context(), buyerID, limit, offset)
	if err != nil {
		respondErr(c, err)
		return
	}

	dtos := make([]*dto.UnlockContactResponse, 0, len(items))
	for _, u := range items {
		dtos = append(dtos, &dto.UnlockContactResponse{
			ID:        u.ID.String(),
			BuyerID:   u.BuyerID.String(),
			ProductID: u.ProductID.String(),
			SellerID:  u.SellerID.String(),
			CreatedAt: u.CreatedAt,
		})
	}

	respond(c, &dto.UnlockListResponse{Items: dtos, Total: total, Limit: limit, Offset: offset})
}

// CheckUnlock handles GET /api/v1/unlocks/check/:product_id
// @Summary Check if product already unlocked
// @Tags Unlocks
// @Produce json
// @Security BearerAuth
// @Param product_id path string true "Product ID (uuid)"
// @Success 200 {object} map[string]interface{} "Wrapped dto.CheckUnlockResponse"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /unlocks/check/{product_id} [get]
func (h *UnlockHandler) CheckUnlock(c *gin.Context) {
	buyerID := mustUserID(c)
	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		respondBadRequest(c, "invalid product_id")
		return
	}

	isUnlocked, err := h.unlockUC.CheckUnlock(c.Request.Context(), buyerID, productID)
	if err != nil {
		respondErr(c, err)
		return
	}

	respond(c, &dto.CheckUnlockResponse{
		ProductID:  productID.String(),
		IsUnlocked: isUnlocked,
	})
}
