package auth

import (
	"testing"
	"time"
	"github.com/dgrijalva/jwt-go"
	"os"
	"fmt"
	"context"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/stretchr/testify/require"
)

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
	jwtSecret = []byte("mysecret")

	type test struct {
		name           string
		token          string
		expectedUserID uint
		expectedError  string
	}

	tests := []test{
		{
			name: "Valid Token - Successful User ID Retrieval",
			token: generateToken(jwtSecret, map[string]interface{}{
				"user_id": 123,
				"exp":     time.Now().Add(time.Hour).Unix(),
				"nbf":     time.Now().Add(-time.Minute).Unix(),
			}),
			expectedUserID: 123,
			expectedError:  "",
		},
		{
			name:           "Malformed Token - Should Return Error",
			token:          "malformed.token.here",
			expectedUserID: 0,
			expectedError:  "invalid token: it's not even a token",
		},
		{
			name: "Expired Token - Should Return Error",
			token: generateToken(jwtSecret, map[string]interface{}{
				"user_id": 123,
				"exp":     time.Now().Add(-time.Hour).Unix(),
				"nbf":     time.Now().Add(-time.Hour).Unix(),
			}),
			expectedUserID: 0,
			expectedError:  "token expired",
		},
		{
			name: "Token Not Valid Yet - Should Return Error",
			token: generateToken(jwtSecret, map[string]interface{}{
				"user_id": 123,
				"exp":     time.Now().Add(time.Hour).Unix(),
				"nbf":     time.Now().Add(time.Hour).Unix(),
			}),
			expectedUserID: 0,
			expectedError:  "token expired",
		},
		{
			name: "Invalid Signing Method - Should Return Error",
			token: generateInvalidSigningToken(map[string]interface{}{
				"user_id": 123,
				"exp":     time.Now().Add(time.Hour).Unix(),
				"nbf":     time.Now().Add(-time.Minute).Unix(),
			}),
			expectedUserID: 0,
			expectedError:  "invalid token: couldn't handle this token; signature is invalid",
		},
		{
			name:           "Missing Token in Context - Should Return Error",
			token:          "",
			expectedUserID: 0,
			expectedError:  "transport: context canceled",
		},
		{
			name: "Invalid Token Claims - Should Return Error",
			token: generateToken(jwtSecret, map[string]interface{}{
				"exp": time.Now().Add(time.Hour).Unix(),
				"nbf": time.Now().Add(-time.Minute).Unix(),
			}),
			expectedUserID: 0,
			expectedError:  "invalid token: cannot map token to claims",
		},
		{
			name: "Token with Invalid Secret - Should Return Error",
			token: generateToken([]byte("invalidsecret"), map[string]interface{}{
				"user_id": 123,
				"exp":     time.Now().Add(time.Hour).Unix(),
				"nbf":     time.Now().Add(-time.Minute).Unix(),
			}),
			expectedUserID: 0,
			expectedError:  "invalid token: couldn't handle this token; signature is invalid",
		},
		{
			name: "Token with Incorrect Type in Claims - Should Return Error",
			token: generateToken(jwtSecret, map[string]interface{}{
				"user_id": "notAnInt",
				"exp":     time.Now().Add(time.Hour).Unix(),
				"nbf":     time.Now().Add(-time.Minute).Unix(),
			}),
			expectedUserID: 0,
			expectedError:  "invalid token: cannot map token to claims",
		},
		{
			name: "Token with Valid Claims but Expired - Should Return Error",
			token: generateToken(jwtSecret, map[string]interface{}{
				"user_id": 123,
				"exp":     time.Now().Add(-time.Minute).Unix(),
				"nbf":     time.Now().Add(-time.Hour).Unix(),
			}),
			expectedUserID: 0,
			expectedError:  "token expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.token != "" {
				ctx = grpc_auth.WithAuth(ctx, "Token "+tt.token)
			}
			userID, err := GetUserID(ctx)
			if tt.expectedError == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expectedUserID, userID)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func generateInvalidSigningToken(claimsMap map[string]interface{}) string {
	claims := jwt.MapClaims{}
	for k, v := range claimsMap {
		claims[k] = v
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	tokenString, _ := token.SignedString([]byte("mysecret"))
	return tokenString
}

func generateToken(secret []byte, claimsMap map[string]interface{}) string {
	claims := jwt.MapClaims{}
	for k, v := range claimsMap {
		claims[k] = v
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(secret)
	return tokenString
}

