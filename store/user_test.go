package store

import (
	"errors"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
)





type mockDB struct {
	createFunc func(*model.User) *gorm.DB
}
type mockDB struct {
	users  map[string]*model.User
	dbErr  error
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
	return m.createFunc(value.(*model.User))
}

func TestUserStoreCreate(t *testing.T) {
	tests := []struct {
		name    string
		user    *model.User
		mockDB  func(*model.User) *gorm.DB
		wantErr bool
	}{
		{
			name: "Successfully Create a New User",
			user: &model.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockDB: func(u *model.User) *gorm.DB {
				return &gorm.DB{Error: nil}
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
			mockDB: func(u *model.User) *gorm.DB {
				return &gorm.DB{Error: errors.New("duplicate key violation")}
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
			mockDB: func(u *model.User) *gorm.DB {
				return &gorm.DB{Error: nil}
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
			mockDB: func(u *model.User) *gorm.DB {
				return &gorm.DB{Error: errors.New("validation error: invalid email format")}
			},
			wantErr: true,
		},
		{
			name: "Database Connection Error During User Creation",
			user: &model.User{
				Username: "dbfailuser",
				Email:    "dbfail@example.com",
				Password: "password123",
			},
			mockDB: func(u *model.User) *gorm.DB {
				return &gorm.DB{Error: errors.New("database connection error")}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDB{
				createFunc: tt.mockDB,
			}

			store := &UserStore{
				db: mock,
			}

			err := store.Create(tt.user)

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
		mockUsers     map[string]*model.User
		mockDBErr     error
		expectedUser  *model.User
		expectedError error
	}{
		{
			name:  "Successfully retrieve a user by email",
			email: "user@example.com",
			mockUsers: map[string]*model.User{
				"user@example.com": {Email: "user@example.com", Username: "testuser"},
			},
			expectedUser:  &model.User{Email: "user@example.com", Username: "testuser"},
			expectedError: nil,
		},
		{
			name:          "Attempt to retrieve a non-existent user",
			email:         "nonexistent@example.com",
			mockUsers:     map[string]*model.User{},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:          "Handle database connection error",
			email:         "user@example.com",
			mockUsers:     map[string]*model.User{},
			mockDBErr:     errors.New("database connection error"),
			expectedUser:  nil,
			expectedError: errors.New("database connection error"),
		},
		{
			name:  "Retrieve user with maximum length email",
			email: "a@" + string(make([]byte, 252)) + ".com",
			mockUsers: map[string]*model.User{
				"a@" + string(make([]byte, 252)) + ".com": {Email: "a@" + string(make([]byte, 252)) + ".com", Username: "maxuser"},
			},
			expectedUser:  &model.User{Email: "a@" + string(make([]byte, 252)) + ".com", Username: "maxuser"},
			expectedError: nil,
		},
		{
			name:          "Attempt retrieval with an empty email string",
			email:         "",
			mockUsers:     map[string]*model.User{},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:  "Handle case-insensitive email lookup",
			email: "user@example.com",
			mockUsers: map[string]*model.User{
				"User@Example.com": {Email: "User@Example.com", Username: "caseuser"},
			},
			expectedUser:  &model.User{Email: "User@Example.com", Username: "caseuser"},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &mockDB{
				users: tt.mockUsers,
				dbErr: tt.mockDBErr,
			}
			store := &UserStore{db: mockDB}

			user, err := store.GetByEmail(tt.email)

			if !mockDB.called {
				t.Error("Expected database to be queried, but it wasn't")
			}

			if (err != nil) != (tt.expectedError != nil) {
				t.Errorf("GetByEmail() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if err != nil && err.Error() != tt.expectedError.Error() {
				t.Errorf("GetByEmail() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if !compareUsers(user, tt.expectedUser) {
				t.Errorf("GetByEmail() = %v, expected %v", user, tt.expectedUser)
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
			username: string(make([]byte, 255)),
			mockDB: &mockDB{
				users: map[string]*model.User{
					string(make([]byte, 255)): {Username: string(make([]byte, 255)), Email: "maxuser@example.com"},
				},
			},
			want:    &model.User{Username: string(make([]byte, 255)), Email: "maxuser@example.com"},
			wantErr: nil,
		},
		{
			name:     "Handle case sensitivity in username lookup",
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

