package domain

// Payment statuses.
const (
	PaymentStatusPending    = "pending"
	PaymentStatusProcessing = "processing"
	PaymentStatusSuccess    = "success"
	PaymentStatusFailed     = "failed"
	PaymentStatusExpired    = "expired"
	PaymentStatusCancelled  = "cancelled"
)

// Payment methods.
const (
	PaymentMethodVNPay        = "vnpay"
	PaymentMethodMoMo         = "momo"
	PaymentMethodZaloPay      = "zalopay"
	PaymentMethodBankTransfer = "bank_transfer"
	PaymentMethodCreditCard   = "credit_card"
)

// Payment purposes.
const (
	PurposeCreditPurchase  = "credit_purchase"
	PurposeSubscription    = "subscription"
)

// Refund statuses.
const (
	RefundStatusPending    = "pending"
	RefundStatusApproved   = "approved"
	RefundStatusProcessing = "processing"
	RefundStatusCompleted  = "completed"
	RefundStatusRejected   = "rejected"
)

// Refund types.
const (
	RefundTypeGateway = "gateway"
	RefundTypeCredit  = "credit"
	RefundTypeManual  = "manual"
)

// Invoice statuses.
const (
	InvoiceStatusDraft  = "draft"
	InvoiceStatusIssued = "issued"
	InvoiceStatusPaid   = "paid"
	InvoiceStatusVoided = "voided"
)

// Payment expiry duration in minutes.
const PaymentExpiryMinutes = 15
