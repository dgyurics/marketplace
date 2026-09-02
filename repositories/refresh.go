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
	GetToken(ctx context.Context, tokenHash string) (types.RefreshToken, error)
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

func (r *refreshRepository) GetToken(ctx context.Context, tokenHash string) (types.RefreshToken, error) {
	query := `
		UPDATE refresh_tokens rt
		SET last_used = NOW()
		FROM v_users u
		WHERE rt.user_id = u.id
			AND rt.token_hash = $1
			AND NOT rt.revoked
			AND rt.expires_at > NOW()
		RETURNING
			rt.id, rt.token_hash, rt.expires_at, rt.revoked,
			rt.last_used, rt.created_at, rt.updated_at, u.id,
			u.email, u.password_hash, u.role, u.created_at,
			u.updated_at
	`

	var refreshToken types.RefreshToken
	var user types.User
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
	if err == sql.ErrNoRows {
		return types.RefreshToken{}, types.ErrNotFound
	}
	if err != nil {
		return types.RefreshToken{}, err
	}

	refreshToken.User = &user
	return refreshToken, nil
}

func (r *refreshRepository) RevokeTokens(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
