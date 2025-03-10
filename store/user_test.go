package store

import (
	errors "errors"
	debug "runtime/debug"
	testing "testing"

	"github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

/*
ROOST_METHOD_HASH=UserStore_GetByEmail_fda09af5c4
ROOST_METHOD_SIG_HASH=UserStore_GetByEmail_9e84f3286b

FUNCTION_DEF=func (s *UserStore) GetByEmail(email string) (*model.User, error) // GetByEmail finds a user from email
*/
func TestUserStoreGetByEmail(t *testing.T) {
	type testCase struct {
		name          string
		email         string
		mockSetup     func(sqlmock.Sqlmock)
		expectedUser  *model.User
		expectedError error
	}

	testCases := []testCase{
		{
			name:  "Scenario 1: Successfully Retrieve User by Email",
			email: "test@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"email", "username"}).
					AddRow("test@example.com", "testuser")
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").WithArgs("test@example.com").WillReturnRows(rows)
			},
			expectedUser:  &model.User{Email: "test@example.com", Username: "testuser"},
			expectedError: nil,
		},
		{
			name:  "Scenario 2: User Not Found by Email",
			email: "notfound@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").WithArgs("notfound@example.com").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:  "Scenario 3: Database Error Occurs during Retrieval",
			email: "error@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").WithArgs("error@example.com").WillReturnError(errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("database error"),
		},
		{
			name:  "Scenario 4: Empty Email String Provided",
			email: "",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").WithArgs("").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:  "Scenario 5: Case Sensitivity in Email Lookup",
			email: "USER@Email.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"email", "username"}).
					AddRow("user@email.com", "testuser")
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").WithArgs("USER@Email.com").WillReturnRows(rows)
			},
			expectedUser:  &model.User{Email: "user@email.com", Username: "testuser"},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			tc.mockSetup(mock)

			gormDB, err := gorm.Open("_", db)
			if err != nil {
				t.Fatalf("An error '%s' was not expected when opening a gorm database connection", err)
			}

			userStore := &UserStore{db: gormDB}

			user, err := userStore.GetByEmail(tc.email)

			if tc.expectedError != nil {
				if err == nil || err.Error() != tc.expectedError.Error() {
					t.Errorf("Expected error: %v, got: %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
			}

			if tc.expectedUser != nil {
				if user == nil || !compareUsers(*user, *tc.expectedUser) {
					t.Errorf("Expected user: %+v, got: %+v", tc.expectedUser, user)
				}
			} else {
				if user != nil {
					t.Errorf("Expected no user, but got: %+v", user)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func compareUsers(user1, user2 model.User) bool {
	return user1.Email == user2.Email && user1.Username == user2.Username
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
		mockSetup     func(mock sqlmock.Sqlmock)
		expectedUser  *model.User
		expectedError error
	}{
		{
			name:     "Successfully Retrieve User by Username",
			username: "validuser",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"username", "email"}).
					AddRow("validuser", "user@example.com")
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1`).
					WithArgs("validuser").WillReturnRows(rows)
			},
			expectedUser:  &model.User{Username: "validuser", Email: "user@example.com"},
			expectedError: nil,
		},
		{
			name:     "User Not Found for Non-Existent Username",
			username: "nonexistent",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1`).
					WithArgs("nonexistent").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:     "Database Error Handling",
			username: "anyuser",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1`).
					WithArgs("anyuser").WillReturnError(errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("database error"),
		},
		{
			name:     "Case Sensitivity in Username Search",
			username: "VALIDUSER",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1`).
					WithArgs("VALIDUSER").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:     "Empty Username Input",
			username: "",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = \$1`).
					WithArgs("").WillReturnError(errors.New("invalid input"))
			},
			expectedUser:  nil,
			expectedError: errors.New("invalid input"),
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
			require.NoError(t, err)
			defer db.Close()

			tt.mockSetup(mock)

			gormDB, err := gorm.Open("postgres", db)
			require.NoError(t, err)

			userStore := &UserStore{db: gormDB}

			user, err := userStore.GetByUsername(tt.username)

			assert.Equal(t, tt.expectedUser, user, "Expected user does not match")
			assert.Equal(t, tt.expectedError, err, "Expected error does not match")

			err = mock.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}
