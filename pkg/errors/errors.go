package errors

import "net/http"

// AppError represents a structured application error.
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	HTTP    int    `json:"-"`
}

func (e *AppError) Error() string { return e.Message }

var (
	ErrInsufficientCredit   = &AppError{Code: "INSUFFICIENT_CREDIT", Message: "Insufficient credits", HTTP: http.StatusBadRequest}
	ErrAlreadyUnlocked      = &AppError{Code: "ALREADY_UNLOCKED", Message: "Already unlocked", HTTP: http.StatusConflict}
	ErrCannotUnlockOwn      = &AppError{Code: "CANNOT_UNLOCK_OWN", Message: "Cannot unlock own product", HTTP: http.StatusBadRequest}
	ErrSubscriptionActive   = &AppError{Code: "SUBSCRIPTION_ACTIVE", Message: "Subscription already active", HTTP: http.StatusConflict}
	ErrSubscriptionNotFound = &AppError{Code: "SUBSCRIPTION_NOT_FOUND", Message: "No active subscription", HTTP: http.StatusNotFound}
	ErrInvalidTransition    = &AppError{Code: "INVALID_TRANSITION", Message: "Invalid state transition", HTTP: http.StatusBadRequest}
	ErrPaymentExpired       = &AppError{Code: "PAYMENT_EXPIRED", Message: "Payment expired", HTTP: 410}
	ErrPaymentFailed        = &AppError{Code: "PAYMENT_FAILED", Message: "Payment failed", HTTP: http.StatusPaymentRequired}
	ErrInvalidSignature     = &AppError{Code: "INVALID_SIGNATURE", Message: "Invalid webhook signature", HTTP: http.StatusUnauthorized}
	ErrDuplicateRequest     = &AppError{Code: "DUPLICATE_REQUEST", Message: "Duplicate request", HTTP: http.StatusConflict}
	ErrGatewayUnavailable   = &AppError{Code: "GATEWAY_UNAVAILABLE", Message: "Gateway unavailable", HTTP: http.StatusServiceUnavailable}
	ErrRefundNotEligible    = &AppError{Code: "REFUND_NOT_ELIGIBLE", Message: "Not eligible for refund", HTTP: http.StatusBadRequest}
	ErrRateLimited          = &AppError{Code: "RATE_LIMITED", Message: "Rate limited", HTTP: http.StatusTooManyRequests}
	ErrWalletNotFound       = &AppError{Code: "WALLET_NOT_FOUND", Message: "Wallet not found", HTTP: http.StatusNotFound}
	ErrUnsupportedGateway   = &AppError{Code: "UNSUPPORTED_GATEWAY", Message: "Unsupported gateway", HTTP: http.StatusBadRequest}
	ErrNotFound             = &AppError{Code: "NOT_FOUND", Message: "Resource not found", HTTP: http.StatusNotFound}
	ErrInternal             = &AppError{Code: "INTERNAL_ERROR", Message: "Internal server error", HTTP: http.StatusInternalServerError}
	ErrUnauthorized         = &AppError{Code: "UNAUTHORIZED", Message: "Unauthorized", HTTP: http.StatusUnauthorized}
	ErrForbidden            = &AppError{Code: "FORBIDDEN", Message: "Forbidden", HTTP: http.StatusForbidden}
)
