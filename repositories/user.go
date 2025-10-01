package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
	"github.com/lib/pq"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *types.User) error
	CreateGuest(ctx context.Context, user *types.User) error
	SetCredentials(ctx context.Context, user *types.User) error
	GetUserByEmail(ctx context.Context, email string) (*types.User, error)
	GetAllUsers(ctx context.Context, page, limit int) ([]types.User, error)
	GetAllAdmins(ctx context.Context) ([]types.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// SetCredentials used for one-time setup of new admin accounts
func (r *userRepository) SetCredentials(ctx context.Context, user *types.User) error {
	query := `
		UPDATE users
		SET
			email = $1,
			password_hash = $2,
			requires_setup = false,
			updated_at = CURRENT_TIMESTAMP
		WHERE
			id = $3 AND
			requires_setup = true
	`

	// Execute the update query
	result, err := r.db.ExecContext(ctx, query, user.Email, user.PasswordHash, user.ID)
	if isUniqueViolation(err) {
		return types.ErrUniqueConstraintViolation
	}
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return types.ErrNotFound
	}
	return nil
}

func (r *userRepository) CreateGuest(ctx context.Context, user *types.User) error {
	query := `
		INSERT INTO users (id, role)
		VALUES ($1, 'guest')
		RETURNING id, role, updated_at
	`
	return r.db.QueryRowContext(ctx, query, user.ID).
		Scan(&user.ID, &user.Role, &user.UpdatedAt)
}

func (r *userRepository) CreateUser(ctx context.Context, user *types.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, role)
		VALUES ($1, $2, $3, 'user')
		RETURNING id, email, role, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, user.ID, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.Email, &user.Role, &user.UpdatedAt)
	if isUniqueViolation(err) {
		return types.ErrUniqueConstraintViolation
	}
	if err != nil {
		return err
	}
	return nil
}

// Helper function to detect unique violations
func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // unique_violation
	}

	return false
}

// GetUserByEmail retrieves a user from the database by email
// Returns nil, nil if no user is found
// FIXME refactor to return types.ErrNotFound when no user is found
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	var user types.User
	query := `
		SELECT
			id,
			email,
			password_hash,
			role,
			COALESCE(requires_setup, false) AS requires_setup,
			updated_at
		FROM users
		WHERE email = $1
	`
	err := r.db.QueryRowContext(ctx, query, email).
		Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.RequiresSetup,
			&user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Return nil, nil when no user is found
	}
	if err != nil {
		return nil, err // Return error only on actual DB issues
	}
	return &user, nil
}

func (r *userRepository) GetAllUsers(ctx context.Context, page, limit int) ([]types.User, error) {
	users := []types.User{}
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

	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) GetAllAdmins(ctx context.Context) ([]types.User, error) {
	admins := []types.User{}
	query := `
		SELECT email
		FROM users
		WHERE role = 'admin'
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user types.User
		err = rows.Scan(&user.Email)
		if err != nil {
			return nil, err
		}
		admins = append(admins, user)
	}

	// Check for errors from iterating over rows.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return admins, nil
}
