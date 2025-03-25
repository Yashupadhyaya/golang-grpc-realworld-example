package store

import (
	testing "testing"
	gosqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	errors "errors"
)





type UserStoreInterface interface {
	Create(user *model.User) error
}
type UserStoreMock struct {
	Db *gorm.DB
}


/*
ROOST_METHOD_HASH=UserStore_Create_9495ddb29d
ROOST_METHOD_SIG_HASH=UserStore_Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user


*/
func TestUserStoreCreate(t *testing.T) {
	testCases := []struct {
		testName      string
		user          *model.User
		mockFunc      func(mock sqlmock.Sqlmock)
		expectedError bool
		errorMessage  string
	}{
		{
			testName: "Successful user creation",
			user: &model.User{

				Username: "Test User",
				Email:    "testuser@example.com",
				Password: "password",
			},
			mockFunc: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "email", "password"}).
					AddRow(1, "Test User", "testuser@example.com", "password")
				mock.ExpectQuery("INSERT INTO \"users\" (.*)").
					WithArgs("Test User", "testuser@example.com", "password").WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			testName: "Unsuccessful user creation due to database error",
			user: &model.User{
				Username: "Test User",
				Email:    "testuser@example.com",
				Password: "password",
			},
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO \"users\" (.*)").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedError: true,
			errorMessage:  gorm.ErrRecordNotFound.Error(),
		},
		{
			testName: "Creating a user with invalid data",
			user: &model.User{
				Username: "Test User",
				Email:    "testuser",
				Password: "password",
			},
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO \"users\" (.*)").WillReturnError(gorm.ErrInvalidSQL)
			},
			expectedError: true,
			errorMessage:  gorm.ErrInvalidSQL.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic encountered during test execution: %v", r)
				}
			}()

			db, mock, _ := sqlmock.New()
			defer db.Close()

			testCase.mockFunc(mock)

			gormDB, _ := gorm.Open("postgres", db)

			userStore := UserStoreMock{
				Db: gormDB,
			}

			err := userStore.Create(testCase.user)

			if testCase.expectedError {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if err.Error() != testCase.errorMessage {
					t.Fatalf("expected error: %v, got: %v", testCase.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("did not expect error but got: %v", err.Error())
				}
			}
		})
	}
}

func (usm UserStoreMock) Create(m *model.User) error {
	return usm.Db.Create(m).Error
}


/*
ROOST_METHOD_HASH=UserStore_GetByUsername_622b1b9e41
ROOST_METHOD_SIG_HASH=UserStore_GetByUsername_992f00baec

FUNCTION_DEF=func (s *UserStore) GetByUsername(username string) (*model.User, error) // GetByUsername finds a user from username


*/
func TestUserStoreGetByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	gormDB, err := gorm.Open("postgres", db)
	if err != nil {
		t.Fatalf("an error '%s' was expected when opening gorm database", err)
	}

	defer db.Close()

	userStore := &UserStore{db: gormDB}

	tests := []struct {
		name        string
		username    string
		mockUser    *model.User
		setupMocks  func()
		expectedErr error
	}{
		{
			name:     "User Found",
			username: "JohnDoe",
			mockUser: &model.User{Username: "JohnDoe"},
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"username"}).AddRow("JohnDoe")
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE").WithArgs("JohnDoe").WillReturnRows(rows)
			},
			expectedErr: nil,
		},
		{
			name:     "User Not Found",
			username: "testuser",
			setupMocks: func() {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE").WithArgs("testuser").WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedErr: gorm.ErrRecordNotFound,
		},
		{
			name:     "Database Error",
			username: "testuser",
			setupMocks: func() {
				mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE").WithArgs("testuser").WillReturnError(errors.New("database error"))
			},
			expectedErr: errors.New("database error"),
		},
		{
			name:        "Empty Username",
			username:    "",
			expectedErr: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			if tt.expectedErr != tt.setupMocks() {
				t.Errorf("Expected and actual errors do not match. Expected: %v | Actual: %v", tt.expectedErr.Error(), tt.setupMocks())
			}

			user, err := userStore.GetByUsername(tt.username)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if tt.mockUser != nil && user.Username != tt.mockUser.Username {
				t.Errorf("Expected Username: %v | Found Username: %v", tt.mockUser.Username, user.Username)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations")
			}
		})
	}
}

