package store

import (
	errors "errors"
	fmt "fmt"
	testing "testing"
	gosqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	assert "github.com/stretchr/testify/assert"
	os "os"
	bytes "bytes"
	debug "runtime/debug"
)








/*
ROOST_METHOD_HASH=UserStore_GetByEmail_fda09af5c4
ROOST_METHOD_SIG_HASH=UserStore_GetByEmail_9e84f3286b

FUNCTION_DEF=func (s *UserStore) GetByEmail(email string) (*model.User, error) // GetByEmail finds a user from email


*/
func TestUserStoreGetByEmail(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error initializing sqlmock: %v", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("Error initializing gorm: %v", err)
	}
	defer gormDB.Close()

	userStore := &UserStore{db: gormDB}

	testCases := []struct {
		name          string
		email         string
		setupMock     func()
		expectedUser  *model.User
		expectedError error
	}{
		{
			name:  "Retrieve Existing User by Email",
			email: "test@example.com",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "email", "username"}).
					AddRow(1, "test@example.com", "testuser")
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expectedUser: &model.User{ID: 1, Email: "test@example.com", Username: "testuser"},
		},
		{
			name:  "Handle Non-Existent Email",
			email: "nonexistent@example.com",
			setupMock: func() {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs("nonexistent@example.com").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:  "Handle Database Error",
			email: "error@example.com",
			setupMock: func() {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs("error@example.com").
					WillReturnError(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
		{
			name:  "Case Sensitivity in Email Retrieval",
			email: "Test@Example.Com",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "email", "username"}).
					AddRow(1, "test@example.com", "testuser")
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs("Test@Example.Com").
					WillReturnRows(rows)
			},
			expectedUser: &model.User{ID: 1, Email: "test@example.com", Username: "testuser"},
		},
		{
			name:  "Validate Email Format Handling",
			email: "invalid-email",
			setupMock: func() {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+)$").
					WithArgs("invalid-email").
					WillReturnError(errors.New("invalid email format"))
			},
			expectedError: errors.New("invalid email format"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			tc.setupMock()

			user, err := userStore.GetByEmail(tc.email)

			if tc.expectedError != nil {
				assert.Error(t, err, fmt.Sprintf("Expected error but got none for case: %s", tc.name))
				assert.Nil(t, user, fmt.Sprintf("Expected no user but got one for case: %s", tc.name))
			} else {
				assert.NoError(t, err, fmt.Sprintf("Expected no error but got one for case: %s", tc.name))
				assert.NotNil(t, user, fmt.Sprintf("Expected user but got none for case: %s", tc.name))
				assert.Equal(t, tc.expectedUser, user, fmt.Sprintf("Expected user mismatch for case: %s", tc.name))
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unmet expectations: %s", err)
			}
		})
	}
}


/*
ROOST_METHOD_HASH=UserStore_GetByUsername_622b1b9e41
ROOST_METHOD_SIG_HASH=UserStore_GetByUsername_992f00baec

FUNCTION_DEF=func (s *UserStore) GetByUsername(username string) (*model.User, error) // GetByUsername finds a user from username


*/
func TestUserStoreGetByUsername(t *testing.T) {

	tests := []struct {
		name         string
		username     string
		setupMock    func(sqlmock.Sqlmock)
		expectedUser *model.User
		expectError  bool
	}{
		{
			name:     "Retrieve Existing User by Username",
			username: "existing_user",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).
					WithArgs("existing_user").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(1, "existing_user"))
			},
			expectedUser: &model.User{ID: 1, Username: "existing_user"},
			expectError:  false,
		},
		{
			name:     "Handle Non-Existent Username Gracefully",
			username: "nonexistent_user",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).
					WithArgs("nonexistent_user").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser: nil,
			expectError:  true,
		},
		{
			name:     "Database Connection Error Handling",
			username: "any_user",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).
					WithArgs("any_user").
					WillReturnError(fmt.Errorf("connection error"))
			},
			expectedUser: nil,
			expectError:  true,
		},
		{
			name:     "Case Sensitivity in Username Search",
			username: "Existing_User",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).
					WithArgs("Existing_User").
					WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedUser: nil,
			expectError:  true,
		},
		{
			name:     "Handling Special Characters in Username",
			username: "user@123!",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT \* FROM "users" WHERE \(username = \?\) ORDER BY "users"\."id" ASC LIMIT 1`).
					WithArgs("user@123!").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username"}).AddRow(2, "user@123!"))
			},
			expectedUser: &model.User{ID: 2, Username: "user@123!"},
			expectError:  false,
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
				t.Fatalf("failed to open sqlmock database: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("failed to open gorm DB: %v", err)
			}
			defer gormDB.Close()

			tt.setupMock(mock)

			store := &UserStore{db: gormDB}

			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			user, err := store.GetByUsername(tt.username)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			fmt.Fscanf(r, "%s", &buf)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("did not expect error but got: %v", err)
			}
			if tt.expectError && user != nil {
				t.Errorf("expected nil user but got: %v", user)
			}
			if !tt.expectError && user == nil {
				t.Errorf("expected user but got nil")
			}
			if !tt.expectError && user != nil && *user != *tt.expectedUser {
				t.Errorf("expected user %v but got %v", *tt.expectedUser, *user)
			}
		})
	}
}

