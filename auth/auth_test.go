package auth

import (
	"os"
	"testing"
	"time"
	"github.com/dgrijalva/jwt-go"
	"math"
	"github.com/stretchr/testify/assert"
	"context"
	"google.golang.org/grpc/metadata"
)

type T struct {
	common
	isEnvSet bool
	context  *testContext
}
/*
ROOST_METHOD_HASH=GenerateTokenWithTime_d0df64aa69
ROOST_METHOD_SIG_HASH=GenerateTokenWithTime_72dd09cde6


 */
func TestGenerateTokenWithTime(t *testing.T) {

	os.Setenv("JWT_SECRET", "test_secret")
	defer os.Unsetenv("JWT_SECRET")

	tests := []struct {
		name    string
		id      uint
		t       time.Time
		wantErr bool
	}{
		{
			name:    "Successful Token Generation",
			id:      1,
			t:       time.Now().Add(time.Hour),
			wantErr: false,
		},
		{
			name:    "Token Generation with Zero User ID",
			id:      0,
			t:       time.Now().Add(time.Hour),
			wantErr: true,
		},
		{
			name:    "Token Generation with Future Time",
			id:      1,
			t:       time.Now().Add(24 * time.Hour),
			wantErr: false,
		},
		{
			name:    "Token Generation with Past Time",
			id:      1,
			t:       time.Now().Add(-1 * time.Hour),
			wantErr: true,
		},
		{
			name:    "Token Generation with Maximum Uint Value",
			id:      ^uint(0),
			t:       time.Now().Add(time.Hour),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateTokenWithTime(tt.id, tt.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateTokenWithTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {

				token, err := jwt.Parse(got, func(token *jwt.Token) (interface{}, error) {
					return []byte(os.Getenv("JWT_SECRET")), nil
				})
				if err != nil {
					t.Errorf("Failed to parse token: %v", err)
					return
				}
				claims, ok := token.Claims.(jwt.MapClaims)
				if !ok || !token.Valid {
					t.Errorf("Invalid token claims")
					return
				}
				if uint(claims["id"].(float64)) != tt.id {
					t.Errorf("Token id = %v, want %v", claims["id"], tt.id)
				}
				if int64(claims["exp"].(float64)) != tt.t.Add(time.Hour*72).Unix() {
					t.Errorf("Token exp = %v, want %v", int64(claims["exp"].(float64)), tt.t.Add(time.Hour*72).Unix())
				}
			}
		})
	}
}

func TestGenerateTokenWithTimeConcurrent(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret")
	defer os.Unsetenv("JWT_SECRET")

	concurrentCalls := 100
	errChan := make(chan error, concurrentCalls)

	for i := 0; i < concurrentCalls; i++ {
		go func(id uint) {
			_, err := GenerateTokenWithTime(id, time.Now().Add(time.Hour))
			errChan <- err
		}(uint(i))
	}

	for i := 0; i < concurrentCalls; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent call %d failed: %v", i, err)
		}
	}
}

func TestGenerateTokenWithTimeInvalidSecret(t *testing.T) {

	os.Setenv("JWT_SECRET", "")
	defer os.Unsetenv("JWT_SECRET")

	_, err := GenerateTokenWithTime(1, time.Now().Add(time.Hour))
	if err == nil {
		t.Errorf("GenerateTokenWithTime() expected error with invalid secret, got nil")
	}
}


/*
ROOST_METHOD_HASH=GenerateToken_b7f5ef3740
ROOST_METHOD_SIG_HASH=GenerateToken_d10a3e47a3


 */
