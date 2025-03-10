package store

import (
	errors "errors"
	testing "testing"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	assert "github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	time "time"
	sync "sync"
)





type MockDB struct {
	mock.Mock
}
type mockDB struct {
	whereFunc func(query interface{}, args ...interface{}) *gorm.DB
	firstFunc func(out interface{}) *gorm.DB
}


/*
ROOST_METHOD_HASH=UserStore_Create_9495ddb29d
ROOST_METHOD_SIG_HASH=UserStore_Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user


*/
func (m *MockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
}

func TestUserStoreCreate(t *testing.T) {
	tests := []struct {
		name    string
		user    *model.User
		dbError error
		wantErr bool
	}{
		{
			name: "Successfully Create a New User",
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create a User with Duplicate Email",
			user: &model.User{
				Username: "duplicateuser",
				Email:    "duplicate@example.com",
				Password: "password123",
			},
			dbError: gorm.ErrRecordNotFound,
			wantErr: true,
		},
		{
			name: "Create User with Minimum Required Fields",
			user: &model.User{
				Username: "minimaluser",
				Email:    "minimal@example.com",
				Password: "password123",
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create User with Invalid Data",
			user: &model.User{
				Username: "invaliduser",
				Email:    "invalid-email",
				Password: "password123",
			},
			dbError: errors.New("validation error"),
			wantErr: true,
		},
		{
			name: "Database Connection Error During User Creation",
			user: &model.User{
				Username: "erroruser",
				Email:    "error@example.com",
				Password: "password123",
			},
			dbError: errors.New("database connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			userStore := &UserStore{db: mockDB}

			mockDB.On("Create", tt.user).Return(&gorm.DB{Error: tt.dbError})

			err := userStore.Create(tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.dbError != nil {
					assert.Equal(t, tt.dbError, err)
				}
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_GetByEmail_fda09af5c4
ROOST_METHOD_SIG_HASH=UserStore_GetByEmail_9e84f3286b

FUNCTION_DEF=func (s *UserStore) GetByEmail(email string) (*model.User, error) // GetByEmail finds a user from email


*/
func TestUserStoreGetByEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		mockDB   *mockDB
		expected *model.User
		wantErr  bool
		err      error
	}{
		{
			name:  "Successfully retrieve a user by email",
			email: "test@example.com",
			mockDB: &mockDB{
				whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
					return &gorm.DB{}
				},
				firstFunc: func(out interface{}) *gorm.DB {
					*(out.(*model.User)) = model.User{
						Model:    gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Username: "testuser",
						Email:    "test@example.com",
						Password: "hashedpassword",
						Bio:      "Test bio",
						Image:    "http://example.com/image.jpg",
					}
					return &gorm.DB{}
				},
			},
			expected: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "test@example.com",
				Password: "hashedpassword",
				Bio:      "Test bio",
				Image:    "http://example.com/image.jpg",
			},
			wantErr: false,
		},
		{
			name:  "Attempt to retrieve a non-existent user",
			email: "nonexistent@example.com",
			mockDB: &mockDB{
				whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
					return &gorm.DB{}
				},
				firstFunc: func(out interface{}) *gorm.DB {
					return &gorm.DB{Error: gorm.ErrRecordNotFound}
				},
			},
			expected: nil,
			wantErr:  true,
			err:      gorm.ErrRecordNotFound,
		},
		{
			name:  "Handle database connection error",
			email: "test@example.com",
			mockDB: &mockDB{
				whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
					return &gorm.DB{Error: errors.New("database connection error")}
				},
				firstFunc: func(out interface{}) *gorm.DB {
					return &gorm.DB{Error: errors.New("database connection error")}
				},
			},
			expected: nil,
			wantErr:  true,
			err:      errors.New("database connection error"),
		},
		{
			name:  "Retrieve user with special characters in email",
			email: "user+test@example.com",
			mockDB: &mockDB{
				whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
					return &gorm.DB{}
				},
				firstFunc: func(out interface{}) *gorm.DB {
					*(out.(*model.User)) = model.User{
						Model:    gorm.Model{ID: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Username: "specialuser",
						Email:    "user+test@example.com",
						Password: "hashedpassword",
						Bio:      "Special user bio",
						Image:    "http://example.com/special.jpg",
					}
					return &gorm.DB{}
				},
			},
			expected: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "specialuser",
				Email:    "user+test@example.com",
				Password: "hashedpassword",
				Bio:      "Special user bio",
				Image:    "http://example.com/special.jpg",
			},
			wantErr: false,
		},
		{
			name:  "Case sensitivity in email lookup",
			email: "TEST@EXAMPLE.COM",
			mockDB: &mockDB{
				whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
					return &gorm.DB{}
				},
				firstFunc: func(out interface{}) *gorm.DB {
					*(out.(*model.User)) = model.User{
						Model:    gorm.Model{ID: 3, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Username: "caseuser",
						Email:    "test@example.com",
						Password: "hashedpassword",
						Bio:      "Case sensitive user",
						Image:    "http://example.com/case.jpg",
					}
					return &gorm.DB{}
				},
			},
			expected: &model.User{
				Model:    gorm.Model{ID: 3},
				Username: "caseuser",
				Email:    "test@example.com",
				Password: "hashedpassword",
				Bio:      "Case sensitive user",
				Image:    "http://example.com/case.jpg",
			},
			wantErr: false,
		},
		{
			name:  "Handling of empty email string",
			email: "",
			mockDB: &mockDB{
				whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
					return &gorm.DB{}
				},
				firstFunc: func(out interface{}) *gorm.DB {
					return &gorm.DB{Error: gorm.ErrRecordNotFound}
				},
			},
			expected: nil,
			wantErr:  true,
			err:      gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &UserStore{
				db: tt.mockDB,
			}

			user, err := store.GetByEmail(tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.err, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, user)
		})
	}
}

