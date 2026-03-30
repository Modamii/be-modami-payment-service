package domain

// Credit transaction types.
const (
	CreditTxPurchase         = "purchase"
	CreditTxUnlock           = "unlock"
	CreditTxRefund           = "refund"
	CreditTxReward           = "reward"
	CreditTxSubscriptionAlloc = "subscription_alloc"
	CreditTxExpire           = "expire"
	CreditTxAdminAdjust      = "admin_adjust"
)

// UnlockCost is the number of credits required to unlock a contact.
const UnlockCost = 1
