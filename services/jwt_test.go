package services_test

import (
	"testing"
	"time"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/assert"
)

const (
	privateKey = `-----BEGIN RSA PRIVATE KEY-----
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
	publicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAmiWxEdesR83+ntu0ENPi
w1MmAs8GFYheMXOzQG3OY4pmOvRdWLK44+aWJ1DrlMtjWCo+7DgPhF5YlmHaezEN
3MFc+5P9jYE2tQSi/3y9KyPASwM9lvNifRVqzKug3qFgv6wiTY/6iLEkzo5FZCAw
G6rJ5V+LY0vA3lBN+3hN5hL3Yv0UZt6yVUeNvh6oHwJngf32rHKysYFsZEQ+xYnG
li61URMDPLtEiyzFLMQ8StVAqB494VwLhAK6Ump/Wa04R1LeoGm+WtMfVeymxQu0
P1n+pUcLTP/HXICizcRvoms41FpjOjVIYatR/bfodjpUtTjmz+xfdw1GXR/0qXga
sQIDAQAB
-----END PUBLIC KEY-----`
	expiry = 24 * time.Hour
)

func createJWTService() services.JWTService {
	return services.NewJWTService(types.JWTConfig{
		PrivateKey: []byte(privateKey),
		PublicKey:  []byte(publicKey),
		Expiry:     expiry,
	})
}

func TestGenerateToken(t *testing.T) {
	service := createJWTService()
	user := types.User{ID: "123", Email: "user@example.com", Role: "admin"}

	token, err := service.GenerateToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestParseToken(t *testing.T) {
	service := createJWTService()
	user := types.User{ID: "123", Email: "user@example.com", Role: "admin"}

	token, _ := service.GenerateToken(user)
	parsedUser, err := service.ParseToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, parsedUser)
	assert.Equal(t, user.ID, parsedUser.ID)
	assert.Equal(t, user.Email, parsedUser.Email)
	assert.Equal(t, user.Role, parsedUser.Role)
}

func TestParseToken_InvalidSignature(t *testing.T) {
	service := createJWTService()
	token := "invalid.token.string"

	parsedUser, err := service.ParseToken(token)
	assert.Error(t, err)
	assert.Nil(t, parsedUser)
}

func TestParseToken_ExpiredToken(t *testing.T) {
	// Create a JWT service with a very short expiration time (1 second)
	shortLivedService := services.NewJWTService(types.JWTConfig{
		PrivateKey: []byte(privateKey),
		PublicKey:  []byte(publicKey),
		Expiry:     1 * time.Second,
	})

	user := types.User{ID: "123", Email: "user@example.com", Role: "admin"}

	// Generate the token
	token, err := shortLivedService.GenerateToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Wait for the token to expire
	time.Sleep(2 * time.Second)

	// Attempt to parse the expired token
	parsedUser, err := shortLivedService.ParseToken(token)

	assert.Error(t, err, "expected an error due to token expiration")
	assert.Nil(t, parsedUser, "expected nil user for expired token")
}
