package store

import (
	errors "errors"
	fmt "fmt"
	testing "testing"
	gosqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	debug "runtime/debug"
)








/*
ROOST_METHOD_HASH=UserStore_GetByEmail_fda09af5c4
ROOST_METHOD_SIG_HASH=UserStore_GetByEmail_9e84f3286b

FUNCTION_DEF=func (s *UserStore) GetByEmail(email string) (*model.User, error) // GetByEmail finds a user from email


*/
func TestUserStoreGetByEmail(t *testing.T) {

	tests := []struct {
		name         string
		email        string
		mockSetup    func(sqlmock.Sqlmock)
		expectedUser *model.User
		expectedErr  error
	}{
		{
			name:  "Retrieve Existing User by Email",
			email: "test@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "name"}).
					AddRow(1, "test@example.com", "Test User")
				mock.ExpectQuery("SELECT * FROM \"users\" WHERE (email = ?)").
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expectedUser: &model.User{ID: 1, Email: "test@example.com", Name: "Test User"},
			expectedErr:  nil,
		},
		{
			name:  "Handle Non-Existent User Email",
			email: "nonexistent@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM \"users\" WHERE (email = ?)").
					WithArgs("nonexistent@example.com").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser: nil,
			expectedErr:  gorm.ErrRecordNotFound,
		},
		{
			name:  "Database Connection Error",
			email: "anyemail@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM \"users\" WHERE (email = ?)").
					WithArgs("anyemail@example.com").
					WillReturnError(errors.New("connection refused"))
			},
			expectedUser: nil,
			expectedErr:  errors.New("connection refused"),
		},
		{
			name:  "Handle Invalid Email Format",
			email: "invalid-email",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM \"users\" WHERE (email = ?)").
					WithArgs("invalid-email").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser: nil,
			expectedErr:  gorm.ErrRecordNotFound,
		},
		{
			name:  "Query with Email Case Sensitivity",
			email: "TEST@EXAMPLE.COM",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "name"}).
					AddRow(2, "test@example.com", "Test User")
				mock.ExpectQuery("SELECT * FROM \"users\" WHERE (email = ?)").
					WithArgs("TEST@EXAMPLE.COM").
					WillReturnRows(rows)
			},
			expectedUser: &model.User{ID: 2, Email: "test@example.com", Name: "Test User"},
			expectedErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tt.mockSetup(mock)

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("An error '%s' was not expected when opening a gorm database connection", err)
			}
			defer gormDB.Close()

			userStore := &UserStore{db: gormDB}

			user, err := userStore.GetByEmail(tt.email)

			if tt.expectedErr != nil && err == nil {
				t.Errorf("Expected error but got none")
			}
			if tt.expectedErr == nil && err != nil {
				t.Errorf("Did not expect error but got %v", err)
			}
			if tt.expectedErr != nil && err != nil && tt.expectedErr.Error() != err.Error() {
				t.Errorf("Expected error %v but got %v", tt.expectedErr, err)
			}

			if tt.expectedUser != nil && user != nil {
				if user.ID != tt.expectedUser.ID || user.Email != tt.expectedUser.Email || user.Name != tt.expectedUser.Name {
					t.Errorf("Expected user %v but got %v", tt.expectedUser, user)
				}
			} else if tt.expectedUser == nil && user != nil {
				t.Errorf("Expected no user but got %v", user)
			} else if tt.expectedUser != nil && user == nil {
				t.Errorf("Expected user %v but got none", tt.expectedUser)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %s", err)
			}

			t.Logf("Test '%s' passed", tt.name)
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_GetByUsername_622b1b9e41
ROOST_METHOD_SIG_HASH=UserStore_GetByUsername_992f00baec

FUNCTION_DEF=func (s *UserStore) GetByUsername(username string) (*model.User, error) // GetByUsername finds a user from username


*/
func TestUserStoreGetByUsername(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open a stub database connection: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("Failed to open gorm DB: %v", err)
	}
	defer gormDB.Close()

	userStore := &UserStore{db: gormDB}

	tests := []struct {
		name          string
		username      string
		mockSetup     func()
		expectedUser  *model.User
		expectedError error
	}{
		{
			name:     "Scenario 1: Retrieve Existing User by Username",
			username: "john_doe",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"username", "email"}).
					AddRow("john_doe", "john@example.com")
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").WillReturnRows(rows)
			},
			expectedUser:  &model.User{Username: "john_doe", Email: "john@example.com"},
			expectedError: nil,
		},
		{
			name:     "Scenario 2: User Not Found by Username",
			username: "non_existent_user",
			mockSetup: func() {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:     "Scenario 3: Database Error Occurrence",
			username: "any_user",
			mockSetup: func() {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").WillReturnError(errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("database error"),
		},
		{
			name:     "Scenario 4: Empty Username Input",
			username: "",
			mockSetup: func() {
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:     "Scenario 5: Special Characters in Username",
			username: "user!@#",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"username", "email"}).
					AddRow("user!@#", "special@example.com")
				mock.ExpectQuery("^SELECT (.+) FROM `users` WHERE (.+)$").WillReturnRows(rows)
			},
			expectedUser:  &model.User{Username: "user!@#", Email: "special@example.com"},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			tt.mockSetup()

			user, err := userStore.GetByUsername(tt.username)

			if tt.expectedError != nil {
				if err == nil || err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
				}
			} else {
				if user == nil || user.Username != tt.expectedUser.Username || user.Email != tt.expectedUser.Email {
					t.Errorf("expected user: %v, got: %v", tt.expectedUser, user)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

