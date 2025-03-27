// ********RoostGPT********
/*

roost_feedback [3/27/2025, 2:36:24 PM]:Use these import statements:\r\n```\r\nimport (\r\n\terrors errors\r\n\truntime/debug\r\n\ttesting testing\r\n\r\n\tsqlmock github.com/DATA-DOG/go-sqlmock\r\n\tgorm github.com/jinzhu/gorm\r\n\tmodel github.com/raahii/golang-grpc-realworld-example/model\r\n\trequire github.com/stretchr/testify/require\r\n)\r\n```\r\n\r\n\r\nThe testing table in the TestUserStoreCreate function must be this:\r\n```\r\ntt := []struct {\r\n\t\tname     string\r\n\t\tuser     *model.User\r\n\t\tmock     func() (sqlmock.Sqlmock, error)\r\n\t\twantErr  bool\r\n\t\texpected *model.User\r\n\t}{\r\n\r\n\t\t{\r\n\t\t\tname: Scenario 1: Normal operation - Create a valid User,\r\n\t\t\tuser: &model.User{\r\n\t\t\t\tUsername: Alice,\r\n\t\t\t},\r\n\t\t\tmock: func() (sqlmock.Sqlmock, error) {\r\n\t\t\t\t_, mock, err := sqlmock.New()\r\n\t\t\t\tif err != nil {\r\n\t\t\t\t\treturn nil, err\r\n\t\t\t\t}\r\n\t\t\t\tmock.ExpectBegin()\r\n\t\t\t\tmock.ExpectExec(INSERT INTO `users` (.+) VALUES (.+)).\r\n\t\t\t\t\tWithArgs(Any, AnyWhere, Any)\r\n\t\t\t\tmock.ExpectCommit()\r\n\t\t\t\treturn mock, err\r\n\t\t\t},\r\n\t\t\twantErr:  false,\r\n\t\t\texpected: nil,\r\n\t\t},\r\n\t}\r\n```\r\n\r\nVERY IMPORTANT NOTE: DO NOT make any other change in the entire test code!!!
*/

// ********RoostGPT********

package store

import (
	"errors"
	"runtime/debug"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/require"
)

type mockDB struct {
	Create func(user *model.User) error
}

func TestUserStoreCreate(t *testing.T) {
	tt := []struct {
		name     string
		user     *model.User
		mock     func() (sqlmock.Sqlmock, error)
		wantErr  bool
		expected *model.User
	}{
		{
			name: "Scenario 1: Normal operation - Create a valid User",
			user: &model.User{
				Username: "Alice",
			},
			mock: func() (sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, err
				}
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\" (.+) VALUES (.+)").
					WithArgs(Any, AnyWhere, Any)
				mock.ExpectCommit()
				return mock, err
			},
			wantErr:  false,
			expected: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v\n%s", r, debug.Stack())
					t.Fail()
				}
			}()

			mock, err := tc.mock()
			if err != nil {
				t.Error(err)
			}
			db, err := gorm.Open("mysql", mock)
			userStore := UserStore{db: db}
			err = userStore.Create(tc.user)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func (mdb *mockDB) Create(user *model.User) error {
	return mdb.Create(user)
}

func TestUserStoreGetByEmail(t *testing.T) {
	testCases := []struct {
		name          string
		email         string
		expectedUser  *model.User
		expectedError error
	}{
		{
			name:  "Valid Email Exist",
			email: "test@example.com",
			expectedUser: &model.User{
				Email: "test@example.com",
			},
			expectedError: nil,
		},
		{
			name:          "Email Does Not Exist",
			email:         "unknown@example.com",
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:          "Empty Email String",
			email:         "",
			expectedUser:  nil,
			expectedError: errors.New("invalid email"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered. Fail test. %v", r)
					t.Fail()
				}
			}()

			db, mock, _ := sqlmock.New()
			gormDB, _ := gorm.Open("postgres", db)
			userStore := &UserStore{db: gormDB}

			user, error := userStore.GetByEmail(tc.email)
			if error != tc.expectedError {
				t.Errorf("Unexpected error. Want '%v'. Got '%v'", tc.expectedError, error)
				return
			}

			if tc.expectedUser != nil && user.Email != tc.expectedUser.Email {
				t.Errorf("Unexpected user returned. Want '%v'. Got '%v'", tc.expectedUser, user)
			}

			_ = mock.ExpectationsWereMet()
		})
	}
}

func TestUserStoreGetByUsername(t *testing.T) {
	scenarios := []struct {
		desc     string
		username string
		mock     func(mock sqlmock.Sqlmock)
		wantErr  bool
	}{
		{
			desc:     "Successful retrieval of user by username",
			username: "testUser",
			mock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"username"}).
					AddRow("testUser")
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (username = (.+))$").WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			desc:     "User retrieval with non-existing username",
			username: "nonExistentUser",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (username = (.+))$").WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: true,
		},
		{
			desc:     "User retrieval with empty username",
			username: "",
			mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (username = (.+))$").WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr: true,
		},
	}

	for _, s := range scenarios {
		t.Run(s.desc, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Panic encountered so failing test. %v", r)
				}
			}()

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
			}

			defer db.Close()
			gDB, _ := gorm.Open("mysql", db)

			s.mock(mock)

			userStore := &UserStore{db: gDB}
			_, err = userStore.GetByUsername(s.username)
			if s.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else {
					t.Logf("expected error occurred: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("no error expected, but got: %v", err)
				} else {
					t.Logf("no error expected and got none")
				}
			}
		})
	}
}
