package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dgyurics/marketplace/types"
)

// RefreshRepository handles the storage and retrieval of refresh tokens.
type RefreshRepository interface {
	StoreToken(ctx context.Context, refreshToken types.RefreshToken) error
	GetToken(ctx context.Context, tokenHash string) (*types.RefreshToken, error)
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
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at, revoked, last_used)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query, token.User.ID, token.TokenHash, token.ExpiresAt, token.CreatedAt, token.Revoked, token.LastUsed)
	return err
}

// FIXME - should pass in user ID (from expired JWT) too
// otherwise it's possible for refresh token collisions to occur
func (r *refreshRepository) GetToken(ctx context.Context, tokenHash string) (*types.RefreshToken, error) {
	var refreshToken types.RefreshToken
	var user types.User

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

	rows, err := r.db.QueryContext(ctx, query, tokenHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Check if a row exists before scanning
	if !rows.Next() {
		return nil, nil
	}

	err = rows.Scan(
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
