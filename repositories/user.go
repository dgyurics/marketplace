package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (email, phone, password_hash)
		VALUES (NULLIF($1, ''), NULLIF($2, ''), $3)
		RETURNING id, COALESCE(email, ''), COALESCE(phone, ''), admin, updated_at
	`
	return r.db.QueryRowContext(ctx, query, user.Email, user.Phone, user.PasswordHash).
		Scan(&user.ID, &user.Email, &user.Phone, &user.Admin, &user.UpdatedAt)
}

func (r *userRepository) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, "SELECT id, COALESCE(email, ''), COALESCE(phone, ''), password_hash, admin, updated_at FROM users WHERE phone = $1", phone).
		Scan(&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.Admin, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, "SELECT id, COALESCE(email, ''), COALESCE(phone, ''), password_hash, admin, updated_at FROM users WHERE email = $1", email).
		Scan(&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.Admin, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	rows, err := r.db.QueryContext(ctx, "SELECT id, COALESCE(email, ''), COALESCE(phone, ''), admin, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Email, &user.Phone, &user.Admin, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
