package store

import (
	"errors"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
)





type mockDB struct {
	createFunc func(interface{}) *gorm.DB
}
type mockDB struct {
	users  []model.User
	err    error
	called bool
}
type mockDB struct {
	users map[string]*model.User
	err   error
}


/*
ROOST_METHOD_HASH=Create_889fc0fc45
ROOST_METHOD_SIG_HASH=Create_4c48ec3920

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error 

*/
func (m *mockDB) Create(value interface{}) *gorm.DB {
	return m.createFunc(value)
}

func TestUserStoreCreate(t *testing.T) {
	tests := []struct {
		name    string
		user    *model.User
		mockDB  func() *mockDB
		wantErr bool
	}{
		{
			name: "Successfully Create a New User",
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockDB: func() *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: nil}
					},
				}
			},
			wantErr: false,
		},
		{
			name: "Attempt to Create a User with Duplicate Username",
			user: &model.User{
				Username: "existinguser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockDB: func() *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: gorm.ErrRecordNotFound}
					},
				}
			},
			wantErr: true,
		},
		{
			name: "Create User with Minimum Required Fields",
			user: &model.User{
				Username: "minimaluser",
				Email:    "minimal@example.com",
				Password: "password123",
			},
			mockDB: func() *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: nil}
					},
				}
			},
			wantErr: false,
		},
		{
			name: "Attempt to Create User with Invalid Email Format",
			user: &model.User{
				Username: "invaliduser",
				Email:    "invalidemail",
				Password: "password123",
			},
			mockDB: func() *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: errors.New("invalid email format")}
					},
				}
			},
			wantErr: true,
		},
		{
			name: "Handle Database Connection Error During User Creation",
			user: &model.User{
				Username: "connectionerroruser",
				Email:    "connection@example.com",
				Password: "password123",
			},
			mockDB: func() *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: errors.New("database connection error")}
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := tt.mockDB()
			s := &UserStore{
				db: mockDB,
			}

			err := s.Create(tt.user)

			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetByEmail_3574af40e5
ROOST_METHOD_SIG_HASH=GetByEmail_5731b833c1

FUNCTION_DEF=func (s *UserStore) GetByEmail(email string) (*model.User, error) 

*/
func TestUserStoreGetByEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		mockUsers     []model.User
		mockErr       error
		expectedUser  *model.User
		expectedError error
	}{
		{
			name:  "Successfully retrieve a user by email",
			email: "user@example.com",
			mockUsers: []model.User{
				{Model: gorm.Model{ID: 1}, Email: "user@example.com", Username: "testuser"},
			},
			expectedUser:  &model.User{Model: gorm.Model{ID: 1}, Email: "user@example.com", Username: "testuser"},
			expectedError: nil,
		},
		{
			name:          "Attempt to retrieve a non-existent user",
			email:         "nonexistent@example.com",
			mockUsers:     []model.User{},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:          "Handle database connection error",
			email:         "user@example.com",
			mockErr:       errors.New("database connection error"),
			expectedUser:  nil,
			expectedError: errors.New("database connection error"),
		},
		{
			name:          "Retrieve user with empty email string",
			email:         "",
			mockUsers:     []model.User{},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:  "Case sensitivity in email lookup",
			email: "USER@EXAMPLE.COM",
			mockUsers: []model.User{
				{Model: gorm.Model{ID: 1}, Email: "user@example.com", Username: "testuser"},
			},
			expectedUser:  &model.User{Model: gorm.Model{ID: 1}, Email: "user@example.com", Username: "testuser"},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{
				users: tt.mockUsers,
				err:   tt.mockErr,
			}
			store := &UserStore{db: mockDB}

			user, err := store.GetByEmail(tt.email)

			if !mockDB.called {
				t.Error("Expected database query, but it was not called")
			}

			if (err != nil) != (tt.expectedError != nil) {
				t.Errorf("GetByEmail() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if err != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("GetByEmail() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if tt.expectedUser == nil && user != nil {
				t.Errorf("GetByEmail() user = %v, expected nil", user)
				return
			}

			if tt.expectedUser != nil {
				if user == nil {
					t.Error("GetByEmail() returned nil user, expected non-nil")
					return
				}
				if user.ID != tt.expectedUser.ID || user.Email != tt.expectedUser.Email || user.Username != tt.expectedUser.Username {
					t.Errorf("GetByEmail() user = %v, expected %v", user, tt.expectedUser)
				}
			}
		})
	}
}


/*
ROOST_METHOD_HASH=GetByUsername_f11f114df2
ROOST_METHOD_SIG_HASH=GetByUsername_954d096e24

FUNCTION_DEF=func (s *UserStore) GetByUsername(username string) (*model.User, error) 

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
					"testuser": {Username: "testuser", Email: "test@example.com"},
				},
			},
			want:    &model.User{Username: "testuser", Email: "test@example.com"},
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
			name:     "Retrieve a user with maximum length username",
			username: string(make([]byte, 255)),
			mockDB: &mockDB{
				users: map[string]*model.User{
					string(make([]byte, 255)): {Username: string(make([]byte, 255)), Email: "max@example.com"},
				},
			},
			want:    &model.User{Username: string(make([]byte, 255)), Email: "max@example.com"},
			wantErr: nil,
		},
		{
			name:     "Attempt retrieval with an empty username",
			username: "",
			mockDB:   &mockDB{users: map[string]*model.User{}},
			want:     nil,
			wantErr:  gorm.ErrRecordNotFound,
		},
		{
			name:     "Handle case sensitivity in username lookup",
			username: "TestUser",
			mockDB: &mockDB{
				users: map[string]*model.User{
					"TestUser": {Username: "TestUser", Email: "test@example.com"},
				},
			},
			want:    &model.User{Username: "TestUser", Email: "test@example.com"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UserStore{
				db: tt.mockDB,
			}
			got, err := s.GetByUsername(tt.username)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("UserStore.GetByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("UserStore.GetByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareUsers(got, tt.want) {
				t.Errorf("UserStore.GetByUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}

