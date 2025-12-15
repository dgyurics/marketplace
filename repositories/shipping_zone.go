package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type ShippingZoneRepository interface {
	// Check shipping availability
	IsShippable(ctx context.Context, address *types.Address) (bool, error)

	// Manage shipping zones
	AddShippingZone(ctx context.Context, zone *types.ShippingZone) error
	RemoveShippingZone(ctx context.Context, zoneID string) error
	GetShippingZones(ctx context.Context) ([]types.ShippingZone, error)

	// Manage restricted zones
	AddExcludedShippingZone(ctx context.Context, zone *types.ExcludedShippingZone) error
	RemoveExcludedShippingZone(ctx context.Context, zoneID string) error
	GetExcludedShippingZones(ctx context.Context) ([]types.ExcludedShippingZone, error)
}

type shippingZone struct {
	db *sql.DB
}

func NewShippingZoneRepository(db *sql.DB) ShippingZoneRepository {
	return &shippingZone{db: db}
}

func (r *shippingZone) IsShippable(ctx context.Context, address *types.Address) (bool, error) {
	// Check exclusions first
	var excluded bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM shipping_exclusions
			WHERE country = $1 AND postal_code = $2
		)
	`, address.Country, address.PostalCode).Scan(&excluded)
	if err != nil || excluded {
		return false, err
	}

	// Check if in allowed zone
	var allowed bool
	err = r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM shipping_zones
			WHERE country = $1
				AND (state = $2 OR state = '')
				AND (postal_code = $3 OR postal_code = '')
		)
	`, address.Country, address.State, address.PostalCode).Scan(&allowed)

	return allowed, err
}

// AddShippingZone identifies a zone where shipping is allowed
func (r *shippingZone) AddShippingZone(ctx context.Context, zone *types.ShippingZone) error {
	state := ""
	if zone.State != nil {
		state = *zone.State
	}

	postalCode := ""
	if zone.PostalCode != nil {
		postalCode = *zone.PostalCode
	}

	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Throw error if postalCode already exists in exclusions
	var exists bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM shipping_exclusions
			WHERE country = $1 AND postal_code =$2
		)`, zone.Country, postalCode).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return types.ErrConstraintViolation
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO shipping_zones (id, country, state, postal_code)
		VALUES ($1, $2, $3, $4)
	`, zone.ID, zone.Country, state, postalCode)
	if isUniqueViolation(err) {
		return types.ErrUniqueConstraintViolation
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *shippingZone) RemoveShippingZone(ctx context.Context, zoneID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM shipping_zones
		WHERE id = $1
	`, zoneID)
	return err
}

func (r *shippingZone) GetShippingZones(ctx context.Context) ([]types.ShippingZone, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, country, state, postal_code
		FROM shipping_zones
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	zones := []types.ShippingZone{}
	for rows.Next() {
		var zone types.ShippingZone
		if err := rows.Scan(&zone.ID, &zone.Country, &zone.State, &zone.PostalCode); err != nil {
			return nil, err
		}
		zones = append(zones, zone)
	}
	return zones, rows.Err()
}

// AddExcludedShippingZone identifies a zone where shipping is not supported
func (r *shippingZone) AddExcludedShippingZone(ctx context.Context, zone *types.ExcludedShippingZone) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Throw error if postalCode already exists in shipping_zones
	var exists bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM shipping_zones
			WHERE country = $1 AND postal_code =$2
		)`, zone.Country, zone.PostalCode).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return types.ErrConstraintViolation
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO shipping_exclusions (id, country, postal_code)
		VALUES ($1, $2, $3)
	`, zone.ID, zone.Country, zone.PostalCode)
	if isUniqueViolation(err) {
		return types.ErrUniqueConstraintViolation
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *shippingZone) RemoveExcludedShippingZone(ctx context.Context, zoneID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM shipping_exclusions
		WHERE id = $1
	`, zoneID)
	return err
}

func (r *shippingZone) GetExcludedShippingZones(ctx context.Context) ([]types.ExcludedShippingZone, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, country, postal_code
		FROM shipping_exclusions
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	zones := []types.ExcludedShippingZone{}
	for rows.Next() {
		var zone types.ExcludedShippingZone
		if err := rows.Scan(&zone.ID, &zone.Country, &zone.PostalCode); err != nil {
			return nil, err
		}
		zones = append(zones, zone)
	}
	return zones, rows.Err()
}
