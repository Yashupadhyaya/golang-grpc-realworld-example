package auth

import (
	"testing"
	"time"
	"github.com/dgrijalva/jwt-go"
	"os"
	"fmt"
	"context"
	"errors"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/stretchr/testify/require"
)

var jwtSecret = []byte("test_secret")
type claims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}
/*
ROOST_METHOD_HASH=generateToken_2cc40e0108
ROOST_METHOD_SIG_HASH=generateToken_9de4114fe8


 */
func TestgenerateToken(t *testing.T) {

	tests := []struct {
		name        string
		userID      uint
		now         time.Time
		jwtSecret   string
		expectError bool
		validate    func(t *testing.T, token string, err error)
	}{
		{
			name:      "Successful Token Generation",
			userID:    1,
			now:       time.Now(),
			jwtSecret: "valid_jwt_secret",
			validate: func(t *testing.T, token string, err error) {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if token == "" {
					t.Fatal("expected a non-empty token")
				}
			},
		},
		{
			name:      "Token Expiry Set Correctly",
			userID:    1,
			now:       time.Now(),
			jwtSecret: "valid_jwt_secret",
			validate: func(t *testing.T, token string, err error) {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				parsedToken, err := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte("valid_jwt_secret"), nil
				})
				if err != nil {
					t.Fatalf("failed to parse token: %v", err)
				}
				if claims, ok := parsedToken.Claims.(*claims); ok && parsedToken.Valid {
					expectedExpiry := time.Now().Add(72 * time.Hour).Unix()
					if claims.ExpiresAt != expectedExpiry {
						t.Fatalf("expected expiry %v, got %v", expectedExpiry, claims.ExpiresAt)
					}
				} else {
					t.Fatal("failed to parse claims")
				}
			},
		},
		{
			name:        "Invalid JWT Secret",
			userID:      1,
			now:         time.Now(),
			jwtSecret:   "",
			expectError: true,
			validate: func(t *testing.T, token string, err error) {
				if err == nil {
					t.Fatal("expected error, got none")
				}
			},
		},
		{
			name:      "Large User ID",
			userID:    ^uint(0),
			now:       time.Now(),
			jwtSecret: "valid_jwt_secret",
			validate: func(t *testing.T, token string, err error) {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if token == "" {
					t.Fatal("expected a non-empty token")
				}
			},
		},
		{
			name:      "Zero User ID",
			userID:    0,
			now:       time.Now(),
			jwtSecret: "valid_jwt_secret",
			validate: func(t *testing.T, token string, err error) {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if token == "" {
					t.Fatal("expected a non-empty token")
				}
			},
		},
		{
			name:      "Time in the Past",
			userID:    1,
			now:       time.Now().Add(-24 * time.Hour),
			jwtSecret: "valid_jwt_secret",
			validate: func(t *testing.T, token string, err error) {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if token == "" {
					t.Fatal("expected a non-empty token")
				}
			},
		},
		{
			name:      "Time in the Future",
			userID:    1,
			now:       time.Now().Add(24 * time.Hour),
			jwtSecret: "valid_jwt_secret",
			validate: func(t *testing.T, token string, err error) {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if token == "" {
					t.Fatal("expected a non-empty token")
				}
			},
		},
		{
			name:        "Empty JWT Secret",
			userID:      1,
			now:         time.Now(),
			jwtSecret:   "",
			expectError: true,
			validate: func(t *testing.T, token string, err error) {
				if err == nil {
					t.Fatal("expected error, got none")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			originalJwtSecret := jwtSecret
			defer func() { jwtSecret = originalJwtSecret }()
			jwtSecret = []byte(tt.jwtSecret)

			token, err := generateToken(tt.userID, tt.now)

			tt.validate(t, token, err)

			if err != nil {
				t.Logf("Test %s failed with error: %v", tt.name, err)
			} else {
				t.Logf("Test %s succeeded with token: %s", tt.name, token)
			}
		})
	}
}

/*
ROOST_METHOD_HASH=GetUserID_f2dd680cb2
ROOST_METHOD_SIG_HASH=GetUserID_e739312e3d


 */
