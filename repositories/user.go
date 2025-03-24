package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dgyurics/marketplace/types"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *types.User) error
	CreateGuest(ctx context.Context, user *types.User) error
	ConvertGuestToUser(ctx context.Context, user *types.User) error
	GetUserByEmail(ctx context.Context, email string) (*types.User, error)
	GetAllUsers(ctx context.Context, page, limit int) ([]types.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) ConvertGuestToUser(ctx context.Context, user *types.User) error {
	query := `
		UPDATE users
		SET email = $1, password_hash = $2, role = 'user', updated_at = CURRENT_TIMESTAMP
		WHERE  id = $3 AND role = 'guest'
		RETURNING id, email, role, updated_at
	`
	return r.db.QueryRowContext(ctx, query, user.Email, user.PasswordHash, user.ID).
		Scan(&user.ID, &user.Email, &user.Role, &user.UpdatedAt)
}

func (r *userRepository) CreateGuest(ctx context.Context, user *types.User) error {
	query := `
		INSERT INTO users (role)
		VALUES ('guest')
		RETURNING id, role, updated_at
	`
	return r.db.QueryRowContext(ctx, query).
		Scan(&user.ID, &user.Role, &user.UpdatedAt)
}

func (r *userRepository) CreateUser(ctx context.Context, user *types.User) error {
	query := `
		INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, 'user')
		RETURNING id, email, role, updated_at
	`
	return r.db.QueryRowContext(ctx, query, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.Email, &user.Role, &user.UpdatedAt)
}

// GetUserByEmail retrieves a user from the database by email
// Returns nil, nil if no user is found
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	var user types.User
	err := r.db.QueryRowContext(ctx, "SELECT id, email, password_hash, role, updated_at FROM users WHERE email = $1", email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Return nil, nil when no user is found
		}
		return nil, err // Return error only on actual DB issues
	}
	return &user, nil
}

func (r *userRepository) GetAllUsers(ctx context.Context, page, limit int) ([]types.User, error) {
	var users []types.User
	query := `
		SELECT id, email, role, updated_at
		FROM users
		WHERE role = 'user'
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user types.User
		err = rows.Scan(&user.ID, &user.Email, &user.Role, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
