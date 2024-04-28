package repositories

import (
	"context"

	"github.com/dgyurics/marketplace/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, username, email, admin, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query, user.Username, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.Username, &user.Email, &user.Admin, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(context.Background(), "SELECT id, username, email, password_hash, admin, created_at, updated_at FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Admin, &user.CreatedAt, &user.UpdatedAt)
	return &user, err
}

func (r *userRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	rows, err := r.pool.Query(ctx, "SELECT id, username, email, admin, created_at, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Username, &user.Email, &user.Admin, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
