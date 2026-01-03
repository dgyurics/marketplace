package services

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const UserKey contextKey = "user"

type UserService interface {
	// CREATE
	CreateUser(ctx context.Context, user *types.User) error
	CreateGuest(ctx context.Context, user *types.User) error
	CreateRegistrationCode(ctx context.Context, userID string) (string, error)
	// UPDATE
	UpdatePassword(ctx context.Context, curPass, newPass string) (*types.User, error)
	UpdateEmail(ctx context.Context, newEmail string) (*types.User, error)
	ConfirmRegistrationCode(ctx context.Context, code string) error
	// GET
	Login(ctx context.Context, credential *types.Credential) (*types.User, error)
	GetUserByEmail(ctx context.Context, email string) (*types.User, error)
	GetAllUsers(ctx context.Context, page, limit int) ([]types.User, error)
	GetAllAdmins(ctx context.Context) ([]types.User, error)
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

func (s *userService) CreateRegistrationCode(ctx context.Context, userID string) (string, error) {
	// generate the code
	code, err := generateCode()
	if err != nil {
		return "", err
	}

	// store the registration code with an expiry of 15 minutes
	now := time.Now().UTC()
	if err := s.repo.CreateRegistrationCode(ctx, userID, code, now.Add(15*time.Minute)); err != nil {
		return "", err
	}

	return code, nil
}

// ConfirmRegistrationCode marks a user as verified if a valid registration code is provided
// Returns an error if the registration code is invalid or expired
func (s *userService) ConfirmRegistrationCode(ctx context.Context, code string) error {
	return s.repo.ConfirmRegistrationCode(ctx, code)
}

func (s *userService) UpdateEmail(ctx context.Context, newEmail string) (*types.User, error) {
	return s.repo.UpdateEmail(ctx, getUserID(ctx), newEmail)
}

func (s *userService) UpdatePassword(ctx context.Context, curPass, newPass string) (*types.User, error) {
	// get the user
	userID := getUserID(ctx)
	usr, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return usr, err
	}

	// compare old passwords
	err = bcrypt.CompareHashAndPassword([]byte(usr.PasswordHash), []byte(curPass))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, types.ErrNotFound
	}

	// hash new password
	hashedPassword, err := generateFromPassword(newPass)
	if err != nil {
		return &types.User{}, err
	}

	// update password
	return s.repo.UpdatePassword(ctx, userID, string(hashedPassword))
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

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, types.ErrNotFound
	}
	return user, err
}

func (s *userService) GetAllUsers(ctx context.Context, page, limit int) ([]types.User, error) {
	return s.repo.GetAllUsers(ctx, page, limit)
}

func (s *userService) GetAllAdmins(ctx context.Context) ([]types.User, error) {
	return s.repo.GetAllAdmins(ctx)
}

// Allowed characters for the registration code
const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const codeLength = 6

// GenerateCode creates a random 6-character alphanumeric code.
func generateCode() (string, error) {
	code := make([]byte, codeLength)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", errors.New("failed to generate code")
		}
		code[i] = charset[num.Int64()]
	}
	return string(code), nil
}
