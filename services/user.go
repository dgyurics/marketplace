package services

import (
	"context"
	"errors"
	"time"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/repositories"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) error
	VerifyCredentials(ctx context.Context, credential *models.User) error
	GetAllUsers(ctx context.Context) ([]models.User, error)
	StoreRefreshToken(ctx context.Context, userID string, token string, expiresAt time.Time) error
}

type userService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)
	return s.repo.CreateUser(ctx, user)
}

func (s *userService) VerifyCredentials(ctx context.Context, user *models.User) error {
	if user.Email != "" {
		return s.verifyEmail(ctx, user)
	}
	if user.Phone != "" {
		return s.verifyPhone(ctx, user)
	}
	return errors.New("invalid credentials: email or phone required")
}

func (s *userService) verifyEmail(ctx context.Context, user *models.User) error {
	dbUser, err := s.repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(user.Password))
}

func (s *userService) verifyPhone(ctx context.Context, user *models.User) error {
	dbUser, err := s.repo.GetUserByPhone(ctx, user.Phone)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(user.Password))
}

func (s *userService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	return s.repo.GetAllUsers(ctx)
}

func (s *userService) StoreRefreshToken(ctx context.Context, userID string, token string, expiresAt time.Time) error {
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	refreshToken := &models.RefreshToken{
		UserID:    userID,
		TokenHash: string(hashedToken),
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	return s.repo.StoreRefreshToken(ctx, refreshToken)
}
