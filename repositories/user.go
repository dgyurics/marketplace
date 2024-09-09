package repositories

import (
	"context"

	"github.com/dgyurics/marketplace/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
	StoreRefreshToken(ctx context.Context, refreshToken *models.RefreshToken) error
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (email, phone, password_hash)
		VALUES (NULLIF($1, ''), NULLIF($2, ''), $3)
		RETURNING id, COALESCE(email, ''), COALESCE(phone, ''), admin, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query, user.Email, user.Phone, user.PasswordHash).
		Scan(&user.ID, &user.Email, &user.Phone, &user.Admin, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(ctx, "SELECT id, COALESCE(email, ''), COALESCE(phone, ''), password_hash, admin, created_at, updated_at FROM users WHERE phone = $1", phone).Scan(&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.Admin, &user.CreatedAt, &user.UpdatedAt)
	return &user, err
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.pool.QueryRow(ctx, "SELECT id, COALESCE(email, ''), COALESCE(phone, ''), password_hash, admin, created_at, updated_at FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.Admin, &user.CreatedAt, &user.UpdatedAt)
	return &user, err
}

func (r *userRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	rows, err := r.pool.Query(ctx, "SELECT id, COALESCE(email, ''), COALESCE(phone, ''), admin, created_at, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Email, &user.Phone, &user.Admin, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *userRepository) StoreRefreshToken(ctx context.Context, refreshToken *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at, revoked, last_used)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query, refreshToken.UserID, refreshToken.TokenHash, refreshToken.ExpiresAt, refreshToken.CreatedAt, refreshToken.Revoked, refreshToken.LastUsed)
	return err
}
