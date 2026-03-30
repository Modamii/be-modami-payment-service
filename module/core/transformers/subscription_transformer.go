package transformers

import (
	"github.com/modami/be-payment-service/module/core/dto"
	"github.com/modami/be-payment-service/module/core/model"
)

// ToPackageResponse converts a Package model to a DTO.
func ToPackageResponse(p *model.Package) *dto.PackageResponse {
	return &dto.PackageResponse{
		ID:              p.ID.String(),
		Code:            p.Code,
		Name:            p.Name,
		Tier:            p.Tier,
		PriceMonthly:    p.PriceMonthly,
		PriceYearly:     p.PriceYearly,
		CreditsPerMonth: p.CreditsPerMonth,
		SearchBoost:     p.SearchBoost,
		SearchPriority:  p.SearchPriority,
		BadgeName:       p.BadgeName,
		PrioritySupport: p.PrioritySupport,
		FeaturedSlots:   p.FeaturedSlots,
	}
}

// ToSubscriptionResponse converts a Subscription model to a DTO.
func ToSubscriptionResponse(s *model.Subscription) *dto.SubscriptionResponse {
	resp := &dto.SubscriptionResponse{
		ID:               s.ID.String(),
		UserID:           s.UserID.String(),
		PackageID:        s.PackageID.String(),
		BillingCycle:     s.BillingCycle,
		PricePaid:        s.PricePaid,
		CreditsAllocated: s.CreditsAllocated,
		CreditsUsed:      s.CreditsUsed,
		Status:           s.Status,
		AutoRenew:        s.AutoRenew,
		StartDate:        s.StartDate,
		EndDate:          s.EndDate,
		CreatedAt:        s.CreatedAt,
	}
	return resp
}