func TestGenerateToken(t *testing.T) {

	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "test_secret")
	defer func() {
		os.Setenv("JWT_SECRET", originalSecret)
	}()

	tests := []struct {
		name    string
		id      uint
		wantErr bool
		setup   func()
		verify  func(*testing.T, string)
	}{
		{
			name:    "Successful Token Generation",
			id:      1,
			wantErr: false,
			verify: func(t *testing.T, token string) {
				if token == "" {
					t.Error("Expected non-empty token, got empty string")
				}
				verifyTokenClaims(t, token, 1)
			},
		},
		{
			name:    "Token Generation with Zero ID",
			id:      0,
			wantErr: false,
			verify: func(t *testing.T, token string) {
				if token == "" {
					t.Error("Expected non-empty token, got empty string")
				}
				verifyTokenClaims(t, token, 0)
			},
		},
		{
			name:    "Token Generation with Maximum uint Value",
			id:      math.MaxUint32,
			wantErr: false,
			verify: func(t *testing.T, token string) {
				if token == "" {
					t.Error("Expected non-empty token, got empty string")
				}
				verifyTokenClaims(t, token, math.MaxUint32)
			},
		},
		{
			name:    "Verification of Generated Token Structure",
			id:      42,
			wantErr: false,
			verify: func(t *testing.T, token string) {
				verifyTokenClaims(t, token, 42)
			},
		},
		{
			name:    "Token Generation with Empty JWT Secret",
			id:      1,
			wantErr: true,
			setup: func() {
				os.Unsetenv("JWT_SECRET")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
				defer os.Setenv("JWT_SECRET", "test_secret")
			}

			got, err := GenerateToken(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.verify != nil {
				tt.verify(t, got)
			}
		})
	}
}

func TestGenerateTokenConcurrent(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret")
	defer os.Unsetenv("JWT_SECRET")

	const numGoroutines = 100
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id uint) {
			_, err := GenerateToken(id)
			errChan <- err
		}(uint(i))
	}

	for i := 0; i < numGoroutines; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Goroutine %d failed: %v", i, err)
		}
	}
}

func TestGenerateTokenPerformance(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret")
	defer os.Unsetenv("JWT_SECRET")

	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, err := GenerateToken(uint(i))
		if err != nil {
			t.Errorf("GenerateToken() failed on iteration %d: %v", i, err)
		}
	}
	duration := time.Since(start)
	t.Logf("Time taken for 1000 token generations: %v", duration)
	if duration > 5*time.Second {
		t.Errorf("Token generation took too long: %v", duration)
	}
}

func verifyTokenClaims(t *testing.T, token string, expectedID uint) {
	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		t.Errorf("Failed to parse token: %v", err)
		return
	}
	if !parsedToken.Valid {
		t.Error("Token is not valid")
		return
	}
	if claims["id"] != float64(expectedID) {
		t.Errorf("Expected id claim to be %v, got %v", expectedID, claims["id"])
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Error("Expiration claim not found or not a number")
		return
	}
	expTime := time.Unix(int64(exp), 0)
	if expTime.Before(time.Now()) {
		t.Error("Token has already expired")
	}
	if expTime.Sub(time.Now()) > 73*time.Hour {
		t.Error("Token expiration time is too far in the future")
	}
}


/*
ROOST_METHOD_HASH=generateToken_2cc40e0108
ROOST_METHOD_SIG_HASH=generateToken_9de4114fe8


 */
func TestgenerateToken(t *testing.T) {
	originalJWTSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalJWTSecret)

	tests := []struct {
		name        string
		userID      uint
		currentTime time.Time
		setupEnv    func()
		wantErr     bool
		validate    func(*testing.T, string, error)
	}{
		{
			name:        "Successfully Generate Token for Valid User ID",
			userID:      1234,
			currentTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "test_secret")
			},
			wantErr: false,
			validate: func(t *testing.T, token string, err error) {
				assert.NotEmpty(t, token)
				assert.NoError(t, err)
			},
		},
		{
			name:        "Token Expiration Time Set Correctly",
			userID:      5678,
			currentTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "test_secret")
			},
			wantErr: false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				claims := &claims{}
				_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte("test_secret"), nil
				})
				assert.NoError(t, err)
				expectedExpiration := time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC).Unix()
				assert.Equal(t, expectedExpiration, claims.ExpiresAt)
			},
		},
		{
			name:        "Error Handling with Empty JWT Secret",
			userID:      9012,
			currentTime: time.Now(),
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "")
			},
			wantErr: true,
			validate: func(t *testing.T, token string, err error) {
				assert.Error(t, err)
				assert.Empty(t, token)
			},
		},
		{
			name:        "Consistent Token Generation for Same Inputs",
			userID:      3456,
			currentTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "test_secret")
			},
			wantErr: false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				token2, err := generateToken(3456, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
				assert.NoError(t, err)
				assert.Equal(t, token, token2)
			},
		},
		{
			name:        "Token Generation with Maximum Uint Value",
			userID:      ^uint(0),
			currentTime: time.Now(),
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "test_secret")
			},
			wantErr: false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			},
		},
		{
			name:        "Token Claims Contain Correct User ID",
			userID:      7890,
			currentTime: time.Now(),
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "test_secret")
			},
			wantErr: false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				claims := &claims{}
				_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte("test_secret"), nil
				})
				assert.NoError(t, err)
				assert.Equal(t, uint(7890), claims.UserID)
			},
		},
		{
			name:        "Performance Test for Token Generation",
			userID:      1111,
			currentTime: time.Now(),
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "test_secret")
			},
			wantErr: false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				start := time.Now()
				for i := 0; i < 10000; i++ {
					_, err := generateToken(uint(i), time.Now())
					assert.NoError(t, err)
				}
				duration := time.Since(start)
				assert.Less(t, duration, 5*time.Second)
			},
		},
		{
			name:        "Token Generation with Zero User ID",
			userID:      0,
			currentTime: time.Now(),
			setupEnv: func() {
				os.Setenv("JWT_SECRET", "test_secret")
			},
			wantErr: false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				claims := &claims{}
				_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte("test_secret"), nil
				})
				assert.NoError(t, err)
				assert.Equal(t, uint(0), claims.UserID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			token, err := generateToken(tt.userID, tt.currentTime)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			tt.validate(t, token, err)
		})
	}
}


