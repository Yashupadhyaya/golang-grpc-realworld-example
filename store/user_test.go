package store

import (
	"errors"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"time"
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
		mockDB  func(user *model.User) *mockDB
		wantErr bool
	}{
		{
			name: "Successfully Create a New User",
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockDB: func(user *model.User) *mockDB {
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
			mockDB: func(user *model.User) *mockDB {
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
			mockDB: func(user *model.User) *mockDB {
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
			mockDB: func(user *model.User) *mockDB {
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
			mockDB: func(user *model.User) *mockDB {
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
			mockDB: func(user *model.User) *mockDB {
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
			mockDB := tt.mockDB(tt.user)
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
		db      *mockDB
		email   string
		want    *model.User
		wantErr error
	}{
		{
			name: "Successfully retrieve a user by email",
			db: &mockDB{
				users: []model.User{
					{Model: gorm.Model{ID: 1}, Email: "user@example.com", Username: "testuser"},
				},
			},
			email: "user@example.com",
			want:  &model.User{Model: gorm.Model{ID: 1}, Email: "user@example.com", Username: "testuser"},
		},
		{
			name:    "Attempt to retrieve a non-existent user",
			db:      &mockDB{},
			email:   "nonexistent@example.com",
			wantErr: gorm.ErrRecordNotFound,
		},
		{
			name: "Handle database connection error",
			db: &mockDB{
				err: errors.New("database connection error"),
			},
			email:   "user@example.com",
			wantErr: errors.New("database connection error"),
		},
		{
			name:    "Retrieve user with empty email string",
			db:      &mockDB{},
			email:   "",
			wantErr: gorm.ErrRecordNotFound,
		},
		{
			name: "Case sensitivity in email lookup",
			db: &mockDB{
				users: []model.User{
					{Model: gorm.Model{ID: 1}, Email: "User@Example.com", Username: "testuser"},
				},
			},
			email:   "user@example.com",
			want:    nil,
			wantErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UserStore{
				db: tt.db,
			}
			got, err := s.GetByEmail(tt.email)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("UserStore.GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("UserStore.GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareUsers(got, tt.want) {
				t.Errorf("UserStore.GetByEmail() = %v, want %v", got, tt.want)
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
		mockDB   func() *mockDB
		want     *model.User
		wantErr  error
	}{
		{
			name:     "Successfully retrieve an existing user",
			username: "existinguser",
			mockDB: func() *mockDB {
				return &mockDB{
					findFunc: func(out interface{}, where ...interface{}) *gorm.DB {
						*(out.(*model.User)) = model.User{Username: "existinguser"}
						return &gorm.DB{Error: nil}
					},
				}
			},
			want:    &model.User{Username: "existinguser"},
			wantErr: nil,
		},
		{
			name:     "Attempt to retrieve a non-existent user",
			username: "nonexistentuser",
			mockDB: func() *mockDB {
				return &mockDB{
					findFunc: func(out interface{}, where ...interface{}) *gorm.DB {
						return &gorm.DB{Error: gorm.ErrRecordNotFound}
					},
				}
			},
			want:    nil,
			wantErr: gorm.ErrRecordNotFound,
		},
		{
			name:     "Handle database connection error",
			username: "anyuser",
			mockDB: func() *mockDB {
				return &mockDB{
					findFunc: func(out interface{}, where ...interface{}) *gorm.DB {
						return &gorm.DB{Error: errors.New("database connection error")}
					},
				}
			},
			want:    nil,
			wantErr: errors.New("database connection error"),
		},
		{
			name:     "Retrieve user with maximum length username",
			username: string(make([]byte, 255)),
			mockDB: func() *mockDB {
				return &mockDB{
					findFunc: func(out interface{}, where ...interface{}) *gorm.DB {
						*(out.(*model.User)) = model.User{Username: string(make([]byte, 255))}
						return &gorm.DB{Error: nil}
					},
				}
			},
			want:    &model.User{Username: string(make([]byte, 255))},
			wantErr: nil,
		},
		{
			name:     "Attempt to retrieve user with empty username",
			username: "",
			mockDB: func() *mockDB {
				return &mockDB{
					findFunc: func(out interface{}, where ...interface{}) *gorm.DB {
						return &gorm.DB{Error: gorm.ErrRecordNotFound}
					},
				}
			},
			want:    nil,
			wantErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UserStore{
				db: tt.mockDB(),
			}
			got, err := s.GetByUsername(tt.username)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("UserStore.GetByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("UserStore.GetByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got != nil) != (tt.want != nil) {
				t.Errorf("UserStore.GetByUsername() got = %v, want %v", got, tt.want)
				return
			}
			if got != nil && tt.want != nil && got.Username != tt.want.Username {
				t.Errorf("UserStore.GetByUsername() got = %v, want %v", got, tt.want)
			}
		})
	}
}

