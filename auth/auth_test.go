package auth_test

import (
	"os"
	"testing"
	"time"
	"github.com/dgrijalva/jwt-go"
	"context"
	"errors"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type AuthMock struct {
	mock.Mock
}
/*
ROOST_METHOD_HASH=generateToken_2cc40e0108
ROOST_METHOD_SIG_HASH=generateToken_9de4114fe8


 */
func TestgenerateToken(t *testing.T) {

	tests := []struct {
		name        string
		userID      uint
		currentTime time.Time
		jwtSecret   string
		expectError bool
	}{
		{
			name:        "Successful Token Generation",
			userID:      1,
			currentTime: time.Now(),
			jwtSecret:   "validSecret",
			expectError: false,
		},
		{
			name:        "Token Expiry Time Check",
			userID:      1,
			currentTime: time.Now(),
			jwtSecret:   "validSecret",
			expectError: false,
		},
		{
			name:        "Invalid JWT Secret",
			userID:      1,
			currentTime: time.Now(),
			jwtSecret:   "",
			expectError: true,
		},
		{
			name:        "User ID Zero",
			userID:      0,
			currentTime: time.Now(),
			jwtSecret:   "validSecret",
			expectError: false,
		},
		{
			name:        "Future Date",
			userID:      1,
			currentTime: time.Now().Add(24 * time.Hour),
			jwtSecret:   "validSecret",
			expectError: false,
		},
		{
			name:        "Past Date",
			userID:      1,
			currentTime: time.Now().Add(-24 * time.Hour),
			jwtSecret:   "validSecret",
			expectError: false,
		},
		{
			name:        "Very Large User ID",
			userID:      9999999999,
			currentTime: time.Now(),
			jwtSecret:   "validSecret",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			os.Setenv("JWT_SECRET", tt.jwtSecret)
			jwtSecret = []byte(tt.jwtSecret)

			token, err := generateToken(tt.userID, tt.currentTime)

			if (err != nil) != tt.expectError {
				t.Errorf("generateToken() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				if token == "" {
					t.Errorf("generateToken() token is empty, expected non-empty token")
				} else {

					parsedToken, err := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
						return jwtSecret, nil
					})

					if err != nil {
						t.Errorf("Error parsing token: %v", err)
					}

					if claims, ok := parsedToken.Claims.(*claims); ok && parsedToken.Valid {

						expectedExpiry := tt.currentTime.Add(time.Hour * 72).Unix()
						if claims.ExpiresAt != expectedExpiry {
							t.Errorf("Token expiry time = %v, expected %v", claims.ExpiresAt, expectedExpiry)
						}
					} else {
						t.Errorf("Invalid token claims")
					}
				}
			}
		})
	}
}

/*
ROOST_METHOD_HASH=GetUserID_f2dd680cb2
ROOST_METHOD_SIG_HASH=GetUserID_e739312e3d


 */
func (m *AuthMock) AuthFromMD(ctx context.Context, expectedScheme string) (string, error) {
	args := m.Called(ctx, expectedScheme)
	return args.String(0), args.Error(1)
}

func TestGetUserID(t *testing.T) {
	jwtSecret = []byte("test_secret")

	tests := []struct {
		name          string
		token         string
		mockAuthError error
		expectedID    uint
		expectedError error
	}{
		{
			name: "Valid Token - Successful User ID Retrieval",
			token: func() string {
				token, _ := generateToken(1, jwtSecret, time.Now().Add(1*time.Hour), time.Now())
				return token
			}(),
			expectedID:    1,
			expectedError: nil,
		},
		{
			name:          "Invalid Token Format",
			token:         "invalid.token.format",
			expectedID:    0,
			expectedError: errors.New("invalid token: it's not even a token"),
		},
		{
			name: "Token Expired",
			token: func() string {
				token, _ := generateToken(1, jwtSecret, time.Now().Add(-1*time.Hour), time.Now().Add(-2*time.Hour))
				return token
			}(),
			expectedID:    0,
			expectedError: errors.New("token expired"),
		},
		{
			name: "Token Not Valid Yet",
			token: func() string {
				token, _ := generateToken(1, jwtSecret, time.Now().Add(1*time.Hour), time.Now().Add(1*time.Hour))
				return token
			}(),
			expectedID:    0,
			expectedError: errors.New("token expired"),
		},
		{
			name: "Token with Invalid Signature",
			token: func() string {
				token, _ := generateToken(1, []byte("wrong_secret"), time.Now().Add(1*time.Hour), time.Now())
				return token
			}(),
			expectedID:    0,
			expectedError: errors.New("invalid token: couldn't handle this token; signature is invalid"),
		},
		{
			name:          "Missing Token in Metadata",
			mockAuthError: errors.New("Request unauthenticated with Token"),
			expectedID:    0,
			expectedError: errors.New("Request unauthenticated with Token"),
		},
		{
			name: "Claims Type Mismatch",
			token: func() string {
				claims := jwt.MapClaims{"user_id": "not_a_number"}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString(jwtSecret)
				return tokenString
			}(),
			expectedID:    0,
			expectedError: errors.New("invalid token: cannot map token to claims"),
		},
		{
			name: "JWT Secret Not Set",
			token: func() string {
				token, _ := generateToken(1, jwtSecret, time.Now().Add(1*time.Hour), time.Now())
				return token
			}(),
			expectedID:    0,
			expectedError: errors.New("invalid token: couldn't handle this token; key is of invalid type"),
		},
		{
			name: "Context Cancellation",
			token: func() string {
				token, _ := generateToken(1, jwtSecret, time.Now().Add(1*time.Hour), time.Now())
				return token
			}(),
			expectedID:    0,
			expectedError: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			authMock := new(AuthMock)
			if tt.mockAuthError != nil {
				authMock.On("AuthFromMD", mock.Anything, "Token").Return("", tt.mockAuthError)
			} else {
				authMock.On("AuthFromMD", mock.Anything, "Token").Return(tt.token, nil)
			}

			originalAuthFromMD := grpc_auth.AuthFromMD
			grpc_auth.AuthFromMD = authMock.AuthFromMD
			defer func() { grpc_auth.AuthFromMD = originalAuthFromMD }()

			if tt.name == "Context Cancellation" {
				cancel()
			}

			userID, err := GetUserID(ctx)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID)
			}
		})
	}
}

func generateToken(userID uint, secret []byte, expirationTime time.Time, notBefore time.Time) (string, error) {
	claims := &claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			NotBefore: notBefore.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

