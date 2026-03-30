package domain

// Subscription statuses.
const (
	SubStatusPending   = "pending"
	SubStatusActive    = "active"
	SubStatusExpired   = "expired"
	SubStatusCancelled = "cancelled"
	SubStatusFailed    = "failed"
)

// Billing cycles.
const (
	BillingMonthly = "monthly"
	BillingYearly  = "yearly"
)

// Actor types for subscription events.
const (
	ActorUser   = "user"
	ActorSystem = "system"
	ActorAdmin  = "admin"
)
