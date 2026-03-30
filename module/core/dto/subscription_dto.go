package dto

import "time"

// PackageResponse represents a subscription package.
type PackageResponse struct {
	ID              string  `json:"id"`
	Code            string  `json:"code"`
	Name            string  `json:"name"`
	Tier            int16   `json:"tier"`
	PriceMonthly    int64   `json:"price_monthly"`
	PriceYearly     int64   `json:"price_yearly"`
	CreditsPerMonth int     `json:"credits_per_month"`
	SearchBoost     bool    `json:"search_boost"`
	SearchPriority  bool    `json:"search_priority"`
	BadgeName       *string `json:"badge_name,omitempty"`
	PrioritySupport bool    `json:"priority_support"`
	FeaturedSlots   int     `json:"featured_slots"`
}

// SubscribeRequest is the request body for POST /subscriptions.
type SubscribeRequest struct {
	PackageCode   string `json:"package_code" binding:"required"`
	BillingCycle  string `json:"billing_cycle" binding:"required,oneof=monthly yearly"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

// SubscriptionResponse represents a subscription.
type SubscriptionResponse struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	PackageID        string     `json:"package_id"`
	PackageCode      string     `json:"package_code,omitempty"`
	BillingCycle     string     `json:"billing_cycle"`
	PricePaid        int64      `json:"price_paid"`
	CreditsAllocated int        `json:"credits_allocated"`
	CreditsUsed      int        `json:"credits_used"`
	Status           string     `json:"status"`
	AutoRenew        bool       `json:"auto_renew"`
	StartDate        *time.Time `json:"start_date,omitempty"`
	EndDate          *time.Time `json:"end_date,omitempty"`
	PaymentURL       string     `json:"payment_url,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// UpgradeSubscriptionRequest is the request body for POST /subscriptions/upgrade.
type UpgradeSubscriptionRequest struct {
	PackageCode   string `json:"package_code" binding:"required"`
	BillingCycle  string `json:"billing_cycle" binding:"required,oneof=monthly yearly"`
	PaymentMethod string `json:"payment_method" binding:"required"`
}

// CancelSubscriptionResponse is returned when cancelling a subscription.
type CancelSubscriptionResponse struct {
	Message string `json:"message"`
}
