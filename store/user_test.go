package store

import (
	"errors"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"reflect"
)








/*
ROOST_METHOD_HASH=Create_9495ddb29d
ROOST_METHOD_SIG_HASH=Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user


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
						return &gorm.DB{Error: errors.New("duplicate key error")}
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
				Password: "minpass",
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
				Username: "invalidemail",
				Email:    "invalid-email",
				Password: "password123",
			},
			mockDB: func() *mockDB {
				return &mockDB{
					createFunc: func(value interface{}) *gorm.DB {
						return &gorm.DB{Error: errors.New("validation error")}
					},
				}
			},
			wantErr: true,
		},
		{
			name: "Create User with Maximum Length Values",
			user: &model.User{
				Username: "maxlengthusername",
				Email:    "maxlength@example.com",
				Password: "maxlengthpassword",
				Bio:      "This is a very long bio with maximum allowed characters.",
				Image:    "https://example.com/very-long-image-url-with-maximum-allowed-characters.jpg",
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
			name: "Database Connection Failure During User Creation",
			user: &model.User{
				Username: "connectionfailure",
				Email:    "failure@example.com",
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
ROOST_METHOD_HASH=GetByEmail_fda09af5c4
ROOST_METHOD_SIG_HASH=GetByEmail_9e84f3286b

FUNCTION_DEF=func (s *UserStore) GetByEmail(email string) (*model.User, error) // GetByEmail finds a user from email


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
			email: "test@example.com",
			mockDB: &mockDB{
				users: map[string]*model.User{
					"test@example.com": {Email: "test@example.com", Username: "testuser"},
				},
			},
			want:    &model.User{Email: "test@example.com", Username: "testuser"},
			wantErr: nil,
		},
		{
			name:    "Attempt to retrieve a non-existent user",
			email:   "nonexistent@example.com",
			mockDB:  &mockDB{users: map[string]*model.User{}},
			want:    nil,
			wantErr: gorm.ErrRecordNotFound,
		},
		{
			name:    "Handle database connection error",
			email:   "test@example.com",
			mockDB:  &mockDB{dbErr: errors.New("database connection error")},
			want:    nil,
			wantErr: errors.New("database connection error"),
		},
		{
			name:  "Retrieve user with a case-insensitive email match",
			email: "TEST@example.com",
			mockDB: &mockDB{
				users: map[string]*model.User{
					"test@example.com": {Email: "test@example.com", Username: "testuser"},
				},
			},
			want:    &model.User{Email: "test@example.com", Username: "testuser"},
			wantErr: nil,
		},
		{
			name:    "Handle empty email input",
			email:   "",
			mockDB:  &mockDB{users: map[string]*model.User{}},
			want:    nil,
			wantErr: gorm.ErrRecordNotFound,
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserStore.GetByEmail() = %v, want %v", got, tt.want)
			}
			if !tt.mockDB.called {
				t.Errorf("UserStore.GetByEmail() did not call the database")
			}
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
		name     string
		username string
		mockDB   *mockDB
		want     *model.User
		wantErr  error
	}{
		{
			name:     "Successfully retrieve an existing user",
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
			name:     "Retrieve user with special characters in username",
			username: "user@example.com",
			mockDB: &mockDB{
				users: map[string]*model.User{
					"user@example.com": {Username: "user@example.com", Email: "user@example.com"},
				},
			},
			want:    &model.User{Username: "user@example.com", Email: "user@example.com"},
			wantErr: nil,
		},
		{
			name:     "Handle case-sensitive usernames",
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

