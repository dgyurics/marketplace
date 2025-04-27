package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dgyurics/marketplace/types"
)

// RefreshRepository handles the storage and retrieval of refresh tokens.
type RefreshRepository interface {
	StoreToken(ctx context.Context, refreshToken types.RefreshToken) error
	GetToken(ctx context.Context, tokenHash string) (*types.RefreshToken, error)
	// GetToken(ctx context.Context, userID, tokenHash string) (*types.RefreshToken, error) // TODO replace with this
	UpdateLastUsed(ctx context.Context, tokenID string, lastUsed time.Time) error
	RevokeTokens(ctx context.Context, userID string) error
}

type refreshRepository struct {
	db *sql.DB
}

func NewRefreshRepository(db *sql.DB) RefreshRepository {
	return &refreshRepository{db: db}
}

func (r *refreshRepository) StoreToken(ctx context.Context, token types.RefreshToken) error {
	if token.User == nil || token.User.ID == "" {
		return errors.New("user.id is required")
	}
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, token.ID, token.User.ID, token.TokenHash, token.ExpiresAt)
	return err
}

func (r *refreshRepository) GetToken(ctx context.Context, tokenHash string) (*types.RefreshToken, error) {
	var refreshToken types.RefreshToken
	var user types.User

	// FIXME last used should be updated
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
			u.role,
			u.created_at,
			u.updated_at
		FROM refresh_tokens rt
		JOIN v_users u ON rt.user_id = u.id
		WHERE rt.token_hash = $1
	`

	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
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
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	refreshToken.User = &user
	return &refreshToken, nil
}

func (r *refreshRepository) RevokeTokens(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *refreshRepository) UpdateLastUsed(ctx context.Context, tokenID string, lastUsed time.Time) error {
	query := `
		UPDATE refresh_tokens
		SET last_used = $2
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, tokenID, lastUsed)
	return err
}
