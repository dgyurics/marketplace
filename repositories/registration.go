package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/dgyurics/marketplace/types"
)

type RegistrationRepository interface {
	CreateCode(ctx context.Context, userID, code string, expires time.Time) error
	VerifyCode(ctx context.Context, code string) (*types.User, error)
}

type registrationRepository struct {
	db *sql.DB
}

func NewRegistrationRepository(db *sql.DB) RegistrationRepository {
	return &registrationRepository{db: db}
}

func (r *registrationRepository) CreateCode(ctx context.Context, userID, code string, expires time.Time) error {
	query := `
		INSERT INTO registration_codes(user_id, code, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query, userID, code, expires)
	return err
}

func (r *registrationRepository) VerifyCode(ctx context.Context, code string) (*types.User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Set verified true if registration code is valid
	var usr types.User
	query := `
		UPDATE users
		SET verified = true, updated_at = NOW()
		WHERE id = (
			SELECT user_id
			FROM registration_codes
			WHERE code = $1 AND expires_at > NOW()
		)
		RETURNING id, email, role
	`
	err = tx.QueryRowContext(ctx, query, code).Scan(&usr.ID, &usr.Email, &usr.Role)
	if err == sql.ErrNoRows {
		return nil, types.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Mark the registration code as used
	query = `
		DELETE FROM registration_codes
		WHERE code = $1
	`
	if _, err := tx.ExecContext(ctx, query, code); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &usr, nil
}
