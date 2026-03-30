package model

import (
	"time"

	"github.com/google/uuid"
)

// Package represents a subscription plan.
type Package struct {
	ID              uuid.UUID `db:"id"`
	Code            string    `db:"code"`
	Name            string    `db:"name"`
	Tier            int16     `db:"tier"`
	PriceMonthly    int64     `db:"price_monthly"`
	PriceYearly     int64     `db:"price_yearly"`
	CreditsPerMonth int       `db:"credits_per_month"`
	SearchBoost     bool      `db:"search_boost"`
	SearchPriority  bool      `db:"search_priority"`
	BadgeName       *string   `db:"badge_name"`
	PrioritySupport bool      `db:"priority_support"`
	FeaturedSlots   int       `db:"featured_slots"`
	IsActive        bool      `db:"is_active"`
	SortOrder       int       `db:"sort_order"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}
