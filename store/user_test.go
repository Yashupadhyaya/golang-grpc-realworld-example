package store

import (
	sql "database/sql"
	errors "errors"
	reflect "reflect"
	debug "runtime/debug"
	strings "strings"
	testing "testing"
	time "time"

	"github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	assert "github.com/stretchr/testify/assert"
)

type testCase struct {
	name          string
	email         string
	expectedUser  *model.User
	expectedError error
	mockBehavior  func()
}

func TestUserStoreGetByEmail(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mockDB, err := gorm.Open("sqlmock", db)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a mock gorm database", err)
	}
	defer mockDB.Close()

	userStore := &UserStore{db: mockDB}

	type testCase struct {
		name          string
		email         string
		expectedUser  *model.User
		expectedError error
		mockBehavior  func()
	}

	testCases := []testCase{
		{
			name:  "Normal Operation - Existing Email",
			email: "test@example.com",
			expectedUser: &model.User{
				Email: "test@example.com",
			},
			expectedError: nil,
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "test@example.com")
				mock.ExpectQuery("SELECT").WithArgs("test@example.com").WillReturnRows(rows)
			},
		},
		{
			name:  "Normal Operation - Multiple Users with Same Email",
			email: "duplicate@example.com",
			expectedUser: &model.User{
				Email: "duplicate@example.com",
			},
			expectedError: nil,
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "duplicate@example.com").AddRow(2, "duplicate@example.com")
				mock.ExpectQuery("SELECT").WithArgs("duplicate@example.com").WillReturnRows(rows)
			},
		},
		{
			name:          "Error Handling - Non-Existing Email",
			email:         "nonexistent@example.com",
			expectedUser:  nil,
			expectedError: errors.New("record not found"),
			mockBehavior: func() {
				mock.ExpectQuery("SELECT").WithArgs("nonexistent@example.com").WillReturnError(errors.New("record not found"))
			},
		},
		{
			name:          "Error Handling - Empty Email",
			email:         "",
			expectedUser:  nil,
			expectedError: errors.New("invalid email"),
			mockBehavior: func() {

			},
		},
		{
			name:          "Error Handling - SQL Injection Attempt",
			email:         "' OR '1'='1",
			expectedUser:  nil,
			expectedError: errors.New("invalid email"),
			mockBehavior: func() {

			},
		},
		{
			name:  "Performance - Large Number of Users",
			email: "large@example.com",
			expectedUser: &model.User{
				Email: "large@example.com",
			},
			expectedError: nil,
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "large@example.com")
				mock.ExpectQuery("SELECT").WithArgs("large@example.com").WillReturnRows(rows)
			},
		},
		{
			name:  "Edge Case - Email with Special Characters",
			email: "test.special@email.com",
			expectedUser: &model.User{
				Email: "test.special@email.com",
			},
			expectedError: nil,
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "test.special@email.com")
				mock.ExpectQuery("SELECT").WithArgs("test.special@email.com").WillReturnRows(rows)
			},
		},
		{
			name:  "Edge Case - Email with Mixed Case",
			email: "MiXeDcAsE@eXaMpLe.CoM",
			expectedUser: &model.User{
				Email: "mixedcase@example.com",
			},
			expectedError: nil,
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "mixedcase@example.com")
				mock.ExpectQuery("SELECT").WithArgs("MiXeDcAsE@eXaMpLe.CoM").WillReturnRows(rows)
			},
		},
		{
			name:  "Edge Case - Email with Leading and Trailing Spaces",
			email: "  test@example.com  ",
			expectedUser: &model.User{
				Email: "test@example.com",
			},
			expectedError: nil,
			mockBehavior: func() {
				rows := sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "test@example.com")
				mock.ExpectQuery("SELECT").WithArgs("test@example.com").WillReturnRows(rows)
			},
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

			tc.mockBehavior()

			email := strings.TrimSpace(tc.email)
			start := time.Now()
			user, err := userStore.GetByEmail(email)
			elapsed := time.Since(start)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedUser, user)

			if elapsed > 100*time.Millisecond {
				t.Logf("Test %s took %s, which is above the performance threshold", tc.name, elapsed)
			}
		})
	}
}

func TestUserStoreGetByUsername(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open("sqlite3", db)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub gorm database connection", err)
	}
	defer gormDB.Close()

	userStore := &UserStore{db: gormDB}

	tests := []testCase{
		{
			name:          "existingUser",
			email:         "existingUser",
			expectedUser:  &model.User{Model: gorm.Model{ID: 1}, Username: "existingUser", Email: "user@example.com"},
			expectedError: nil,
			mockBehavior: func() {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \\(username = \\?\\) ORDER BY \"users\"\\.\"id\" ASC LIMIT 1").
					WithArgs("existingUser").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email"}).AddRow(1, "existingUser", "user@example.com"))
			},
		},
		{
			name:          "nonExistentUser",
			email:         "nonExistentUser",
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
			mockBehavior: func() {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \\(username = \\?\\) ORDER BY \"users\"\\.\"id\" ASC LIMIT 1").
					WithArgs("nonExistentUser").
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:          "emptyUsername",
			email:         "",
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
			mockBehavior: func() {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \\(username = \\?\\) ORDER BY \"users\"\\.\"id\" ASC LIMIT 1").
					WithArgs("").
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:          "user@123",
			email:         "user@123",
			expectedUser:  &model.User{Model: gorm.Model{ID: 2}, Username: "user@123", Email: "user@123@example.com"},
			expectedError: nil,
			mockBehavior: func() {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \\(username = \\?\\) ORDER BY \"users\"\\.\"id\" ASC LIMIT 1").
					WithArgs("user@123").
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email"}).AddRow(2, "user@123", "user@123@example.com"))
			},
		},
		{
			name:          "ExIstInGUsEr",
			email:         "ExIstInGUsEr",
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
			mockBehavior: func() {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \\(username = \\?\\) ORDER BY \"users\"\\.\"id\" ASC LIMIT 1").
					WithArgs("ExIstInGUsEr").
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:          "SQL Injection Attempt",
			email:         "' OR '1'='1",
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
			mockBehavior: func() {
				mock.ExpectQuery("SELECT \\* FROM \"users\" WHERE \\(username = \\?\\) ORDER BY \"users\"\\.\"id\" ASC LIMIT 1").
					WithArgs("' OR '1'='1").
					WillReturnError(sql.ErrNoRows)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, string(debug.Stack()))
					t.Fail()
				}
			}()

			test.mockBehavior()

			user, err := userStore.GetByUsername(test.email)

			if test.expectedError != nil && err == nil {
				t.Fatalf("expected error %v, got %v", test.expectedError, err)
			}
			if !errors.Is(err, test.expectedError) {
				t.Fatalf("expected error %v, got %v", test.expectedError, err)
			}
			if !reflect.DeepEqual(user, test.expectedUser) {
				t.Fatalf("expected user %v, got %v", test.expectedUser, user)
			}
			t.Logf("Test passed for username: %s", test.name)
		})
	}
}
