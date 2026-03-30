package domain

// Kafka topic names.
const (
	TopicOrderEvents     = "modami.core.order"
	TopicUserEvents      = "modami.user.events"
	TopicPaymentEvents   = "modami.payment.events"
	TopicCreditEvents    = "modami.payment.credits"
	TopicSubEvents       = "modami.payment.subscriptions"
)

// Outbox event types.
const (
	EventCreditDeducted        = "credit.deducted"
	EventCreditAdded           = "credit.added"
	EventContactUnlocked       = "contact.unlocked"
	EventSubscriptionActivated = "subscription.activated"
	EventSubscriptionExpired   = "subscription.expired"
	EventSubscriptionCancelled = "subscription.cancelled"
	EventPaymentSuccess        = "payment.success"
	EventPaymentFailed         = "payment.failed"
)

// Outbox statuses.
const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
)

// Incoming event types from other services.
const (
	EventOrderCompleted  = "OrderCompleted"
	EventOrderCancelled  = "OrderCancelled"
	EventUserCreated     = "UserCreated"
)
