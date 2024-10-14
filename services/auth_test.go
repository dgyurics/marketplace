package services

import (
	"context"
	"testing"
	"time"

	"github.com/dgyurics/marketplace/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	privateKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAmiWxEdesR83+ntu0ENPiw1MmAs8GFYheMXOzQG3OY4pmOvRd
WLK44+aWJ1DrlMtjWCo+7DgPhF5YlmHaezEN3MFc+5P9jYE2tQSi/3y9KyPASwM9
lvNifRVqzKug3qFgv6wiTY/6iLEkzo5FZCAwG6rJ5V+LY0vA3lBN+3hN5hL3Yv0U
Zt6yVUeNvh6oHwJngf32rHKysYFsZEQ+xYnGli61URMDPLtEiyzFLMQ8StVAqB49
4VwLhAK6Ump/Wa04R1LeoGm+WtMfVeymxQu0P1n+pUcLTP/HXICizcRvoms41Fpj
OjVIYatR/bfodjpUtTjmz+xfdw1GXR/0qXgasQIDAQABAoIBAQCG7RkaChNl4rzO
JnduB1nFKQHrkXS84lm4pZKwga0XWixzzDPtELtf2RVzopQi8QirQodDUyrZ7Y9T
SqHoFR8SLTsLhxV4iDLvrfhS88fNfAS0ZEjD2ZRK8rVCI7SzSsSZ4b1A8RcWESCr
oMLCip4xiYQhv0kOCF/w+I/Z3wsop/ON32rM6H2oRlpTtNRXGpM89wcMair6ZVRU
15SuJTlbNcvqRwe2SaXGCh97pYA97vN69ojYKgt+wa9HDxB+WmXF8cW/fVQ46mVN
ZpbCMNktYbbCpqEBIgFezFvxPC/5PR1QouDusK1IPHI0mb3IjxRlrcXnK0WHz0wW
d+TtWcQFAoGBAMncKgnuneZayWvYT9f4GnpQu2RfWcTy5u45FmPaKiQAX3hpA8Wh
pLISl1RnY8/duFQJqhDaoXs1YnaqnqNDk/j3ixMRrdFWbUrbd53NldkxK9Nw9Zug
qJek1i9BCEcKYG3ToNFMoXTgBPsOk7w6cqzO6psoiRc+4Tmdt89/uFILAoGBAMN9
iaiUfKyTazZph//hXWZcI/ZMXC3ZF6h14Bh4M69GaMrwX6qmN2dSs2VdkQf/LLmF
oMWjDS2F1AbDgyf8OF6JCKlBpn42gZGg9PqBEWd7Cc1O/VkX32E7P5FLjYy5mLB+
7F/xnQmIHOa+LWU3PM9Am1l6urKnPme3JYL1P1ezAoGBAJ90C70mwZIqWvuWtpN6
R6ghR7Wk4GuEGMlLTRV5S1p+9OtPwQwHgOqtZt7kgOK9WRMBQ1bm7TI/XFUyt/dt
tWCwYiqhB3XaWKEONjHwKROVFPKEQ284/JQ1QH+5VkmPt9Zpmppadxu0rhqHTEoe
vWEmXgpMfeZf5Fe372+4iyg7AoGAIpY0Y8IZqMLQRik3qZrq1nBY4Hu0F1yAZgqs
4kdqBYm0gqsykdOkm8AzAy0husN32z78KdtmOnaiA6xVqR5jrr4Z7TAzT8M++0/5
59QsCx3mpw9hnYCuwdokrgUq/wnbLObX1UW/He+aBW0CRRUXyidJFPS00WTrkpgB
qADR+ycCgYAiqFz+G1Rh+GSQeULE4E/248SrgID8fTWbEKra/45ulYwxb54DnbBi
s6xg4dmVfAzVSsqqZVHRL6yK2cbrm577YOK+vcpCosxhXPmqS0PGo1XbpRAGZzUX
dy+r6vZgwbokaeC2QQ9+/H89rmhJ5K3XV+5z91rvrasrQXdpIcV7QQ==
-----END RSA PRIVATE KEY-----`
	publicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAmiWxEdesR83+ntu0ENPi
w1MmAs8GFYheMXOzQG3OY4pmOvRdWLK44+aWJ1DrlMtjWCo+7DgPhF5YlmHaezEN
3MFc+5P9jYE2tQSi/3y9KyPASwM9lvNifRVqzKug3qFgv6wiTY/6iLEkzo5FZCAw
G6rJ5V+LY0vA3lBN+3hN5hL3Yv0UZt6yVUeNvh6oHwJngf32rHKysYFsZEQ+xYnG
li61URMDPLtEiyzFLMQ8StVAqB494VwLhAK6Ump/Wa04R1LeoGm+WtMfVeymxQu0
P1n+pUcLTP/HXICizcRvoms41FpjOjVIYatR/bfodjpUtTjmz+xfdw1GXR/0qXga
sQIDAQAB
-----END PUBLIC KEY-----`
	hmacSecret = "big_secret"
)

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) StoreRefreshToken(ctx context.Context, refreshToken models.RefreshToken) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}

