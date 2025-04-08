package store

import (
	errors "errors"
	fmt "fmt"
	os "os"
	debug "runtime/debug"
	strings "strings"
	testing "testing"
	gosqlmock "github.com/DATA-DOG/go-sqlmock"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	gorm "github.com/jinzhu/gorm"
)








/*
ROOST_METHOD_HASH=UserStore_Create_9495ddb29d
ROOST_METHOD_SIG_HASH=UserStore_Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user


*/
func TestUserStoreCreate(t *testing.T) {

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {

		os.Stdout = oldStdout
	}()

	mockUser := &model.User{
		Username: "testUser",
		Email:    "testEmail@test.com",
		Password: "testPassword",
	}

	testScenarios := []struct {
		name          string
		input         *model.User
		mockDBFunc    func() (*gorm.DB, sqlmock.Sqlmock)
		expectedError string
	}{
		{
			name:  "Normal operation - Create a valid User",
			input: mockUser,
			mockDBFunc: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").
					WithArgs(mockUser.Username, mockUser.Email, mockUser.Password).
					WillReturnResult(sqlmock.NewResult(1, 1))

				gormDB, _ := gorm.Open("postgres", db)
				return gormDB, mock
			},
			expectedError: "",
		},
		{
			name:  "Create a User with missing fields",
			input: &model.User{},
			mockDBFunc: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").
					WithArgs("").
					WillReturnError(errors.New("Required fields are empty"))

				gormDB, _ := gorm.Open("postgres", db)
				return gormDB, mock
			},
			expectedError: "Required fields are empty",
		},
		{
			name:  "Error while connecting to the database",
			input: mockUser,
			mockDBFunc: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `users`").
					WithArgs(mockUser.Username, mockUser.Email, mockUser.Password).
					WillReturnError(errors.New("Failed to connect to database"))

				gormDB, _ := gorm.Open("postgres", db)
				return gormDB, mock
			},
			expectedError: "Failed to connect to database",
		},
	}

	for _, tt := range testScenarios {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic recovered: %v\n", r)
					t.Logf("Stack: \n%s", string(debug.Stack()))
					t.Fail()
				}
			}()

			db, _ := tt.mockDBFunc()

			defer db.Close()
			mockUserStore := &UserStore{db: db}

			err := mockUserStore.Create(tt.input)

			if tt.expectedError != "" {
				if strings.Compare(err.Error(), tt.expectedError) != 0 {
					t.Errorf("Expected error: %v, got: %v", tt.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

