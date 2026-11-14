package services

import (
	"context"
	"errors"
	"testing"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"golang.org/x/crypto/bcrypt"
)

type passwordUserRepository struct {
	repositories.UserRepository
	user    *types.User
	updated bool
}

func (r *passwordUserRepository) GetUserByID(ctx context.Context, userID string) (*types.User, error) {
	return r.user, nil
}

func (r *passwordUserRepository) UpdatePassword(ctx context.Context, userID, hash string) (*types.User, error) {
	r.updated = true
	r.user.PasswordHash = &hash
	return r.user, nil
}

func TestUpdatePassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("current-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	validHash := string(hash)
	invalidHash := "invalid-hash"

	for _, tt := range []struct {
		name     string
		hash     *string
		password string
		wantErr  error
	}{
		{"no password", nil, "current-password", types.ErrNotFound},
		{"wrong password", &validHash, "wrong-password", types.ErrNotFound},
		{"invalid hash", &invalidHash, "current-password", bcrypt.ErrHashTooShort},
		{"valid password", &validHash, "current-password", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &passwordUserRepository{user: &types.User{ID: "user-123", PasswordHash: tt.hash}}
			service := NewUserService(repo)
			ctx := contextWithUserID(context.Background(), repo.user.ID)

			user, err := service.UpdatePassword(ctx, tt.password, "new-password")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr != nil {
				if repo.updated || user != nil {
					t.Fatal("password verification failure must not update the user")
				}
				return
			}
			if !repo.updated || user == nil {
				t.Fatal("expected updated user")
			}
			if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte("new-password")); err != nil {
				t.Fatalf("new password does not match stored hash: %v", err)
			}
		})
	}
}