func (m *MockAuthRepository) RevokeAllRefreshTokens(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

func TestGenerateAccessToken(t *testing.T) {
	repo := new(MockAuthRepository)
	authService := NewAuthService(repo, []byte(privateKeyPEM), []byte(publicKeyPEM), []byte(hmacSecret))

	userID := "user123"
	token, err := authService.GenerateAccessToken(userID)
	assert.NoError(t, err, "expected no error in generating access token")
	assert.NotEmpty(t, token, "expected a non-empty token string")

	// Parse and validate the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
	})
	assert.NoError(t, err, "expected no error in parsing access token")
	assert.True(t, parsedToken.Valid, "expected token to be valid")

	claims := parsedToken.Claims.(jwt.MapClaims)
	assert.Equal(t, userID, claims["user_id"], "expected user_id to match")
}

func TestValidateAccessToken(t *testing.T) {
	repo := new(MockAuthRepository)
	authService := NewAuthService(repo, []byte(privateKeyPEM), []byte(publicKeyPEM), []byte(hmacSecret))

	userID := "user123"
	token, err := authService.GenerateAccessToken(userID)
	assert.NoError(t, err, "expected no error in generating access token")

	// Validate the token
	validatedUserID, err := authService.ValidateAccessToken(token)
	assert.NoError(t, err, "expected no error in validating access token")
	assert.Equal(t, userID, validatedUserID, "expected user ID to match")
}

func TestGenerateRefreshToken(t *testing.T) {
	repo := new(MockAuthRepository)
	authService := NewAuthService(repo, []byte(privateKeyPEM), []byte(publicKeyPEM), []byte(hmacSecret))

	refreshToken, err := authService.GenerateRefreshToken()
	assert.NoError(t, err, "expected no error in generating refresh token")
	assert.NotEmpty(t, refreshToken, "expected a non-empty refresh token")
	assert.Equal(t, 64, len(refreshToken), "expected refresh token to be 64 characters long")
}

func TestValidateRefreshToken(t *testing.T) {
	repo := new(MockAuthRepository)
	authService := NewAuthService(repo, []byte(privateKeyPEM), []byte(publicKeyPEM), []byte(hmacSecret))

	// Mock the behavior of the repository
	refreshToken := "test_refresh_token"
	userID := "user123"
	expiresAt := time.Now().Add(24 * time.Hour)

	// Mock the repository to return a valid refresh token object
	repo.On("GetRefreshToken", mock.Anything, refreshToken).Return(&models.RefreshToken{
		UserID:    userID,
		TokenHash: hashRefreshToken(refreshToken, []byte(hmacSecret)),
		ExpiresAt: expiresAt,
		Revoked:   false,
		LastUsed:  time.Now(),
	}, nil)
	// Mock the repository to return no error when storing the refresh token (service will update the LastUsed field)
	repo.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("models.RefreshToken")).Return(nil)

	valid, err := authService.ValidateRefreshToken(context.Background(), refreshToken)
	assert.NoError(t, err, "expected no error in validating refresh token")
	assert.True(t, valid, "expected refresh token to be valid")
}

func TestStoreRefreshToken(t *testing.T) {
	repo := new(MockAuthRepository)
	authService := NewAuthService(repo, []byte(privateKeyPEM), []byte(publicKeyPEM), []byte(hmacSecret))

	// Mock the behavior of the repository
	token := "test_refresh_token"
	userID := "user123"

	// Expect that the StoreRefreshToken method will be called with any context and a RefreshToken struct
	repo.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("models.RefreshToken")).Return(nil)

	err := authService.StoreRefreshToken(context.Background(), userID, token)
	assert.NoError(t, err, "expected no error in storing refresh token")
}

func TestRevokeAllRefreshTokens(t *testing.T) {
	repo := new(MockAuthRepository)
	authService := NewAuthService(repo, []byte(privateKeyPEM), []byte(publicKeyPEM), []byte(hmacSecret))

	// Mock the behavior of the repository
	refreshToken := "test_refresh_token"
	repo.On("RevokeAllRefreshTokens", mock.Anything, refreshToken).Return(nil)

	err := authService.RevokeAllRefreshTokens(context.Background(), refreshToken)
	assert.NoError(t, err, "expected no error in revoking all refresh tokens")
}
