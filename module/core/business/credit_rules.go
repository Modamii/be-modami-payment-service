package business

import (
	apperrors "github.com/modami/be-payment-service/pkg/errors"
)

// ValidateDeductCredit validates whether a deduction can proceed.
func ValidateDeductCredit(balance, amount int) error {
	if amount <= 0 {
		return &apperrors.AppError{Code: "INVALID_AMOUNT", Message: "Amount must be positive", HTTP: 400}
	}
	if balance < amount {
		return apperrors.ErrInsufficientCredit
	}
	return nil
}

// ValidateCreditAmount validates that a credit amount is positive.
func ValidateCreditAmount(amount int) error {
	if amount <= 0 {
		return &apperrors.AppError{Code: "INVALID_AMOUNT", Message: "Amount must be positive", HTTP: 400}
	}
	return nil
}
