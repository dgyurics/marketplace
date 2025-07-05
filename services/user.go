package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, user *types.User) error
	CreateGuest(ctx context.Context, user *types.User) error
	SetCredentials(ctx context.Context, credential types.Credential) (types.User, error)
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

	// TODO - check if session/request already has a guest account/user
	// if so, use that ID instead of generating a new one

	userID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}
	user.ID = userID
	return s.repo.CreateUser(ctx, user)
}

func (s *userService) SetCredentials(ctx context.Context, credentials types.Credential) (types.User, error) {
	hashedPassword, err := generateFromPassword(credentials.Password)
	if err != nil {
		return types.User{}, err
	}
	usr := getUser(ctx)
	usr.Email = credentials.Email
	usr.PasswordHash = string(hashedPassword)

	// Update/Set credentials in database
	// If the account type is guest, it will be converted to a user account
	if err := s.repo.SetCredentials(ctx, &usr); err != nil {
		return types.User{}, err
	}

	return types.User{
		ID:    usr.ID,
		Email: credentials.Email,
		Role:  usr.Role,
	}, nil
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
		return nil, types.ErrNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, types.ErrNotFound
	}
	return user, err
}

func (s *userService) GetAllUsers(ctx context.Context, page, limit int) ([]types.User, error) {
	return s.repo.GetAllUsers(ctx, page, limit)
}