func TestGetUserID(t *testing.T) {
	type testCase struct {
		name        string
		token       string
		expectedID  uint
		expectedErr error
	}

	validToken := createToken(1, time.Now().Add(time.Hour).Unix(), time.Now().Unix(), time.Now().Add(-time.Hour).Unix(), jwtSecret)
	malformedToken := "thisisnotatoken"
	expiredToken := createToken(1, time.Now().Add(-time.Hour).Unix(), time.Now().Unix(), time.Now().Add(-2*time.Hour).Unix(), jwtSecret)
	notValidYetToken := createToken(1, time.Now().Add(time.Hour).Unix(), time.Now().Add(time.Hour).Unix(), time.Now().Unix(), jwtSecret)
	invalidSigningMethodToken := createTokenWithInvalidSigningMethod(1, time.Now().Add(time.Hour).Unix(), time.Now().Unix(), time.Now().Add(-time.Hour).Unix())
	missingToken := ""
	invalidClaimsToken := createTokenWithInvalidClaims(time.Now().Add(time.Hour).Unix(), time.Now().Unix(), time.Now().Add(-time.Hour).Unix(), jwtSecret)
	invalidSecretToken := createToken(1, time.Now().Add(time.Hour).Unix(), time.Now().Unix(), time.Now().Add(-time.Hour).Unix(), []byte("wrong_secret"))
	incorrectTypeClaimsToken := createTokenWithIncorrectTypeClaims(time.Now().Add(time.Hour).Unix(), time.Now().Unix(), time.Now().Add(-time.Hour).Unix(), jwtSecret)
	validButExpiredToken := createToken(1, time.Now().Add(-time.Hour).Unix(), time.Now().Unix(), time.Now().Add(-2*time.Hour).Unix(), jwtSecret)

	testCases := []testCase{
		{
			name:        "Valid Token - Successful User ID Retrieval",
			token:       validToken,
			expectedID:  1,
			expectedErr: nil,
		},
		{
			name:        "Malformed Token - Should Return Error",
			token:       malformedToken,
			expectedID:  0,
			expectedErr: errors.New("invalid token: it's not even a token"),
		},
		{
			name:        "Expired Token - Should Return Error",
			token:       expiredToken,
			expectedID:  0,
			expectedErr: errors.New("token expired"),
		},
		{
			name:        "Token Not Valid Yet - Should Return Error",
			token:       notValidYetToken,
			expectedID:  0,
			expectedErr: errors.New("token expired"),
		},
		{
			name:        "Invalid Signing Method - Should Return Error",
			token:       invalidSigningMethodToken,
			expectedID:  0,
			expectedErr: errors.New("invalid token: couldn't handle this token; signing method (alg) is invalid"),
		},
		{
			name:        "Missing Token in Context - Should Return Error",
			token:       missingToken,
			expectedID:  0,
			expectedErr: errors.New("Request unauthenticated with Token"),
		},
		{
			name:        "Invalid Token Claims - Should Return Error",
			token:       invalidClaimsToken,
			expectedID:  0,
			expectedErr: errors.New("invalid token: cannot map token to claims"),
		},
		{
			name:        "Token with Invalid Secret - Should Return Error",
			token:       invalidSecretToken,
			expectedID:  0,
			expectedErr: errors.New("invalid token: couldn't handle this token; signature is invalid"),
		},
		{
			name:        "Token with Incorrect Type in Claims - Should Return Error",
			token:       incorrectTypeClaimsToken,
			expectedID:  0,
			expectedErr: errors.New("invalid token: cannot map token to claims"),
		},
		{
			name:        "Token with Valid Claims but Expired - Should Return Error",
			token:       validButExpiredToken,
			expectedID:  0,
			expectedErr: errors.New("token expired"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctx := context.Background()
			if tc.token != "" {
				ctx = grpc_auth.WithToken(ctx, "Token "+tc.token)
			}

			userID, err := GetUserID(ctx)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedID, userID)
			}
		})
	}
}

func createToken(userID uint, exp, iat, nbf int64, secret []byte) string {
	claims := &claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: exp,
			IssuedAt:  iat,
			NotBefore: nbf,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(secret)
	return tokenString
}

func createTokenWithIncorrectTypeClaims(exp, iat, nbf int64, secret []byte) string {
	claims := map[string]interface{}{
		"user_id": "not_a_uint",
		"exp":     exp,
		"iat":     iat,
		"nbf":     nbf,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	tokenString, _ := token.SignedString(secret)
	return tokenString
}

func createTokenWithInvalidClaims(exp, iat, nbf int64, secret []byte) string {
	claims := &jwt.StandardClaims{
		ExpiresAt: exp,
		IssuedAt:  iat,
		NotBefore: nbf,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(secret)
	return tokenString
}

func createTokenWithInvalidSigningMethod(userID uint, exp, iat, nbf int64) string {
	claims := &claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: exp,
			IssuedAt:  iat,
			NotBefore: nbf,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)
	return tokenString
}

