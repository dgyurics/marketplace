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
	AuthenticateUser(ctx context.Context, credential *models.Credential) (*models.User, error)
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

func (s *userService) AuthenticateUser(ctx context.Context, credentials *models.Credential) (*models.User, error) {
	if credentials.Email != "" {
		return s.verifyEmail(ctx, credentials)
	}
	if credentials.Phone != "" {
		return s.verifyPhone(ctx, credentials)
	}
	return nil, errors.New("invalid credentials: email or phone required")
}

func (s *userService) verifyEmail(ctx context.Context, credentials *models.Credential) (*models.User, error) {
	user, _ := s.repo.GetUserByEmail(ctx, credentials.Email)
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password))
	return user, err
}

func (s *userService) verifyPhone(ctx context.Context, credentials *models.Credential) (*models.User, error) {
	user, _ := s.repo.GetUserByPhone(ctx, credentials.Phone)
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password))
	return user, err
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
