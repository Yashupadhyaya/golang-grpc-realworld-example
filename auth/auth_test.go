package auth

import (
	"os"
	"testing"
	"time"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"context"
	"errors"
	"fmt"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

var mockJwtSecret = []byte("mock_secret")
type claims struct {
	ID uint `json:"id"`
	jwt.StandardClaims
}
type claims struct {
	UserID    uint  `json:"user_id"`
	ExpiresAt int64 `json:"exp"`
	jwt.StandardClaims
}
/*
ROOST_METHOD_HASH=generateToken_2cc40e0108
ROOST_METHOD_SIG_HASH=generateToken_9de4114fe8


 */
func TestgenerateToken(t *testing.T) {

	os.Setenv("JWT_SECRET", string(mockJwtSecret))
	jwtSecret = mockJwtSecret

	tests := []struct {
		name      string
		id        uint
		now       time.Time
		expectErr bool
		validate  func(t *testing.T, token string, err error, now time.Time)
	}{
		{
			name: "Successful Token Generation",
			id:   12345,
			now:  time.Now(),
			validate: func(t *testing.T, token string, err error, now time.Time) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				t.Log("Token generated successfully")
			},
		},
		{
			name: "Token Generation with Expiry",
			id:   12345,
			now:  time.Now(),
			validate: func(t *testing.T, token string, err error, now time.Time) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, _ := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})

				if claims, ok := parsedToken.Claims.(*claims); ok && parsedToken.Valid {
					expectedExpiry := now.Add(time.Hour * 72).Unix()
					assert.Equal(t, expectedExpiry, claims.StandardClaims.ExpiresAt)
					t.Log("Token expiry time validated")
				} else {
					t.Error("Failed to parse token claims")
				}
			},
		},
		{
			name:      "Error When Signing Token",
			id:        12345,
			now:       time.Now(),
			expectErr: true,
			validate: func(t *testing.T, token string, err error, now time.Time) {
				assert.Error(t, err)
				assert.Empty(t, token)
				t.Log("Error during token signing as expected")
			},
		},
		{
			name:      "Invalid JWT Secret",
			id:        12345,
			now:       time.Now(),
			expectErr: true,
			validate: func(t *testing.T, token string, err error, now time.Time) {
				assert.Error(t, err)
				assert.Empty(t, token)
				t.Log("Invalid JWT secret handled correctly")
			},
		},
		{
			name: "Large User ID",
			id:   1 << 31,
			now:  time.Now(),
			validate: func(t *testing.T, token string, err error, now time.Time) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				t.Log("Token generated successfully for large user ID")
			},
		},
		{
			name: "Token Generation at Epoch Time",
			id:   12345,
			now:  time.Unix(0, 0),
			validate: func(t *testing.T, token string, err error, now time.Time) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, _ := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})

				if claims, ok := parsedToken.Claims.(*claims); ok && parsedToken.Valid {
					expectedExpiry := now.Add(time.Hour * 72).Unix()
					assert.Equal(t, expectedExpiry, claims.StandardClaims.ExpiresAt)
					t.Log("Token expiry time validated for epoch time")
				} else {
					t.Error("Failed to parse token claims")
				}
			},
		},
		{
			name: "Token Generation with Future Time",
			id:   12345,
			now:  time.Now().Add(time.Hour * 24),
			validate: func(t *testing.T, token string, err error, now time.Time) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, _ := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})

				if claims, ok := parsedToken.Claims.(*claims); ok && parsedToken.Valid {
					expectedExpiry := now.Add(time.Hour * 72).Unix()
					assert.Equal(t, expectedExpiry, claims.StandardClaims.ExpiresAt)
					t.Log("Token expiry time validated for future time")
				} else {
					t.Error("Failed to parse token claims")
				}
			},
		},
		{
			name: "Token Generation with Past Time",
			id:   12345,
			now:  time.Now().Add(-time.Hour * 24),
			validate: func(t *testing.T, token string, err error, now time.Time) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, _ := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})

				if claims, ok := parsedToken.Claims.(*claims); ok && parsedToken.Valid {
					expectedExpiry := now.Add(time.Hour * 72).Unix()
					assert.Equal(t, expectedExpiry, claims.StandardClaims.ExpiresAt)
					t.Log("Token expiry time validated for past time")
				} else {
					t.Error("Failed to parse token claims")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := generateToken(tt.id, tt.now)
			tt.validate(t, token, err, tt.now)
		})
	}
}

