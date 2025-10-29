package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dgyurics/marketplace/types"
)

// PasswordRepository is the interface that defines the methods for interacting with password reset codes
type PasswordRepository interface {
	StoreResetCode(ctx context.Context, code *types.PasswordReset) error
	GetResetCode(ctx context.Context, userID string) (*types.PasswordReset, error)
	MarkResetCodeUsed(ctx context.Context, email string) error
	UpdatePassword(ctx context.Context, email, password string) error
}

type passwordRepository struct {
	db *sql.DB
}

func NewPasswordRepository(db *sql.DB) PasswordRepository {
	return &passwordRepository{db: db}
}

// StoreResetCode stores a password reset code in the database
func (r *passwordRepository) StoreResetCode(ctx context.Context, code *types.PasswordReset) error {
	if code.User == nil || code.User.ID == "" {
		return errors.New("user.id is required")
	}
	query := `
		INSERT INTO password_reset_codes (id, user_id, code_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, code.ID, code.User.ID, code.CodeHash, code.ExpiresAt)
	return err
}

func (r *passwordRepository) GetResetCode(ctx context.Context, email string) (*types.PasswordReset, error) {
	query := `
		SELECT
			prc.id,
			prc.code_hash,
			prc.expires_at,
			prc.used,
			prc.created_at,
			prc.updated_at,
			u.id,
			u.email,
			u.role,
			u.created_at,
			u.updated_at
		FROM password_reset_codes prc
		JOIN v_users u ON prc.user_id = u.id
		WHERE u.email = $1
		ORDER BY prc.created_at DESC
		LIMIT 1
	`
	var code types.PasswordReset
	var user types.User
	if err := r.db.QueryRowContext(ctx, query, email).Scan(
		&code.ID,
		&code.CodeHash,
		&code.ExpiresAt,
		&code.Used,
		&code.CreatedAt,
		&code.UpdatedAt,
		&user.ID,
		&user.Email,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}
	code.User = &user
	return &code, nil
}

// MarkResetCodeUsed updates a password reset code to mark it as used
func (r *passwordRepository) MarkResetCodeUsed(ctx context.Context, email string) error {
	query := `
		UPDATE password_reset_codes
		SET used = TRUE, updated_at = NOW()
		WHERE user_id = (
			SELECT id FROM users WHERE email = $1
		)
		AND used = FALSE
		AND expires_at > NOW()
	`
	res, err := r.db.ExecContext(ctx, query, email)
	if err != nil {
		return err
	}
	// lib/pq always returns nil error for RowsAffected()
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return types.ErrNotFound
	}
	return nil
}

// UpdatePassword updates a user's password
func (r *passwordRepository) UpdatePassword(ctx context.Context, email, password string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE email = $2
	`
	_, err := r.db.ExecContext(ctx, query, string(password), email)
	return err
}
