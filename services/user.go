package services

import (
	"context"
	"errors"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, user *types.User) error
	CreateGuest(ctx context.Context, user *types.User) error
	ConvertGuestToUser(ctx context.Context, user *types.User) error
	GetUserByEmail(ctx context.Context, email string) (*types.User, error)
	Login(ctx context.Context, credential *types.Credential) (*types.User, error)
	GetAllUsers(ctx context.Context, page, limit int) ([]types.User, error)
}

type userService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateGuest(ctx context.Context, user *types.User) error {
	userID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}
	user.ID = userID
	return s.repo.CreateGuest(ctx, user)
}

func (s *userService) CreateUser(ctx context.Context, user *types.User) error {
	hashedPassword, err := generateFromPassword(user.Password)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)

	userID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}
	user.ID = userID
	return s.repo.CreateUser(ctx, user)
}

func (s *userService) ConvertGuestToUser(ctx context.Context, user *types.User) error {
	hashedPassword, err := generateFromPassword(user.Password)
	if err != nil {
		return err
	}
	user.ID = getUserID(ctx)
	user.PasswordHash = string(hashedPassword)
	return s.repo.ConvertGuestToUser(ctx, user)
}

// generateFromPassword generates a hashed password from a plaintext password
func generateFromPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*types.User, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

func (s *userService) Login(ctx context.Context, credentials *types.Credential) (*types.User, error) {
	return s.verifyEmail(ctx, credentials)
}

func (s *userService) verifyEmail(ctx context.Context, credentials *types.Credential) (*types.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, credentials.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password))
	return user, err
}

func (s *userService) GetAllUsers(ctx context.Context, page, limit int) ([]types.User, error) {
	return s.repo.GetAllUsers(ctx, page, limit)
}
