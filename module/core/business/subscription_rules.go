package business

import (
	"time"

	"github.com/modami/be-payment-service/module/core/model"
)

// CalculateEndDate returns the subscription end date from the given start date and billing cycle.
func CalculateEndDate(start time.Time, cycle string) time.Time {
	switch cycle {
	case "yearly":
		return start.AddDate(1, 0, 0)
	default: // monthly
		return start.AddDate(0, 1, 0)
	}
}

// GetPriceForCycle returns the price for the given billing cycle.
func GetPriceForCycle(pkg *model.Package, cycle string) int64 {
	if cycle == "yearly" {
		return pkg.PriceYearly
	}
	return pkg.PriceMonthly
}

// IsSubscriptionExpired returns true if the subscription has passed its end date.
func IsSubscriptionExpired(sub *model.Subscription) bool {
	if sub.EndDate == nil {
		return false
	}
	return time.Now().After(*sub.EndDate)
}
