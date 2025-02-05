package store

import (
	"errors"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
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
		mockErr error
		wantErr bool
	}{
		{
			name: "Successfully Create a New User",
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create a User with Duplicate Username",
			user: &model.User{
				Username: "existinguser",
				Email:    "new@example.com",
				Password: "password123",
			},
			mockErr: errors.New("duplicate username"),
			wantErr: true,
		},
		{
			name: "Attempt to Create a User with Duplicate Email",
			user: &model.User{
				Username: "newuser",
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockErr: errors.New("duplicate email"),
			wantErr: true,
		},
		{
			name: "Create User with Minimum Required Fields",
			user: &model.User{
				Username: "minuser",
				Email:    "min@example.com",
				Password: "password123",
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "Attempt to Create User with Invalid Email Format",
			user: &model.User{
				Username: "invaliduser",
				Email:    "invalid-email",
				Password: "password123",
			},
			mockErr: errors.New("invalid email format"),
			wantErr: true,
		},
		{
			name: "Database Connection Error During User Creation",
			user: &model.User{
				Username: "erroruser",
				Email:    "error@example.com",
				Password: "password123",
			},
			mockErr: errors.New("database connection error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{
				createFunc: func(value interface{}) *gorm.DB {
					return &gorm.DB{Error: tt.mockErr}
				},
			}

			store := &UserStore{
				db: mockDB,
			}

			err := store.Create(tt.user)

			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err != tt.mockErr {
				t.Errorf("UserStore.Create() error = %v, expected error %v", err, tt.mockErr)
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
				users: []model.User{{Email: "test@example.com", Username: "testuser"}},
			},
			email: "test@example.com",
			want:  &model.User{Email: "test@example.com", Username: "testuser"},
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
				dbErr: errors.New("database connection error"),
			},
			email:   "test@example.com",
			wantErr: errors.New("database connection error"),
		},
		{
			name: "Retrieve user with a case-insensitive email match",
			db: &mockDB{
				users: []model.User{{Email: "Test@Example.com", Username: "testuser"}},
			},
			email: "test@example.com",
			want:  &model.User{Email: "Test@Example.com", Username: "testuser"},
		},
		{
			name:    "Handle empty email input",
			db:      &mockDB{},
			email:   "",
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
			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("UserStore.GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.db.called {
				t.Error("UserStore.GetByEmail() did not call database methods")
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
			username: "existinguser",
			mockDB: &mockDB{
				users: map[string]*model.User{
					"existinguser": {Username: "existinguser", Email: "user@example.com"},
				},
			},
			want:    &model.User{Username: "existinguser", Email: "user@example.com"},
			wantErr: nil,
		},
		{
			name:     "Attempt to retrieve a non-existent user",
			username: "nonexistentuser",
			mockDB:   &mockDB{users: map[string]*model.User{}},
			want:     nil,
			wantErr:  gorm.ErrRecordNotFound,
		},
		{
			name:     "Handle database connection error",
			username: "anyuser",
			mockDB:   &mockDB{err: errors.New("database connection error")},
			want:     nil,
			wantErr:  errors.New("database connection error"),
		},
		{
			name:     "Retrieve user with maximum length username",
			username: "maxlengthusername1234567890",
			mockDB: &mockDB{
				users: map[string]*model.User{
					"maxlengthusername1234567890": {Username: "maxlengthusername1234567890", Email: "max@example.com"},
				},
			},
			want:    &model.User{Username: "maxlengthusername1234567890", Email: "max@example.com"},
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
			name:     "Case sensitivity check",
			username: "EXISTINGUSER",
			mockDB: &mockDB{
				users: map[string]*model.User{
					"existinguser": {Username: "existinguser", Email: "user@example.com"},
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

