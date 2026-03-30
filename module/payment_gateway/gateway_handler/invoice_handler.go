package gateway_handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/modami/be-payment-service/module/core/dto"
	"github.com/modami/be-payment-service/module/core/model"
	"github.com/modami/be-payment-service/module/core/usecases"
)

// InvoiceHandler handles invoice endpoints.
type InvoiceHandler struct {
	invoiceUC *usecases.InvoiceUsecase
}

// NewInvoiceHandler creates an InvoiceHandler.
func NewInvoiceHandler(invoiceUC *usecases.InvoiceUsecase) *InvoiceHandler {
	return &InvoiceHandler{invoiceUC: invoiceUC}
}

// ListInvoices handles GET /api/v1/invoices
func (h *InvoiceHandler) ListInvoices(c *gin.Context) {
	userID := mustUserID(c)
	invoices, err := h.invoiceUC.ListByUser(c.Request.Context(), userID)
	if err != nil {
		respondErr(c, err)
		return
	}
	dtos := make([]*dto.InvoiceResponse, 0, len(invoices))
	for _, inv := range invoices {
		dtos = append(dtos, modelToInvoiceDTO(inv))
	}
	respond(c, dtos)
}

// GetInvoice handles GET /api/v1/invoices/:id
func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid id")
		return
	}
	inv, err := h.invoiceUC.GetInvoice(c.Request.Context(), id)
	if err != nil {
		respondErr(c, err)
		return
	}
	respond(c, modelToInvoiceDTO(inv))
}

// DownloadInvoice handles GET /api/v1/invoices/:id/download
func (h *InvoiceHandler) DownloadInvoice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid id")
		return
	}
	inv, err := h.invoiceUC.GetInvoice(c.Request.Context(), id)
	if err != nil {
		respondErr(c, err)
		return
	}
	// Return plain-text invoice (PDF generation would require a separate library).
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=invoice-%s.txt", inv.InvoiceNumber))
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK,
		"INVOICE\nNumber: %s\nDate: %s\nTotal: %d VND\n",
		inv.InvoiceNumber,
		inv.CreatedAt.Format("2006-01-02"),
		inv.Total,
	)
}

func modelToInvoiceDTO(inv *model.Invoice) *dto.InvoiceResponse {
	return &dto.InvoiceResponse{
		ID:            inv.ID.String(),
		InvoiceNumber: inv.InvoiceNumber,
		UserID:        inv.UserID.String(),
		Subtotal:      inv.Subtotal,
		TaxAmount:     inv.TaxAmount,
		Total:         inv.Total,
		Description:   inv.Description,
		Status:        inv.Status,
		BillingName:   inv.BillingName,
		BillingEmail:  inv.BillingEmail,
		CreatedAt:     inv.CreatedAt,
	}
}
