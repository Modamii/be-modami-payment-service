package gateway_handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/modami/be-payment-service/module/core/dto"
	"github.com/modami/be-payment-service/module/core/transformers"
	"github.com/modami/be-payment-service/module/core/usecases"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
)

// CreditHandler handles credit-related HTTP endpoints.
type CreditHandler struct {
	creditUC *usecases.CreditUsecase
}

// NewCreditHandler creates a CreditHandler.
func NewCreditHandler(creditUC *usecases.CreditUsecase) *CreditHandler {
	return &CreditHandler{creditUC: creditUC}
}

// GetBalance handles GET /api/v1/credits/balance
// @Summary Get current user's credit balance
// @Tags Credits
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Wrapped dto.CreditBalanceResponse"
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /credits/balance [get]
func (h *CreditHandler) GetBalance(c *gin.Context) {
	userID := mustUserID(c)
	wallet, err := h.creditUC.GetBalance(c.Request.Context(), userID)
	if err != nil {
		respondErr(c, err)
		return
	}
	respond(c, transformers.ToCreditBalanceResponse(wallet))
}

// GetTransactions handles GET /api/v1/credits/transactions
// @Summary List credit transactions for current user
// @Tags Credits
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Page size" default(20)
// @Param offset query int false "Offset" default(0)
// @Param type query string false "Filter by transaction type"
// @Success 200 {object} map[string]interface{} "Wrapped dto.CreditTransactionListResponse"
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /credits/transactions [get]
func (h *CreditHandler) GetTransactions(c *gin.Context) {
	userID := mustUserID(c)
	limit := queryInt(c, "limit", 20)
	offset := queryInt(c, "offset", 0)

	var txType *string
	if t := c.Query("type"); t != "" {
		txType = &t
	}

	items, total, err := h.creditUC.GetTransactionHistory(c.Request.Context(), userID, limit, offset, txType)
	if err != nil {
		respondErr(c, err)
		return
	}
	respond(c, transformers.ToCreditTransactionListResponse(items, total, limit, offset))
}

// Purchase handles POST /api/v1/credits/purchase
// @Summary Purchase credits (deprecated)
// @Description This endpoint is not implemented; use subscriptions flow instead.
// @Tags Credits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.PurchaseCreditRequest true "Purchase credit request"
// @Success 501 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /credits/purchase [post]
func (h *CreditHandler) Purchase(c *gin.Context) {
	var req dto.PurchaseCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	// Delegate to subscription/payment use case — returns a payment URL.
	c.JSON(http.StatusNotImplemented, gin.H{"success": false, "error": "use POST /subscriptions for package-based purchases"})
}

// AdminAdjust handles POST /api/v1/admin/credits/adjust
// @Summary Admin adjust user credits
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AdminCreditAdjustRequest true "Admin credit adjust request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/credits/adjust [post]
func (h *CreditHandler) AdminAdjust(c *gin.Context) {
	var req dto.AdminCreditAdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		respondBadRequest(c, "invalid user_id")
		return
	}

	if err := h.creditUC.AdminAdjust(c.Request.Context(), userID, req.Amount, req.Description); err != nil {
		respondErr(c, err)
		return
	}
	respond(c, gin.H{"message": "credit adjusted"})
}

// --- helpers ---

func mustUserID(c *gin.Context) uuid.UUID {
	raw, _ := c.Get("user_id")
	id, _ := uuid.Parse(raw.(string))
	return id
}

func respond(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       data,
		"request_id": c.GetString("request_id"),
	})
}

func respondErr(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.JSON(appErr.HTTP, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			},
			"request_id": c.GetString("request_id"),
		})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		},
		"request_id": c.GetString("request_id"),
	})
}

func respondBadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "VALIDATION_ERROR",
			"message": msg,
		},
		"request_id": c.GetString("request_id"),
	})
}

func queryInt(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return def
}
