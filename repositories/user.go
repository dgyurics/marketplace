package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dgyurics/marketplace/types"
	"github.com/lib/pq"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *types.User) error
	CreateGuest(ctx context.Context, user *types.User) error
	ConvertGuestToUser(ctx context.Context, user *types.User) error
	GetUserByEmail(ctx context.Context, email string) (*types.User, error)
	GetAllUsers(ctx context.Context, page, limit int) ([]types.User, error)
	// TODO move these to a separate address repository
	CreateAddress(ctx context.Context, address *types.Address) error
	GetAddresses(ctx context.Context, userID string) ([]types.Address, error)
	RemoveAddress(ctx context.Context, userID, addressID string) error
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

func (r *userRepository) CreateAddress(ctx context.Context, address *types.Address) error {
	query := `
		INSERT INTO addresses (
			user_id,
			addressee,
			address_line1,
			address_line2,
			city,
			state_code,
			postal_code,
			phone
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, created_at, updated_at
	`

	// Execute the query
	err := r.db.QueryRowContext(ctx, query,
		address.UserID,
		address.Addressee,
		address.AddressLine1,
		address.AddressLine2,
		address.City,
		address.StateCode,
		address.PostalCode,
		address.Phone,
	).Scan(
		&address.ID,
		&address.UserID,
		&address.CreatedAt,
		&address.UpdatedAt,
	)
	return err
}

func (r *userRepository) GetPrimaryAddress(ctx context.Context, userID string) (*types.Address, error) {
	query := `
		SELECT id, user_id, addressee, address_line1, address_line2, city, state_code, postal_code, phone, is_deleted, created_at, updated_at
		FROM addresses
		WHERE user_id = $1 AND is_deleted = FALSE
	`
	var address types.Address
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&address.ID,
		&address.UserID,
		&address.Addressee,
		&address.AddressLine1,
		&address.AddressLine2,
		&address.City,
		&address.StateCode,
		&address.PostalCode,
		&address.Phone,
		&address.IsDeleted,
		&address.CreatedAt,
		&address.UpdatedAt,
	)
	return &address, err
}

func (r *userRepository) GetAddresses(ctx context.Context, userID string) ([]types.Address, error) {
	query := `
		SELECT id, user_id, addressee, address_line1, address_line2, city, state_code, postal_code, phone, is_deleted, created_at, updated_at
		FROM addresses
		WHERE user_id = $1 AND is_deleted = FALSE
	`

	addresses := []types.Address{}
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate over the rows and populate the addresses slice
	for rows.Next() {
		var address types.Address
		if err := rows.Scan(
			&address.ID,
			&address.UserID,
			&address.Addressee,
			&address.AddressLine1,
			&address.AddressLine2,
			&address.City,
			&address.StateCode,
			&address.PostalCode,
			&address.Phone,
			&address.IsDeleted,
			&address.CreatedAt,
			&address.UpdatedAt,
		); err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	// Check for any iteration error
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return addresses, nil
}

// RemoveAddress removes an address from the database. If the address has been associated with an order,
// a soft delete is performed instead
func (r *userRepository) RemoveAddress(ctx context.Context, userID, addressID string) error {
	queryHardDelete := `
		DELETE FROM addresses
		WHERE id = $1 AND user_id = $2
	`

	// Execute the hard delete query
	result, err := r.db.ExecContext(ctx, queryHardDelete, addressID, userID)
	if err != nil {
		// Check if the error is a foreign key constraint violation
		if isForeignKeyConstraintError(err) {
			// Perform a soft delete as a fallback
			querySoftDelete := `
				UPDATE addresses
				SET
					is_deleted = TRUE,
					updated_at = CURRENT_TIMESTAMP
				WHERE id = $1 AND user_id = $2 AND is_deleted = FALSE
			`

			result, softDeleteErr := r.db.ExecContext(ctx, querySoftDelete, addressID, userID)
			if softDeleteErr != nil {
				return softDeleteErr
			}

			// Check if the soft delete affected any rows
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if rowsAffected == 0 {
				return fmt.Errorf("address with ID %s not found or already deleted", addressID)
			}

			// Soft delete was successful
			return nil
		}
		// If the error is not a foreign key constraint error, return it
		return err
	}

	// Check if the hard delete affected any rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("address with ID %s not found", addressID)
	}

	// Hard delete was successful
	return nil
}

// isForeignKeyConstraintError checks if the error is a foreign key constraint violation in PostgreSQL.
func isForeignKeyConstraintError(err error) bool {
	pqErr, ok := err.(*pq.Error)
	if ok && pqErr.Code == "23503" { // Foreign key violation code in PostgreSQL
		return true
	}
	return false
}
