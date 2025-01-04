package auth

import (
	"testing"
	"time"
	"math"
	"github.com/dgrijalva/jwt-go"
	"os"
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
			t:       time.Now(),
			wantErr: true,
		},
		{
			name:    "Token Generation with Future Time",
			id:      2,
			t:       time.Now().Add(24 * time.Hour),
			wantErr: false,
		},
		{
			name:    "Token Generation with Past Time",
			id:      3,
			t:       time.Now().Add(-1 * time.Hour),
			wantErr: false,
		},
		{
			name:    "Token Generation with Maximum Uint Value",
			id:      math.MaxUint32,
			t:       time.Now().Add(time.Hour),
			wantErr: false,
		},
		{
			name:    "Token Generation with Zero Time",
			id:      4,
			t:       time.Time{},
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
				if got == "" {
					t.Errorf("GenerateTokenWithTime() returned empty token")
				}

				token, err := jwt.Parse(got, func(token *jwt.Token) (interface{}, error) {
					return jwtSecret, nil
				})

				if err != nil {
					t.Errorf("Failed to parse token: %v", err)
				}

				claims, ok := token.Claims.(jwt.MapClaims)
				if !ok {
					t.Errorf("Failed to get claims from token")
				}

				if uint(claims["id"].(float64)) != tt.id {
					t.Errorf("Token has incorrect id. got %v, want %v", uint(claims["id"].(float64)), tt.id)
				}

				expTime := time.Unix(int64(claims["exp"].(float64)), 0)
				expectedExpTime := tt.t.Add(time.Hour * 72)
				if math.Abs(float64(expTime.Sub(expectedExpTime))) > float64(time.Second) {
					t.Errorf("Token has incorrect expiration time. got %v, want %v", expTime, expectedExpTime)
				}
			}
		})
	}

	t.Run("Multiple Token Generation", func(t *testing.T) {
		id := uint(5)
		time := time.Now().Add(time.Hour)
		token1, err1 := GenerateTokenWithTime(id, time)
		token2, err2 := GenerateTokenWithTime(id, time)

		if err1 != nil || err2 != nil {
			t.Errorf("Failed to generate tokens: %v, %v", err1, err2)
		}

		if token1 == token2 {
			t.Errorf("Multiple calls with same parameters produced identical tokens")
		}
	})
}


/*
ROOST_METHOD_HASH=GenerateToken_b7f5ef3740
ROOST_METHOD_SIG_HASH=GenerateToken_d10a3e47a3


 */
func TestGenerateToken(t *testing.T) {

	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	os.Setenv("JWT_SECRET", "test_secret")

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "Successful Token Generation",
			id:      1,
			wantErr: false,
		},
		{
			name:    "Token Generation with Zero User ID",
			id:      0,
			wantErr: true,
		},
		{
			name:    "Token Generation with Maximum uint Value",
			id:      ^uint(0),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if token == "" {
					t.Errorf("GenerateToken() returned an empty token")
					return
				}

				parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					return []byte(os.Getenv("JWT_SECRET")), nil
				})

				if err != nil {
					t.Errorf("Failed to parse token: %v", err)
					return
				}

				claims, ok := parsedToken.Claims.(jwt.MapClaims)
				if !ok {
					t.Errorf("Failed to get claims from token")
					return
				}

				if uint(claims["id"].(float64)) != tt.id {
					t.Errorf("Token claim 'id' = %v, want %v", claims["id"], tt.id)
				}

				exp, ok := claims["exp"].(float64)
				if !ok {
					t.Errorf("Failed to get expiration time from token")
					return
				}

				expectedExp := time.Now().Add(time.Hour * 72).Unix()
				if int64(exp) < expectedExp-5 || int64(exp) > expectedExp+5 {
					t.Errorf("Token expiration time is not within the expected range")
				}
			}
		})
	}
}

func TestGenerateTokenConcurrent(t *testing.T) {

	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	os.Setenv("JWT_SECRET", "test_secret")

	numGoroutines := 100
	done := make(chan bool)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id uint) {
			_, err := GenerateToken(id)
			if err != nil {
				errors <- err
			}
			done <- true
		}(uint(i))
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	close(errors)
	for err := range errors {
		t.Errorf("GenerateToken() failed in goroutine: %v", err)
	}
}

func TestGenerateTokenMissingSecret(t *testing.T) {

	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	os.Unsetenv("JWT_SECRET")

	_, err := GenerateToken(1)
	if err == nil {
		t.Errorf("GenerateToken() did not return an error when JWT_SECRET is missing")
	}
}

func TestGenerateTokenPerformance(t *testing.T) {

	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	os.Setenv("JWT_SECRET", "test_secret")

	numTokens := 1000
	start := time.Now()

	for i := 0; i < numTokens; i++ {
		_, err := GenerateToken(uint(i))
		if err != nil {
			t.Errorf("GenerateToken() failed during performance test: %v", err)
		}
	}

	duration := time.Since(start)
	averageTime := duration / time.Duration(numTokens)

	if averageTime > 1*time.Millisecond {
		t.Errorf("GenerateToken() average time %v exceeds threshold", averageTime)
	}
}


/*
ROOST_METHOD_HASH=generateToken_2cc40e0108
ROOST_METHOD_SIG_HASH=generateToken_9de4114fe8


 */
