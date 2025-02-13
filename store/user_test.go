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








/*
ROOST_METHOD_HASH=Create_9495ddb29d
ROOST_METHOD_SIG_HASH=Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user


*/
func (m *mockDB) Create(value interface{}) *gorm.DB {
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
				Username: "testuser" + string(make([]byte, 50)),
				Email:    "test@example.com" + string(make([]byte, 100)),
				Password: string(make([]byte, 100)),
				Bio:      string(make([]byte, 500)),
				Image:    string(make([]byte, 200)),
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


/*
ROOST_METHOD_HASH=GetByEmail_fda09af5c4
ROOST_METHOD_SIG_HASH=GetByEmail_9e84f3286b

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
			email: "user@example.com",
			mockSetup: func(m *mockDB) {
				m.On("Where", "email = ?", "user@example.com").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:    gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Username: "testuser",
						Email:    "user@example.com",
						Password: "hashedpassword",
						Bio:      "Test bio",
						Image:    "test-image.jpg",
					}
				}).Return(&gorm.DB{Error: nil})
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "user@example.com",
				Password: "hashedpassword",
				Bio:      "Test bio",
				Image:    "test-image.jpg",
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
			email: "user@example.com",
			mockSetup: func(m *mockDB) {
				m.On("Where", "email = ?", "user@example.com").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Return(&gorm.DB{Error: errors.New("database connection error")})
			},
			expectedUser:  nil,
			expectedError: errors.New("database connection error"),
		},
		{
			name:  "Retrieve user with empty email string",
			email: "",
			mockSetup: func(m *mockDB) {
				m.On("Where", "email = ?", "").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:  "Handle case-insensitive email matching",
			email: "User@Example.com",
			mockSetup: func(m *mockDB) {
				m.On("Where", "email = ?", "User@Example.com").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{
						Model:    gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
						Username: "testuser",
						Email:    "User@Example.com",
						Password: "hashedpassword",
						Bio:      "Test bio",
						Image:    "test-image.jpg",
					}
				}).Return(&gorm.DB{Error: nil})
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "User@Example.com",
				Password: "hashedpassword",
				Bio:      "Test bio",
				Image:    "test-image.jpg",
			},
			expectedError: nil,
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


/*
ROOST_METHOD_HASH=GetByUsername_622b1b9e41
ROOST_METHOD_SIG_HASH=GetByUsername_992f00baec

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
					*arg = model.User{Username: "testuser", Email: "test@example.com"}
				}).Return(&gorm.DB{Error: nil})
			},
			expectedUser:  &model.User{Username: "testuser", Email: "test@example.com"},
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
			name:     "Handle database connection error",
			username: "anyuser",
			mockSetup: func(m *mockDB) {
				m.On("Where", "username = ?", "anyuser").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Return(&gorm.DB{Error: errors.New("database connection error")})
			},
			expectedUser:  nil,
			expectedError: errors.New("database connection error"),
		},
		{
			name:     "Retrieve user with special characters in username",
			username: "user@special",
			mockSetup: func(m *mockDB) {
				m.On("Where", "username = ?", "user@special").Return(m)
				m.On("First", mock.AnythingOfType("*model.User"), mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(0).(*model.User)
					*arg = model.User{Username: "user@special", Email: "special@example.com"}
				}).Return(&gorm.DB{Error: nil})
			},
			expectedUser:  &model.User{Username: "user@special", Email: "special@example.com"},
			expectedError: nil,
		},
		{
			name:     "Case sensitivity check",
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

			store := &UserStore{db: mockDB}

			user, err := store.GetByUsername(tt.username)

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