/*
ROOST_METHOD_HASH=GetUserID_f2dd680cb2
ROOST_METHOD_SIG_HASH=GetUserID_e739312e3d


 */
func TestGetUserID(t *testing.T) {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	type testCase struct {
		name        string
		setupCtx    func() context.Context
		expectedID  uint
		expectedErr error
	}

	tests := []testCase{
		{
			name: "Valid Token",
			setupCtx: func() context.Context {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
					UserID:    123,
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
				})
				tokenString, _ := token.SignedString(jwtSecret)
				ctx := grpc_auth.WithContext(context.Background(), tokenString)
				return ctx
			},
			expectedID:  123,
			expectedErr: nil,
		},
		{
			name: "Missing Token",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectedID:  0,
			expectedErr: errors.New("Request unauthenticated with Token"),
		},
		{
			name: "Invalid Token Format",
			setupCtx: func() context.Context {
				ctx := grpc_auth.WithContext(context.Background(), "InvalidToken")
				return ctx
			},
			expectedID:  0,
			expectedErr: errors.New("invalid token: it's not even a token"),
		},
		{
			name: "Expired Token",
			setupCtx: func() context.Context {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
					UserID:    123,
					ExpiresAt: time.Now().Add(-time.Hour).Unix(),
				})
				tokenString, _ := token.SignedString(jwtSecret)
				ctx := grpc_auth.WithContext(context.Background(), tokenString)
				return ctx
			},
			expectedID:  0,
			expectedErr: errors.New("token expired"),
		},
		{
			name: "Token with Invalid Claims",
			setupCtx: func() context.Context {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"invalidClaim": "invalid",
				})
				tokenString, _ := token.SignedString(jwtSecret)
				ctx := grpc_auth.WithContext(context.Background(), tokenString)
				return ctx
			},
			expectedID:  0,
			expectedErr: errors.New("invalid token: cannot map token to claims"),
		},
		{
			name: "Token with Future Not Before Time",
			setupCtx: func() context.Context {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
					UserID:    123,
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
					StandardClaims: jwt.StandardClaims{
						NotBefore: time.Now().Add(time.Hour).Unix(),
					},
				})
				tokenString, _ := token.SignedString(jwtSecret)
				ctx := grpc_auth.WithContext(context.Background(), tokenString)
				return ctx
			},
			expectedID:  0,
			expectedErr: errors.New("token expired"),
		},
		{
			name: "Invalid Signature",
			setupCtx: func() context.Context {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
					UserID:    123,
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
				})
				tokenString, _ := token.SignedString([]byte("wrongsecret"))
				ctx := grpc_auth.WithContext(context.Background(), tokenString)
				return ctx
			},
			expectedID:  0,
			expectedErr: errors.New("invalid token: couldn't handle this token; signature is invalid"),
		},
		{
			name: "Token with No UserID Claim",
			setupCtx: func() context.Context {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"exp": time.Now().Add(time.Hour).Unix(),
				})
				tokenString, _ := token.SignedString(jwtSecret)
				ctx := grpc_auth.WithContext(context.Background(), tokenString)
				return ctx
			},
			expectedID:  0,
			expectedErr: errors.New("invalid token: cannot map token to claims"),
		},
		{
			name: "Invalid JWT Secret",
			setupCtx: func() context.Context {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
					UserID:    123,
					ExpiresAt: time.Now().Add(time.Hour).Unix(),
				})
				tokenString, _ := token.SignedString([]byte("anotherwrongsecret"))
				ctx := grpc_auth.WithContext(context.Background(), tokenString)
				return ctx
			},
			expectedID:  0,
			expectedErr: errors.New("invalid token: couldn't handle this token; signature is invalid"),
		},
		{
			name: "Context Cancellation",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expectedID:  0,
			expectedErr: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.setupCtx()
			id, err := GetUserID(ctx)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedID, id)
			}
			t.Logf("Test %s: expected ID %d, got ID %d, expected error %v, got error %v", tc.name, tc.expectedID, id, tc.expectedErr, err)
		})
	}
}

func init() {
	os.Setenv("JWT_SECRET", "mysecretkey")
}

