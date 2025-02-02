package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dgyurics/marketplace/models"
)

type AuthRepository interface {
	StoreRefreshToken(ctx context.Context, refreshToken models.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	RevokeRefreshTokens(ctx context.Context, userID string) error
	StoreInviteCode(ctx context.Context, code string, used bool) error
	GetInviteCode(ctx context.Context, code string) (used bool, exists bool, err error)
}

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) StoreRefreshToken(ctx context.Context, refreshToken models.RefreshToken) error {
	if refreshToken.User == nil || refreshToken.User.ID == "" {
		return errors.New("user.id is required")
	}
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at, revoked, last_used)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query, refreshToken.User.ID, refreshToken.TokenHash, refreshToken.ExpiresAt, refreshToken.CreatedAt, refreshToken.Revoked, refreshToken.LastUsed)
	return err
}

func (r *authRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	var user models.User

	query := `
		SELECT
			rt.id,
			rt.token_hash,
			rt.expires_at,
			rt.revoked,
			rt.last_used,
			rt.created_at,
			rt.updated_at,
			u.id,
			u.email,
			u.password_hash,
			u.admin,
			u.created_at,
			u.updated_at
		FROM refresh_tokens rt
		JOIN users u ON rt.user_id = u.id
		WHERE rt.token_hash = $1
	`
	if err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&refreshToken.ID,
		&refreshToken.TokenHash,
		&refreshToken.ExpiresAt,
		&refreshToken.Revoked,
		&refreshToken.LastUsed,
		&refreshToken.CreatedAt,
		&refreshToken.UpdatedAt,
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Admin,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}
	refreshToken.User = &user
	return &refreshToken, nil
}

func (r *authRepository) RevokeRefreshTokens(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

// StoreInviteCode inserts or updates an invitation code.
func (r *authRepository) StoreInviteCode(ctx context.Context, code string, used bool) error {
	var usedAt interface{}
	if used {
		usedAt = time.Now()
	} else {
		usedAt = nil
	}

	query := `
		INSERT INTO invitation_codes (code, used_at)
		VALUES ($1, $2)
		ON CONFLICT (code) DO UPDATE
		SET used_at = EXCLUDED.used_at
	`
	_, err := r.db.ExecContext(ctx, query, code, usedAt)
	return err
}

// GetInviteCode retrieves an invite code from the database
// and returns whether it has been used and if it exists
func (r *authRepository) GetInviteCode(ctx context.Context, code string) (used bool, exists bool, err error) {
	query := `SELECT used_at FROM invitation_codes WHERE code = $1`
	var usedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query, code).Scan(&usedAt)
	if err == sql.ErrNoRows {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}

	return usedAt.Valid, true, nil
}