func TestgenerateToken(t *testing.T) {

	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	os.Setenv("JWT_SECRET", "test_secret")

	tests := []struct {
		name        string
		userID      uint
		currentTime time.Time
		wantErr     bool
		setup       func()
		validate    func(*testing.T, string, error)
	}{
		{
			name:        "Successful Token Generation",
			userID:      1234,
			currentTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:     false,
			validate: func(t *testing.T, token string, err error) {
				assert.NotEmpty(t, token)
				assert.NoError(t, err)
			},
		},
		{
			name:        "Token Expiration Time",
			userID:      5678,
			currentTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:     false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				claims := &claims{}
				parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte("test_secret"), nil
				})
				assert.NoError(t, err)
				assert.True(t, parsedToken.Valid)
				expectedExpiration := time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC).Unix()
				assert.Equal(t, expectedExpiration, claims.ExpiresAt)
			},
		},
		{
			name:        "Token Contains Correct User ID",
			userID:      9012,
			currentTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:     false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				claims := &claims{}
				parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte("test_secret"), nil
				})
				assert.NoError(t, err)
				assert.True(t, parsedToken.Valid)
				assert.Equal(t, uint(9012), claims.UserID)
			},
		},
		{
			name:        "Error Handling - Missing JWT Secret",
			userID:      1234,
			currentTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:     true,
			setup: func() {
				os.Unsetenv("JWT_SECRET")
			},
			validate: func(t *testing.T, token string, err error) {
				assert.Error(t, err)
				assert.Empty(t, token)
			},
		},
		{
			name:        "Token Signing Method",
			userID:      1234,
			currentTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:     false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					return []byte("test_secret"), nil
				})
				assert.NoError(t, err)
				assert.True(t, parsedToken.Valid)
				assert.Equal(t, jwt.SigningMethodHS256, parsedToken.Method)
			},
		},
		{
			name:        "Generate Tokens for Multiple Users",
			userID:      1,
			currentTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:     false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				token2, err := generateToken(2, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
				assert.NoError(t, err)
				token3, err := generateToken(3, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
				assert.NoError(t, err)

				assert.NotEqual(t, token, token2)
				assert.NotEqual(t, token, token3)
				assert.NotEqual(t, token2, token3)

				validateToken := func(t *testing.T, token string, expectedUserID uint) {
					claims := &claims{}
					parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
						return []byte("test_secret"), nil
					})
					assert.NoError(t, err)
					assert.True(t, parsedToken.Valid)
					assert.Equal(t, expectedUserID, claims.UserID)
				}

				validateToken(t, token, 1)
				validateToken(t, token2, 2)
				validateToken(t, token3, 3)
			},
		},
		{
			name:        "Token Generation with Zero User ID",
			userID:      0,
			currentTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			wantErr:     false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				claims := &claims{}
				parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte("test_secret"), nil
				})
				assert.NoError(t, err)
				assert.True(t, parsedToken.Valid)
				assert.Equal(t, uint(0), claims.UserID)
			},
		},
		{
			name:        "Token Generation at Unix Epoch",
			userID:      1234,
			currentTime: time.Unix(0, 0),
			wantErr:     false,
			validate: func(t *testing.T, token string, err error) {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				claims := &claims{}
				parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte("test_secret"), nil
				})
				assert.NoError(t, err)
				assert.True(t, parsedToken.Valid)
				assert.Equal(t, int64(72*60*60), claims.ExpiresAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			} else {
				os.Setenv("JWT_SECRET", "test_secret")
			}

			token, err := generateToken(tt.userID, tt.currentTime)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.validate != nil {
				tt.validate(t, token, err)
			}
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
					UserID: 123,
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: time.Now().Add(time.Hour).Unix(),
					},
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
					UserID: 123,
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: time.Now().Add(-time.Hour).Unix(),
					},
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
					"user_id": "not_a_number",
				})
				tokenString, _ := token.SignedString(jwtSecret)
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Token "+tokenString))
			},
			expectedUserID: 0,
			expectedError:  "invalid token: cannot map token to claims",
		},
		{
			name: "Valid Token with Future Expiration",
			setupContext: func() context.Context {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
					UserID: 456,
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: time.Now().Add(time.Hour).Unix(),
					},
				})
				tokenString, _ := token.SignedString(jwtSecret)
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Token "+tokenString))
			},
			expectedUserID: 456,
			expectedError:  "",
		},
		{
			name: "Token with Invalid Signature",
			setupContext: func() context.Context {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
					UserID: 789,
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: time.Now().Add(time.Hour).Unix(),
					},
				})
				tokenString, _ := token.SignedString([]byte("wrong_secret"))
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Token "+tokenString))
			},
			expectedUserID: 0,
			expectedError:  "invalid token: couldn't handle this token",
		},
		{
			name: "Environment Variable Not Set",
			setupContext: func() context.Context {
				os.Unsetenv("JWT_SECRET")
				jwtSecret = []byte{}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
					UserID: 101,
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: time.Now().Add(time.Hour).Unix(),
					},
				})
				tokenString, _ := token.SignedString([]byte("any_secret"))
				return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Token "+tokenString))
			},
			expectedUserID: 0,
			expectedError:  "invalid token: couldn't handle this token",
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

