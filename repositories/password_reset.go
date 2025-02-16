package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dgyurics/marketplace/models"
)

type PasswordResetRepository interface {
	StorePasswordResetCode(ctx context.Context, code *models.PasswordResetCode) error
	GetPasswordResetCode(ctx context.Context, userID string) (*models.PasswordResetCode, error)
	MarkPasswordResetCodeUsed(ctx context.Context, email string) error
	UpdatePassword(ctx context.Context, email, password string) error
}

type passwordResetRepository struct {
	db *sql.DB
}

func NewPasswordResetRepository(db *sql.DB) PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

// StorePasswordResetCode stores a password reset code in the database
func (r *passwordResetRepository) StorePasswordResetCode(ctx context.Context, code *models.PasswordResetCode) error {
	if code.User == nil || code.User.ID == "" {
		return errors.New("user.id is required")
	}
	query := `
		INSERT INTO password_reset_codes (user_id, code_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query, code.User.ID, code.CodeHash, code.ExpiresAt)
	return err
}

func (r *passwordResetRepository) GetPasswordResetCode(ctx context.Context, email string) (*models.PasswordResetCode, error) {
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
			u.admin,
			u.created_at,
			u.updated_at
		FROM password_reset_codes prc
		JOIN users u ON prc.user_id = u.id
		WHERE u.email = $1
		ORDER BY prc.created_at DESC
		LIMIT 1
	`
	var code models.PasswordResetCode
	var user models.User
	if err := r.db.QueryRowContext(ctx, query, email).Scan(
		&code.ID,
		&code.CodeHash,
		&code.ExpiresAt,
		&code.Used,
		&code.CreatedAt,
		&code.UpdatedAt,
		&user.ID,
		&user.Email,
		&user.Admin,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}
	code.User = &user
	return &code, nil
}

// MarkPasswordResetCodeUsed updates a password reset code to mark it as used
func (r *passwordResetRepository) MarkPasswordResetCodeUsed(ctx context.Context, email string) error {
	query := `
		UPDATE password_reset_codes
		SET used = TRUE, updated_at = NOW()
		WHERE user_id = (
			SELECT id FROM users WHERE email = $1
		)
		AND used = FALSE
		AND expires_at > NOW()
	`
	result, err := r.db.ExecContext(ctx, query, email)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no valid password reset code found for user")
	}
	return nil
}

// UpdatePassword updates a user's password
func (r *passwordResetRepository) UpdatePassword(ctx context.Context, email, password string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE email = $2
	`
	_, err := r.db.ExecContext(ctx, query, string(password), email)
	return err
}
