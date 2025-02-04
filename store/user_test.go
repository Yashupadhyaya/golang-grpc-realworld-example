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
				Email:    "new@example.com",
				Password: "password123",
			},
			mockDB: func() *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: errors.New("UNIQUE constraint failed: users.username")}
					},
				}
			},
			wantErr: true,
		},
		{
			name: "Attempt to Create a User with Duplicate Email",
			user: &model.User{
				Username: "newuser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockDB: func() *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: errors.New("UNIQUE constraint failed: users.email")}
					},
				}
			},
			wantErr: true,
		},
		{
			name: "Create User with Minimum Required Fields",
			user: &model.User{
				Username: "minuser",
				Email:    "min@example.com",
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
			name: "Attempt to Create User with Invalid Data",
			user: &model.User{
				Username: "",
				Email:    "invalid@example.com",
				Password: "password123",
			},
			mockDB: func() *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: errors.New("validation error: username cannot be empty")}
					},
				}
			},
			wantErr: true,
		},
		{
			name: "Database Connection Error During User Creation",
			user: &model.User{
				Username: "disconnecteduser",
				Email:    "disconnected@example.com",
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
		name    string
		email   string
		mockDB  *mockDB
		want    *model.User
		wantErr error
	}{
		{
			name:  "Successfully retrieve a user by email",
			email: "user@example.com",
			mockDB: &mockDB{
				users: []model.User{{Email: "user@example.com", Username: "testuser"}},
			},
			want:    &model.User{Email: "user@example.com", Username: "testuser"},
			wantErr: nil,
		},
		{
			name:    "Attempt to retrieve a non-existent user",
			email:   "nonexistent@example.com",
			mockDB:  &mockDB{},
			want:    nil,
			wantErr: gorm.ErrRecordNotFound,
		},
		{
			name:    "Handle database connection error",
			email:   "user@example.com",
			mockDB:  &mockDB{err: errors.New("database connection error")},
			want:    nil,
			wantErr: errors.New("database connection error"),
		},
		{
			name:    "Retrieve user with empty email string",
			email:   "",
			mockDB:  &mockDB{},
			want:    nil,
			wantErr: gorm.ErrRecordNotFound,
		},
		{
			name:  "Case sensitivity in email lookup",
			email: "User@Example.com",
			mockDB: &mockDB{
				users: []model.User{{Email: "User@Example.com", Username: "testuser"}},
			},
			want:    &model.User{Email: "User@Example.com", Username: "testuser"},
			wantErr: nil,
		},
		{
			name:  "Handle special characters in email",
			email: "user+test@example.com",
			mockDB: &mockDB{
				users: []model.User{{Email: "user+test@example.com", Username: "testuser"}},
			},
			want:    &model.User{Email: "user+test@example.com", Username: "testuser"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UserStore{
				db: tt.mockDB,
			}
			got, err := s.GetByEmail(tt.email)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("UserStore.GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("UserStore.GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.mockDB.called {
				t.Errorf("UserStore.GetByEmail() database not called")
			}
			if got != nil && tt.want != nil {
				if got.Email != tt.want.Email || got.Username != tt.want.Username {
					t.Errorf("UserStore.GetByEmail() = %v, want %v", got, tt.want)
				}
			} else if (got == nil) != (tt.want == nil) {
				t.Errorf("UserStore.GetByEmail() = %v, want %v", got, tt.want)
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
			name:     "Successfully retrieve a user",
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
			name:     "Non-existent user",
			username: "nonexistent",
			mockDB:   &mockDB{users: map[string]*model.User{}},
			want:     nil,
			wantErr:  gorm.ErrRecordNotFound,
		},
		{
			name:     "Database connection error",
			username: "testuser",
			mockDB:   &mockDB{err: errors.New("connection error")},
			want:     nil,
			wantErr:  errors.New("connection error"),
		},
		{
			name:     "Maximum length username",
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
			name:     "Empty username",
			username: "",
			mockDB:   &mockDB{users: map[string]*model.User{}},
			want:     nil,
			wantErr:  gorm.ErrRecordNotFound,
		},
		{
			name:     "Case-sensitive username lookup",
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
			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("UserStore.GetByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && tt.want != nil {
				if got.Username != tt.want.Username || got.Email != tt.want.Email {
					t.Errorf("UserStore.GetByUsername() = %v, want %v", got, tt.want)
				}
			} else if (got == nil) != (tt.want == nil) {
				t.Errorf("UserStore.GetByUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}

