package services

import (
	"context"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/repositories"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) error
	VerifyCredentials(ctx context.Context, username, password string) error
	GetAllUsers(ctx context.Context) ([]models.User, error)
}

type userService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return err
	}
	return nil
}

func (s *userService) VerifyCredentials(ctx context.Context, username, password string) error {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
}

func (s *userService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	return s.repo.GetAllUsers(ctx)
}
