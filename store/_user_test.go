package store

import (
	"errors"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"time"
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
			name: "Create User with Maximum Field Lengths",
			user: &model.User{
				Username: string(make([]byte, 255)),
				Email:    string(make([]byte, 255)) + "@example.com",
				Password: string(make([]byte, 255)),
				Bio:      string(make([]byte, 1000)),
				Image:    string(make([]byte, 1000)),
			},
			dbError: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create a Duplicate User",
			user: &model.User{
				Username: "existinguser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			dbError: errors.New("unique constraint violation"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			userStore := &UserStore{db: mockDB}

			mockDB.On("Create", tt.user).Return(&gorm.DB{Error: tt.dbError})

			err := userStore.Create(tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.dbError, err)
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
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
		name          string
		email         string
		mockSetup     func(*mockDB)
		expectedUser  *model.User
		expectedError error
	}{
		{
			name:  "Successfully retrieve a user by email",
			email: "test@example.com",
			mockSetup: func(m *mockDB) {
				m.On("Where", "email = ?", "test@example.com").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:    gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Email:    "test@example.com",
						Username: "testuser",
					}
				}).Return(&gorm.DB{Error: nil})
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Email:    "test@example.com",
				Username: "testuser",
			},
			expectedError: nil,
		},
		{
			name:  "Attempt to retrieve a non-existent user",
			email: "nonexistent@example.com",
			mockSetup: func(m *mockDB) {
				m.On("Where", "email = ?", "nonexistent@example.com").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:  "Handle database connection error",
			email: "test@example.com",
			mockSetup: func(m *mockDB) {
				m.On("Where", "email = ?", "test@example.com").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Return(&gorm.DB{Error: errors.New("database connection error")})
			},
			expectedUser:  nil,
			expectedError: errors.New("database connection error"),
		},
		{
			name:  "Retrieve user with special characters in email",
			email: "user+test@example.com",
			mockSetup: func(m *mockDB) {
				m.On("Where", "email = ?", "user+test@example.com").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:    gorm.Model{ID: 2, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Email:    "user+test@example.com",
						Username: "specialuser",
					}
				}).Return(&gorm.DB{Error: nil})
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 2},
				Email:    "user+test@example.com",
				Username: "specialuser",
			},
			expectedError: nil,
		},
		{
			name:  "Handle empty email input",
			email: "",
			mockSetup: func(m *mockDB) {
				m.On("Where", "email = ?", "").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			tt.mockSetup(mockDB)

			userStore := &UserStore{db: mockDB}

			user, err := userStore.GetByEmail(tt.email)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedUser, user)

			mockDB.AssertExpectations(t)
		})
	}
}

func (m *mockDB) First(out interface{}, where ...interface{}) *gorm.DB {
	return m.Called(out, where).Get(0).(*gorm.DB)
}

func (m *mockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	return m.Called(query, args).Get(0).(*gorm.DB)
}


/*
ROOST_METHOD_HASH=UserStore_GetByUsername_622b1b9e41
ROOST_METHOD_SIG_HASH=UserStore_GetByUsername_992f00baec

FUNCTION_DEF=func (s *UserStore) GetByUsername(username string) (*model.User, error) // GetByUsername finds a user from username


*/
func TestUserStoreGetByUsername(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		mockSetup     func(*mockDB)
		expectedUser  *model.User
		expectedError error
	}{
		{
			name:     "Successfully retrieve a user by username",
			username: "testuser",
			mockSetup: func(m *mockDB) {
				m.On("Where", "username = ?", "testuser").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:    gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Username: "testuser",
						Email:    "test@example.com",
					}
				}).Return(&gorm.DB{Error: nil})
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "test@example.com",
			},
			expectedError: nil,
		},
		{
			name:     "Attempt to retrieve a non-existent user",
			username: "nonexistent",
			mockSetup: func(m *mockDB) {
				m.On("Where", "username = ?", "nonexistent").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:     "Database connection error",
			username: "testuser",
			mockSetup: func(m *mockDB) {
				m.On("Where", "username = ?", "testuser").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Return(&gorm.DB{Error: errors.New("database connection error")})
			},
			expectedUser:  nil,
			expectedError: errors.New("database connection error"),
		},
		{
			name:     "Case sensitivity in username lookup",
			username: "TestUser",
			mockSetup: func(m *mockDB) {
				m.On("Where", "username = ?", "TestUser").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(mockDB)
			tt.mockSetup(mockDB)

			userStore := &UserStore{db: mockDB}

			user, err := userStore.GetByUsername(tt.username)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedUser, user)

			mockDB.AssertExpectations(t)
		})
	}
}

