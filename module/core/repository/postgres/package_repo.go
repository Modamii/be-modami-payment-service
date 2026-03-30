package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
	"github.com/modami/be-payment-service/module/core/model"
)

// PackageRepo implements repository.PackageRepository.
type PackageRepo struct {
	db *sqlx.DB
}

func NewPackageRepo(db *sqlx.DB) *PackageRepo {
	return &PackageRepo{db: db}
}

func (r *PackageRepo) List(ctx context.Context) ([]*model.Package, error) {
	var pkgs []*model.Package
	err := r.db.SelectContext(ctx, &pkgs,
		`SELECT id, code, name, tier, price_monthly, price_yearly, credits_per_month,
		        search_boost, search_priority, badge_name, priority_support,
		        featured_slots, is_active, sort_order, created_at, updated_at
		 FROM packages WHERE is_active = TRUE ORDER BY sort_order`)
	return pkgs, err
}

func (r *PackageRepo) GetByCode(ctx context.Context, code string) (*model.Package, error) {
	var pkg model.Package
	err := r.db.GetContext(ctx, &pkg,
		`SELECT id, code, name, tier, price_monthly, price_yearly, credits_per_month,
		        search_boost, search_priority, badge_name, priority_support,
		        featured_slots, is_active, sort_order, created_at, updated_at
		 FROM packages WHERE code = $1`, code)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	return &pkg, err
}

func (r *PackageRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Package, error) {
	var pkg model.Package
	err := r.db.GetContext(ctx, &pkg,
		`SELECT id, code, name, tier, price_monthly, price_yearly, credits_per_month,
		        search_boost, search_priority, badge_name, priority_support,
		        featured_slots, is_active, sort_order, created_at, updated_at
		 FROM packages WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	return &pkg, err
}

func (r *PackageRepo) Create(ctx context.Context, pkg *model.Package) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO packages (id, code, name, tier, price_monthly, price_yearly, credits_per_month,
		 search_boost, search_priority, badge_name, priority_support, featured_slots, is_active, sort_order)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		pkg.ID, pkg.Code, pkg.Name, pkg.Tier, pkg.PriceMonthly, pkg.PriceYearly, pkg.CreditsPerMonth,
		pkg.SearchBoost, pkg.SearchPriority, pkg.BadgeName, pkg.PrioritySupport,
		pkg.FeaturedSlots, pkg.IsActive, pkg.SortOrder)
	return err
}

func (r *PackageRepo) Update(ctx context.Context, pkg *model.Package) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE packages SET name=$1, tier=$2, price_monthly=$3, price_yearly=$4,
		 credits_per_month=$5, search_boost=$6, search_priority=$7, badge_name=$8,
		 priority_support=$9, featured_slots=$10, is_active=$11, sort_order=$12
		 WHERE id=$13`,
		pkg.Name, pkg.Tier, pkg.PriceMonthly, pkg.PriceYearly, pkg.CreditsPerMonth,
		pkg.SearchBoost, pkg.SearchPriority, pkg.BadgeName, pkg.PrioritySupport,
		pkg.FeaturedSlots, pkg.IsActive, pkg.SortOrder, pkg.ID)
	return err
}
