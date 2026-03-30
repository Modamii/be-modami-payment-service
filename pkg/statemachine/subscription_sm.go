package statemachine

import apperrors "github.com/modami/be-payment-service/pkg/errors"

// subscriptionTransitions defines allowed state transitions for subscriptions.
var subscriptionTransitions = map[string][]string{
	"pending":   {"active", "failed"},
	"active":    {"expired", "cancelled", "active"},
	"expired":   {"active"},
	"cancelled": {"active"},
	"failed":    {"pending", "cancelled"},
}

// ValidateSubscriptionTransition returns nil if the from→to transition is allowed.
func ValidateSubscriptionTransition(from, to string) error {
	allowed, ok := subscriptionTransitions[from]
	if !ok {
		return apperrors.ErrInvalidTransition
	}
	for _, a := range allowed {
		if a == to {
			return nil
		}
	}
	return apperrors.ErrInvalidTransition
}