func (m *mockDB) First(out interface{}) *gorm.DB {
	return m.firstFunc(out)
}

func (m *mockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	return m.whereFunc(query, args...)
}


/*
ROOST_METHOD_HASH=UserStore_GetByUsername_622b1b9e41
ROOST_METHOD_SIG_HASH=UserStore_GetByUsername_992f00baec

FUNCTION_DEF=func (s *UserStore) GetByUsername(username string) (*model.User, error) // GetByUsername finds a user from username


*/
func TestUserStoreGetByUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		mockDB   *mockDB
		want     *model.User
		wantErr  error
	}{
		{
			name:     "Successfully retrieve a user by username",
			username: "testuser",
			mockDB: &mockDB{
				users: map[string]*model.User{
					"testuser": {
						Model:    gorm.Model{ID: 1},
						Username: "testuser",
						Email:    "test@example.com",
					},
				},
			},
			want: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "test@example.com",
			},
			wantErr: nil,
		},
		{
			name:     "Attempt to retrieve a non-existent user",
			username: "nonexistent",
			mockDB:   &mockDB{users: map[string]*model.User{}},
			want:     nil,
			wantErr:  gorm.ErrRecordNotFound,
		},
		{
			name:     "Handle database connection error",
			username: "testuser",
			mockDB:   &mockDB{err: errors.New("database connection error")},
			want:     nil,
			wantErr:  errors.New("database connection error"),
		},
		{
			name:     "Retrieve user with special characters in username",
			username: "user@example.com",
			mockDB: &mockDB{
				users: map[string]*model.User{
					"user@example.com": {
						Model:    gorm.Model{ID: 2},
						Username: "user@example.com",
						Email:    "user@example.com",
					},
				},
			},
			want: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "user@example.com",
				Email:    "user@example.com",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &UserStore{db: tt.mockDB}
			got, err := store.GetByUsername(tt.username)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserStoreGetByUsernameConcurrent(t *testing.T) {
	users := map[string]*model.User{
		"user1": {Model: gorm.Model{ID: 1}, Username: "user1", Email: "user1@example.com"},
		"user2": {Model: gorm.Model{ID: 2}, Username: "user2", Email: "user2@example.com"},
		"user3": {Model: gorm.Model{ID: 3}, Username: "user3", Email: "user3@example.com"},
	}

	mockDB := &mockDB{users: users}
	store := &UserStore{db: mockDB}

	var wg sync.WaitGroup
	wg.Add(3)

	for _, username := range []string{"user1", "user2", "user3"} {
		go func(u string) {
			defer wg.Done()
			user, err := store.GetByUsername(u)
			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, u, user.Username)
		}(username)
	}

	wg.Wait()
}

func TestUserStoreGetByUsernamePerformance(t *testing.T) {
	users := make(map[string]*model.User)
	for i := 0; i < 100000; i++ {
		username := fmt.Sprintf("user%d", i)
		users[username] = &model.User{
			Model:    gorm.Model{ID: uint(i)},
			Username: username,
			Email:    fmt.Sprintf("%s@example.com", username),
		}
	}

	mockDB := &mockDB{users: users}
	store := &UserStore{db: mockDB}

	start := time.Now()
	user, err := store.GetByUsername("user99999")
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user99999", user.Username)
	assert.Less(t, duration, 100*time.Millisecond)
}

