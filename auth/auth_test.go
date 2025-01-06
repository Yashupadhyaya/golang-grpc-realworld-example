package auth

import (
	"testing"
	"time"
	"os"
	"github.com/dgrijalva/jwt-go"
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)





type B struct {
	common
	importPath       string
	context          *benchContext
	N                int
	previousN        int
	previousDuration time.Duration
	benchFunc        func(b *B)
	benchTime        durationOrCountFlag
	bytes            int64
	missingBytes     bool
	timerOn          bool
	showAllocResult  bool
	result           BenchmarkResult
	parallelism      int

	startAllocs uint64
	startBytes  uint64

	netAllocs uint64
	netBytes  uint64

	extra map[string]float64
}
type T struct {
	common
	isEnvSet bool
	context  *testContext
}


/*
ROOST_METHOD_HASH=GenerateTokenWithTime_d0df64aa69
ROOST_METHOD_SIG_HASH=GenerateTokenWithTime_72dd09cde6


 */
func BenchmarkGenerateTokenWithTime(b *testing.B) {

	os.Setenv("JWT_SECRET", "test_secret")
	defer os.Unsetenv("JWT_SECRET")

	id := uint(6)
	ti := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GenerateTokenWithTime(id, ti)
		if err != nil {
			b.Fatalf("GenerateTokenWithTime() failed: %v", err)
		}
	}
}

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
			t:       time.Now(),
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
			t:       time.Now().Add(-24 * time.Hour),
			wantErr: false,
		},
		{
			name:    "Token Generation with Maximum Uint Value",
			id:      ^uint(0),
			t:       time.Now(),
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

				if len(got) < 20 || len(got) > 500 {
					t.Errorf("GenerateTokenWithTime() returned token with unexpected length: %d", len(got))
				}
			}
		})
	}
}

func TestGenerateTokenWithTimeConsistency(t *testing.T) {

	os.Setenv("JWT_SECRET", "test_secret")
	defer os.Unsetenv("JWT_SECRET")

	id := uint(5)
	ti := time.Now()

	token1, err1 := GenerateTokenWithTime(id, ti)
	if err1 != nil {
		t.Fatalf("First call to GenerateTokenWithTime() failed: %v", err1)
	}

	token2, err2 := GenerateTokenWithTime(id, ti)
	if err2 != nil {
		t.Fatalf("Second call to GenerateTokenWithTime() failed: %v", err2)
	}

	if token1 != token2 {
		t.Errorf("Tokens are not consistent. Token1: %s, Token2: %s", token1, token2)
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
	}{
		{
			name:    "Successful Token Generation",
			id:      1,
			wantErr: false,
		},
		{
			name:    "Token Generation with Zero ID",
			id:      0,
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
					t.Error("GenerateToken() returned an empty token")
					return
				}

				claims := jwt.MapClaims{}
				parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte(os.Getenv("JWT_SECRET")), nil
				})

				if err != nil {
					t.Errorf("Failed to parse generated token: %v", err)
					return
				}

				if !parsedToken.Valid {
					t.Error("Generated token is not valid")
					return
				}

				if id, ok := claims["id"].(float64); !ok || uint(id) != tt.id {
					t.Errorf("Token claim 'id' = %v, want %v", id, tt.id)
				}

				if exp, ok := claims["exp"].(float64); !ok {
					t.Error("Token does not contain 'exp' claim")
				} else {
					expTime := time.Unix(int64(exp), 0)
					now := time.Now()
					if expTime.Before(now) {
						t.Error("Token has already expired")
					}
					if expTime.Sub(now) > 24*time.Hour+time.Minute {
						t.Error("Token expiration time is too far in the future")
					}
				}
			}
		})
	}
}

func TestGenerateTokenConcurrent(t *testing.T) {
	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "test_secret")
	defer func() {
		os.Setenv("JWT_SECRET", originalSecret)
	}()

	const numGoroutines = 100
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id uint) {
			token, err := GenerateToken(id)
			if err != nil {
				errChan <- fmt.Errorf("Concurrent GenerateToken() failed: %v", err)
				return
			}
			if token == "" {
				errChan <- fmt.Errorf("Concurrent GenerateToken() returned empty token")
				return
			}
			errChan <- nil
		}(uint(i))
	}

	for i := 0; i < numGoroutines; i++ {
		if err := <-errChan; err != nil {
			t.Error(err)
		}
	}
}

func TestGenerateTokenInvalidSecret(t *testing.T) {

	originalSecret := os.Getenv("JWT_SECRET")

	os.Unsetenv("JWT_SECRET")

	defer func() {
		os.Setenv("JWT_SECRET", originalSecret)
	}()

	_, err := GenerateToken(1)
	if err == nil {
		t.Error("GenerateToken() did not return an error with invalid JWT secret")
	}
}

