package store

import (
	errors "errors"
	testing "testing"
	time "time"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	assert "github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)





type mockDB struct {
	mock.Mock
}


/*
ROOST_METHOD_HASH=UserStore_Create_9495ddb29d
ROOST_METHOD_SIG_HASH=UserStore_Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user


*/
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
				Bio:      "Test bio",
				Image:    "http://example.com/image.jpg",
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create a User with Invalid Data",
			user: &model.User{
				Username: "",
				Email:    "test@example.com",
				Password: "password123",
			},
			dbError: errors.New("validation error"),
			wantErr: true,
		},
		{
			name: "Handle Database Connection Error",
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			dbError: errors.New("database connection error"),
			wantErr: true,
		},
		{
			name: "Create User with Existing Username or Email",
			user: &model.User{
				Username: "existinguser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			dbError: errors.New("unique constraint violation"),
			wantErr: true,
		},
		{
			name: "Create User with Maximum Length Fields",
			user: &model.User{
				Username: "testuser_with_maximum_length_username",
				Email:    "very_long_email_address@very_long_domain_name.com",
				Password: "very_long_password_that_meets_maximum_length_requirements",
				Bio:      "This is a very long bio that tests the maximum length of the bio field in the database. It should be long enough to test any potential issues with storing long text.",
				Image:    "https://very_long_image_url.com/with_many_parameters_and_a_long_file_name_to_test_maximum_length.jpg",
			},
			dbError: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			userStore := &UserStore{db: mockDB}

			mockDB.On("Create", mock.AnythingOfType("*model.User")).Return(&gorm.DB{Error: tt.dbError})

			err := userStore.Create(tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.dbError, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertCalled(t, "Create", tt.user)

			if !tt.wantErr {
				assert.NotZero(t, tt.user.CreatedAt)
				assert.NotZero(t, tt.user.UpdatedAt)
			}
		})
	}
}

func TestUserStoreCreateAutomaticTimestamp(t *testing.T) {
	mockDB := new(mockDB)
	userStore := &UserStore{db: mockDB}

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	mockDB.On("Create", mock.AnythingOfType("*model.User")).Run(func(args mock.Arguments) {
		u := args.Get(0).(*model.User)
		u.CreatedAt = time.Now()
		u.UpdatedAt = time.Now()
	}).Return(&gorm.DB{Error: nil})

	err := userStore.Create(user)

	assert.NoError(t, err)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.UpdatedAt)

	mockDB.AssertCalled(t, "Create", user)
}

func (m *mockDB) Create(value interface{}) *gorm.DB {
	args := m.Called(value)
	return args.Get(0).(*gorm.DB)
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
					*(out.(*model.User)) = model.User{Email: "test@example.com"}
					return &gorm.DB{}
				},
			},
			expected: &model.User{Email: "test@example.com"},
			err:      nil,
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
			err:      errors.New("database connection error"),
		},
		{
			name:  "Retrieve user with a case-insensitive email match",
			email: "TEST@EXAMPLE.COM",
			mockDB: &mockDB{
				whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
					return &gorm.DB{}
				},
				firstFunc: func(out interface{}) *gorm.DB {
					*(out.(*model.User)) = model.User{Email: "test@example.com"}
					return &gorm.DB{}
				},
			},
			expected: &model.User{Email: "test@example.com"},
			err:      nil,
		},
		{
			name:  "Handle empty email input",
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
			err:      gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &UserStore{db: tt.mockDB}
			user, err := store.GetByEmail(tt.email)

			assert.Equal(t, tt.expected, user)
			assert.Equal(t, tt.err, err)
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
			name:     "Successfully Retrieve a User by Username",
			username: "existingUser",
			mockDB: &mockDB{
				whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
					return &gorm.DB{}
				},
				firstFunc: func(out interface{}) *gorm.DB {
					*(out.(*model.User)) = model.User{Username: "existingUser"}
					return &gorm.DB{}
				},
			},
			want:    &model.User{Username: "existingUser"},
			wantErr: nil,
		},
		{
			name:     "Attempt to Retrieve a Non-existent User",
			username: "nonExistentUser",
			mockDB: &mockDB{
				whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
					return &gorm.DB{}
				},
				firstFunc: func(out interface{}) *gorm.DB {
					return &gorm.DB{Error: gorm.ErrRecordNotFound}
				},
			},
			want:    nil,
			wantErr: gorm.ErrRecordNotFound,
		},
		{
			name:     "Handle Database Connection Error",
			username: "anyUser",
			mockDB: &mockDB{
				whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
					return &gorm.DB{Error: errors.New("connection error")}
				},
				firstFunc: func(out interface{}) *gorm.DB {
					return &gorm.DB{Error: errors.New("connection error")}
				},
			},
			want:    nil,
			wantErr: errors.New("connection error"),
		},
		{
			name:     "Retrieve User with Special Characters in Username",
			username: "user@123!",
			mockDB: &mockDB{
				whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
					return &gorm.DB{}
				},
				firstFunc: func(out interface{}) *gorm.DB {
					*(out.(*model.User)) = model.User{Username: "user@123!"}
					return &gorm.DB{}
				},
			},
			want:    &model.User{Username: "user@123!"},
			wantErr: nil,
		},
		{
			name:     "Handle Empty Username Input",
			username: "",
			mockDB: &mockDB{
				whereFunc: func(query interface{}, args ...interface{}) *gorm.DB {
					return &gorm.DB{}
				},
				firstFunc: func(out interface{}) *gorm.DB {
					return &gorm.DB{Error: gorm.ErrRecordNotFound}
				},
			},
			want:    nil,
			wantErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UserStore{
				db: tt.mockDB,
			}
			got, err := s.GetByUsername(tt.username)
			if (err != nil) != (tt.wantErr != nil) || (err != nil && err.Error() != tt.wantErr.Error()) {
				t.Errorf("UserStore.GetByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got != nil && tt.want == nil) || (got == nil && tt.want != nil) || (got != nil && tt.want != nil && got.Username != tt.want.Username) {
				t.Errorf("UserStore.GetByUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}

