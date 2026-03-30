package transformers

import (
	"github.com/modami/be-payment-service/module/core/dto"
	"github.com/modami/be-payment-service/module/core/model"
)

// ToCreditBalanceResponse converts a CreditWallet model to a balance DTO.
func ToCreditBalanceResponse(w *model.CreditWallet) *dto.CreditBalanceResponse {
	return &dto.CreditBalanceResponse{
		UserID:      w.UserID.String(),
		Balance:     w.Balance,
		TotalEarned: w.TotalEarned,
		TotalSpent:  w.TotalSpent,
	}
}

// ToCreditTransactionResponse converts a CreditTransaction model to a DTO.
func ToCreditTransactionResponse(t *model.CreditTransaction) *dto.CreditTransactionResponse {
	resp := &dto.CreditTransactionResponse{
		ID:           t.ID.String(),
		Amount:       t.Amount,
		Type:         t.Type,
		BalanceAfter: t.BalanceAfter,
		Description:  t.Description,
		CreatedAt:    t.CreatedAt,
	}
	if t.RefType != nil {
		resp.RefType = t.RefType
	}
	if t.RefID != nil {
		s := t.RefID.String()
		resp.RefID = &s
	}
	return resp
}

// ToCreditTransactionListResponse converts a slice of CreditTransaction models to a list DTO.
func ToCreditTransactionListResponse(items []*model.CreditTransaction, total, limit, offset int) *dto.CreditTransactionListResponse {
	dtos := make([]*dto.CreditTransactionResponse, 0, len(items))
	for _, t := range items {
		dtos = append(dtos, ToCreditTransactionResponse(t))
	}
	return &dto.CreditTransactionListResponse{
		Items:  dtos,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}
