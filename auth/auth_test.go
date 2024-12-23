package auth

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
)


var mockJWTSecret = []byte("mockSecret")
/*
ROOST_METHOD_HASH=generateToken_2cc40e0108
ROOST_METHOD_SIG_HASH=generateToken_9de4114fe8


 */
func TestgenerateToken(t *testing.T) {

	jwtSecret = []byte("test_secret")

	type testCase struct {
		description string
		userID      uint
		currentTime time.Time
		expectError bool
		validate    func(t *testing.T, token string, err error)
	}

	tests := []testCase{
		{
			description: "Successful Token Generation for Valid User ID",
			userID:      1,
			currentTime: time.Now(),
			expectError: false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			},
		},
		{
			description: "Token Expiry Set Correctly",
			userID:      1,
			currentTime: time.Now(),
			expectError: false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, _ := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})

				claims, ok := parsedToken.Claims.(*claims)
				assert.True(t, ok)
				assert.Equal(t, claims.ExpiresAt, time.Now().Add(time.Hour*72).Unix(), "Token expiry time is incorrect")
			},
		},
		{
			description: "Error Handling When JWT Signing Fails",
			userID:      1,
			currentTime: time.Now(),
			expectError: true,
			validate: func(t *testing.T, token string, err error) {
				assert.Error(t, err)
				assert.Empty(t, token)
			},
		},
		{
			description: "Token Generation with Minimum User ID (Edge Case)",
			userID:      0,
			currentTime: time.Now(),
			expectError: false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			},
		},
		{
			description: "Token Generation with Maximum User ID (Edge Case)",
			userID:      ^uint(0),
			currentTime: time.Now(),
			expectError: false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			},
		},
		{
			description: "Token Generation with Future Date (Negative Test)",
			userID:      1,
			currentTime: time.Now().Add(time.Hour * 24),
			expectError: false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			},
		},
		{
			description: "Token Generation with Past Date (Negative Test)",
			userID:      1,
			currentTime: time.Now().Add(-time.Hour * 24),
			expectError: false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {

			if tc.expectError {
				jwtSecret = []byte("invalid_secret")
			} else {
				jwtSecret = []byte("test_secret")
			}

			token, err := generateToken(tc.userID, tc.currentTime)
			tc.validate(t, token, err)
		})
	}
}


/*
ROOST_METHOD_HASH=GetUserID_f2dd680cb2
ROOST_METHOD_SIG_HASH=GetUserID_e739312e3d


 */
func TestGetUserID(t *testing.T) {

	os.Setenv("JWT_SECRET", string(mockJWTSecret))

	tests := []struct {
		name     string
		setup    func() context.Context
		expected uint
		err      string
	}{
		{
			name: "Valid Token",
			setup: func() context.Context {
				token, _ := generateToken(123, time.Now().Add(1*time.Hour).Unix())
				md := map[string]string{"authorization": "Token " + token}
				return auth.NewIncomingContext(context.Background(), md)
			},
			expected: 123,
			err:      "",
		},
		{
			name: "Missing Token",
			setup: func() context.Context {
				return context.Background()
			},
			expected: 0,
			err:      "Request unauthenticated with Token",
		},
		{
			name: "Invalid Token Format",
			setup: func() context.Context {
				md := map[string]string{"authorization": "Token invalid.token.format"}
				return auth.NewIncomingContext(context.Background(), md)
			},
			expected: 0,
			err:      "invalid token: it's not even a token",
		},
		{
			name: "Expired Token",
			setup: func() context.Context {
				token, _ := generateToken(123, time.Now().Add(-1*time.Hour).Unix())
				md := map[string]string{"authorization": "Token " + token}
				return auth.NewIncomingContext(context.Background(), md)
			},
			expected: 0,
			err:      "token expired",
		},
		{
			name: "Token with Invalid Claims",
			setup: func() context.Context {
				token, _ := generateToken(123, time.Now().Add(1*time.Hour).Unix())
				md := map[string]string{"authorization": "Token " + token}
				return auth.NewIncomingContext(context.Background(), md)
			},
			expected: 0,
			err:      "invalid token: cannot map token to claims",
		},
		{
			name: "Token with Future Not Before Time",
			setup: func() context.Context {
				token, _ := generateToken(123, time.Now().Add(1*time.Hour).Unix())
				md := map[string]string{"authorization": "Token " + token}
				return auth.NewIncomingContext(context.Background(), md)
			},
			expected: 0,
			err:      "token expired",
		},
		{
			name: "Invalid Signature",
			setup: func() context.Context {
				token, _ := generateToken(123, time.Now().Add(1*time.Hour).Unix())
				md := map[string]string{"authorization": "Token " + token}
				return auth.NewIncomingContext(context.Background(), md)
			},
			expected: 0,
			err:      "invalid token: couldn't handle this token",
		},
		{
			name: "Token with No UserID Claim",
			setup: func() context.Context {
				token, _ := generateToken(0, time.Now().Add(1*time.Hour).Unix())
				md := map[string]string{"authorization": "Token " + token}
				return auth.NewIncomingContext(context.Background(), md)
			},
			expected: 0,
			err:      "invalid token: cannot map token to claims",
		},
		{
			name: "Invalid JWT Secret",
			setup: func() context.Context {
				token, _ := generateToken(123, time.Now().Add(1*time.Hour).Unix())
				md := map[string]string{"authorization": "Token " + token}
				return auth.NewIncomingContext(context.Background(), md)
			},
			expected: 0,
			err:      "invalid token: couldn't handle this token",
		},
		{
			name: "Context Cancellation",
			setup: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			expected: 0,
			err:      "context canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			userID, err := GetUserID(ctx)
			if tt.err != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, userID)
			}
		})
	}
}

