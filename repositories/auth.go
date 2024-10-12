package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/dgyurics/marketplace/models"
)

type AuthRepository interface {
	StoreRefreshToken(ctx context.Context, refreshToken *models.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	RevokeAllRefreshTokens(ctx context.Context, tokenHash string) error
}

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) StoreRefreshToken(ctx context.Context, refreshToken *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at, revoked, last_used)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query, refreshToken.UserID, refreshToken.TokenHash, refreshToken.ExpiresAt, refreshToken.CreatedAt, refreshToken.Revoked, refreshToken.LastUsed)
	return err
}

func (r *authRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.db.QueryRowContext(ctx, "SELECT id, user_id, token_hash, expires_at, created_at, revoked, last_used FROM refresh_tokens WHERE token_hash = $1", tokenHash).
		Scan(&refreshToken.ID, &refreshToken.UserID, &refreshToken.TokenHash, &refreshToken.ExpiresAt, &refreshToken.CreatedAt, &refreshToken.Revoked, &refreshToken.LastUsed)
	return &refreshToken, err
}

func (r *authRepository) RevokeAllRefreshTokens(ctx context.Context, tokenHash string) error {
	// Begin a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback() // Roll back the transaction in case of an error

	// Fetch the refresh token
	var refreshToken models.RefreshToken
	query := `SELECT id, user_id, token_hash, expires_at, created_at, revoked, last_used
	          FROM refresh_tokens WHERE token_hash = $1`
	if err = tx.QueryRowContext(ctx, query, tokenHash).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.TokenHash,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
		&refreshToken.Revoked,
		&refreshToken.LastUsed,
	); err != nil {
		return err
	}

	// Check if the refresh token is valid (not revoked and not expired)
	if refreshToken.Revoked || refreshToken.ExpiresAt.Before(time.Now()) {
		return sql.ErrNoRows // Invalid or expired token
	}

	// Revoke all refresh tokens for the user
	revokeQuery := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`
	if _, err = tx.ExecContext(ctx, revokeQuery, refreshToken.UserID); err != nil {
		return err
	}

	return tx.Commit()
}