/*
ROOST_METHOD_HASH=GetUserID_f2dd680cb2
ROOST_METHOD_SIG_HASH=GetUserID_e739312e3d


 */
func TestGetUserID(t *testing.T) {

	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	os.Setenv("JWT_SECRET", "test_secret")
	jwtSecret = []byte("test_secret")

	tests := []struct {
		name           string
		setupContext   func() context.Context
		expectedUserID uint
		expectedError  string
	}{
		{
			name: "Valid Token with Correct User ID",
			setupContext: func() context.Context {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: time.Now().Add(time.Hour).Unix(),
					},
					UserID: 123,
				})
				tokenString, _ := token.SignedString(jwtSecret)
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Token "+tokenString))
			},
			expectedUserID: 123,
			expectedError:  "",
		},
		{
			name: "Expired Token",
			setupContext: func() context.Context {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: time.Now().Add(-time.Hour).Unix(),
					},
					UserID: 123,
				})
				tokenString, _ := token.SignedString(jwtSecret)
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Token "+tokenString))
			},
			expectedUserID: 0,
			expectedError:  "token expired",
		},
		{
			name: "Malformed Token",
			setupContext: func() context.Context {
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Token malformed.token.here"))
			},
			expectedUserID: 0,
			expectedError:  "invalid token: it's not even a token",
		},
		{
			name: "Missing Token in Context",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedUserID: 0,
			expectedError:  "Request unauthenticated with Token",
		},
		{
			name: "Token with Invalid Claims",
			setupContext: func() context.Context {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"exp": time.Now().Add(time.Hour).Unix(),
					"uid": "not_a_uint",
				})
				tokenString, _ := token.SignedString(jwtSecret)
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Token "+tokenString))
			},
			expectedUserID: 0,
			expectedError:  "invalid token: cannot map token to claims",
		},
		{
			name: "Valid Token but JWT_SECRET Environment Variable Not Set",
			setupContext: func() context.Context {
				os.Setenv("JWT_SECRET", "")
				jwtSecret = []byte("")
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: time.Now().Add(time.Hour).Unix(),
					},
					UserID: 123,
				})
				tokenString, _ := token.SignedString([]byte("test_secret"))
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Token "+tokenString))
			},
			expectedUserID: 0,
			expectedError:  "invalid token: couldn't handle this token",
		},
		{
			name: "Token with Future Not Before Claim",
			setupContext: func() context.Context {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: time.Now().Add(time.Hour).Unix(),
						NotBefore: time.Now().Add(time.Hour).Unix(),
					},
					UserID: 123,
				})
				tokenString, _ := token.SignedString(jwtSecret)
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Token "+tokenString))
			},
			expectedUserID: 0,
			expectedError:  "token expired",
		},
		{
			name: "Valid Token with Maximum Allowed User ID",
			setupContext: func() context.Context {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: time.Now().Add(time.Hour).Unix(),
					},
					UserID: ^uint(0),
				})
				tokenString, _ := token.SignedString(jwtSecret)
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Token "+tokenString))
			},
			expectedUserID: ^uint(0),
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			userID, err := GetUserID(ctx)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedUserID, userID)
		})
	}
}

