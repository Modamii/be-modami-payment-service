package dto

import "time"

// UnlockContactRequest is the request body for POST /unlocks.
type UnlockContactRequest struct {
	ProductID      string `json:"product_id" binding:"required"`
	SellerID       string `json:"seller_id" binding:"required"`
	IdempotencyKey string `json:"idempotency_key"`
}

// UnlockContactResponse is returned after a successful unlock.
type UnlockContactResponse struct {
	ID        string    `json:"id"`
	BuyerID   string    `json:"buyer_id"`
	ProductID string    `json:"product_id"`
	SellerID  string    `json:"seller_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CheckUnlockResponse is returned for GET /unlocks/check/:product_id.
type CheckUnlockResponse struct {
	ProductID  string `json:"product_id"`
	IsUnlocked bool   `json:"is_unlocked"`
}

// UnlockListResponse is returned for GET /unlocks.
type UnlockListResponse struct {
	Items  []*UnlockContactResponse `json:"items"`
	Total  int                      `json:"total"`
	Limit  int                      `json:"limit"`
	Offset int                      `json:"offset"`
}