func TestGenerateTokenPerformance(t *testing.T) {
	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "test_secret")
	defer func() {
		os.Setenv("JWT_SECRET", originalSecret)
	}()

	const numTokens = 1000
	start := time.Now()

	for i := 0; i < numTokens; i++ {
		_, err := GenerateToken(uint(i))
		if err != nil {
			t.Fatalf("GenerateToken() failed: %v", err)
		}
	}

	duration := time.Since(start)
	avgTime := duration / time.Duration(numTokens)

	t.Logf("Average token generation time: %v", avgTime)
	if avgTime > time.Millisecond {
		t.Errorf("Average token generation time %v exceeds 1ms", avgTime)
	}
}

func validateToken(t *testing.T, token string, expectedID uint) {
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

	if id, ok := claims["id"].(float64); !ok || uint(id) != expectedID {
		t.Errorf("Token claim 'id' = %v, want %v", id, expectedID)
	}

	if exp, ok := claims["exp"].(float64); !ok {
		t.Error("Token does not contain 'exp' claim")
	} else {
		expTime := time.Unix(int64(exp), 0)
		now := time.Now()
		if expTime.Before(now) {
			t.Error("Token has already expired")
		}
		if expTime.Sub(now) > 24*time.Hour+time.Minute {
			t.Error("Token expiration time is too far in the future")
		}
	}
}


/*
ROOST_METHOD_HASH=generateToken_2cc40e0108
ROOST_METHOD_SIG_HASH=generateToken_9de4114fe8


 */
func TestgenerateToken(t *testing.T) {

	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	tests := []struct {
		name        string
		userID      uint
		currentTime time.Time
		setupEnv    func()
		wantErr     bool
	}{
		{
			name:        "Successful Token Generation",
			userID:      1234,
			currentTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			setupEnv:    func() { os.Setenv("JWT_SECRET", "test_secret") },
			wantErr:     false,
		},
		{
			name:        "Token Expiration Time",
			userID:      5678,
			currentTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			setupEnv:    func() { os.Setenv("JWT_SECRET", "test_secret") },
			wantErr:     false,
		},
		{
			name:        "Invalid JWT Secret",
			userID:      9012,
			currentTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			setupEnv:    func() { os.Unsetenv("JWT_SECRET") },
			wantErr:     true,
		},
		{
			name:        "Very Large User ID",
			userID:      ^uint(0),
			currentTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			setupEnv:    func() { os.Setenv("JWT_SECRET", "test_secret") },
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()

			token, err := generateToken(tt.userID, tt.currentTime)

			if (err != nil) != tt.wantErr {
				t.Errorf("generateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if token == "" {
					t.Error("generateToken() returned an empty token")
				}

				parsedToken, err := jwt.ParseWithClaims(token, &claims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte(os.Getenv("JWT_SECRET")), nil
				})

				if err != nil {
					t.Errorf("Failed to parse token: %v", err)
				}

				if claims, ok := parsedToken.Claims.(*claims); ok && parsedToken.Valid {
					if claims.UserID != tt.userID {
						t.Errorf("Token UserID = %v, want %v", claims.UserID, tt.userID)
					}

					expectedExpiration := tt.currentTime.Add(time.Hour * 72).Unix()
					if claims.ExpiresAt != expectedExpiration {
						t.Errorf("Token ExpiresAt = %v, want %v", claims.ExpiresAt, expectedExpiration)
					}
				} else {
					t.Error("Failed to extract claims from token")
				}
			}
		})
	}

	t.Run("Token Uniqueness", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "test_secret")
		userID := uint(1234)
		time1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		time2 := time1.Add(time.Hour)

		token1, _ := generateToken(userID, time1)
		token2, _ := generateToken(userID, time2)

		if token1 == token2 {
			t.Error("Tokens generated at different times should be unique")
		}
	})
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
				md := metadata.Pairs("authorization", "Token "+tokenString)
				return metadata.NewIncomingContext(context.Background(), md)
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
				md := metadata.Pairs("authorization", "Token "+tokenString)
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedUserID: 0,
			expectedError:  "token expired",
		},
		{
			name: "Malformed Token",
			setupContext: func() context.Context {
				md := metadata.Pairs("authorization", "Token malformed.token.here")
				return metadata.NewIncomingContext(context.Background(), md)
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
				md := metadata.Pairs("authorization", "Token "+tokenString)
				return metadata.NewIncomingContext(context.Background(), md)
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
				md := metadata.Pairs("authorization", "Token "+tokenString)
				return metadata.NewIncomingContext(context.Background(), md)
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
				md := metadata.Pairs("authorization", "Token "+tokenString)
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedUserID: 0,
			expectedError:  "invalid token: couldn't handle this token",
		},
		{
			name: "Environmental Variable Dependency",
			setupContext: func() context.Context {
				os.Setenv("JWT_SECRET", "")
				jwtSecret = []byte("")
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
					UserID: 101,
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: time.Now().Add(time.Hour).Unix(),
					},
				})
				tokenString, _ := token.SignedString([]byte("test_secret"))
				md := metadata.Pairs("authorization", "Token "+tokenString)
				return metadata.NewIncomingContext(context.Background(), md)
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

