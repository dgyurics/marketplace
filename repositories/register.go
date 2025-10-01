package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type RegisterRepository interface {
	EmailInUse(ctx context.Context, email string) (bool, error)
	CreatePendingUser(ctx context.Context, usr *types.PendingUser) error
	GetPendingUser(ctx context.Context, email string) (*types.PendingUser, error)
	MarkCodeUsed(ctx context.Context, id string) error
}

type registerRepository struct {
	db *sql.DB
}

func NewRegisterRepository(db *sql.DB) RegisterRepository {
	return &registerRepository{db: db}
}

func (r *registerRepository) EmailInUse(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM pending_users WHERE email = $1
			UNION
			SELECT 1 FROM users WHERE email = $1
		)
	`
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}

func (r *registerRepository) CreatePendingUser(ctx context.Context, usr *types.PendingUser) error {
	query := `
		INSERT INTO pending_users(id, email, code_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, usr.ID, usr.Email, usr.CodeHash, usr.ExpiresAt)
	return err
}

func (r *registerRepository) GetPendingUser(ctx context.Context, email string) (*types.PendingUser, error) {
	var usr types.PendingUser
	query := `
		SELECT id, email, used, code_hash, expires_at
		FROM pending_users
		WHERE email = $1
	`
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&usr.ID,
		&usr.Email,
		&usr.Used,
		&usr.CodeHash,
		&usr.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, types.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &usr, nil
}

func (r *registerRepository) MarkCodeUsed(ctx context.Context, id string) error {
	query := `
		UPDATE pending_users
		SET used = TRUE
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id) // TODO verify single row affected
	return err
}
