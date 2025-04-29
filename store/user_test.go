package store

import (
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"errors"
)





type ExpectedBegin struct {
	commonExpectation
	delay time.Duration
}
type ExpectedCommit struct {
	commonExpectation
}
type ExpectedExec struct {
	queryBasedExpectation
	result driver.Result
	delay  time.Duration
}
type ExpectedRollback struct {
	commonExpectation
}
type T struct {
	common
	isEnvSet bool
	context  *testContext
}


/*
ROOST_METHOD_HASH=Create_889fc0fc45
ROOST_METHOD_SIG_HASH=Create_4c48ec3920


 */
func TestUserStoreCreate(t *testing.T) {

	tests := []struct {
		name        string
		user        *model.User
		setupMock   func(sqlmock.Sqlmock)
		expectedErr error
	}{
		{
			name: "Creating a Valid User",
			user: &model.User{
				Username: "validuser",
				Email:    "validuser@example.com",
				Password: "password123",
				Bio:      "A valid user",
				Image:    "validuser.png",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs(sqlmock.AnyArg(), "validuser", "validuser@example.com", "password123", "A valid user", "validuser.png").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: nil,
		},
		{
			name: "Creating a User with Duplicate Username",
			user: &model.User{
				Username: "duplicateuser",
				Email:    "uniqueemail@example.com",
				Password: "password123",
				Bio:      "A duplicate user",
				Image:    "duplicateuser.png",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs(sqlmock.AnyArg(), "duplicateuser", "uniqueemail@example.com", "password123", "A duplicate user", "duplicateuser.png").WillReturnError(errors.New("duplicate key value violates unique constraint"))
				mock.ExpectRollback()
			},
			expectedErr: errors.New("duplicate key value violates unique constraint"),
		},
		{
			name: "Creating a User with Duplicate Email",
			user: &model.User{
				Username: "uniqueuser",
				Email:    "duplicateemail@example.com",
				Password: "password123",
				Bio:      "A user with duplicate email",
				Image:    "uniqueuser.png",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs(sqlmock.AnyArg(), "uniqueuser", "duplicateemail@example.com", "password123", "A user with duplicate email", "uniqueuser.png").WillReturnError(errors.New("duplicate key value violates unique constraint"))
				mock.ExpectRollback()
			},
			expectedErr: errors.New("duplicate key value violates unique constraint"),
		},
		{
			name: "Creating a User with a Null Required Field",
			user: &model.User{
				Email:    "nullusername@example.com",
				Password: "password123",
				Bio:      "A user with null username",
				Image:    "nullusername.png",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs(sqlmock.AnyArg(), nil, "nullusername@example.com", "password123", "A user with null username", "nullusername.png").WillReturnError(errors.New("invalid input syntax for type"))
				mock.ExpectRollback()
			},
			expectedErr: errors.New("invalid input syntax for type"),
		},
		{
			name: "Creating a User with a Valid Foreign Key Relationship",
			user: &model.User{
				Username: "follower",
				Email:    "follower@example.com",
				Password: "password123",
				Bio:      "A follower user",
				Image:    "follower.png",
				Follows: []model.User{
					{
						Username: "followee",
						Email:    "followee@example.com",
						Password: "password123",
						Bio:      "A followee user",
						Image:    "followee.png",
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs(sqlmock.AnyArg(), "follower", "follower@example.com", "password123", "A follower user", "follower.png").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO \"follows\"").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: nil,
		},
		{
			name: "Handling Database Connection Failure",
			user: &model.User{
				Username: "dbfailuser",
				Email:    "dbfailuser@example.com",
				Password: "password123",
				Bio:      "A db fail user",
				Image:    "dbfailuser.png",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("invalid transaction"))
			},
			expectedErr: errors.New("invalid transaction"),
		},
		{
			name: "Creating a User with Optional Fields",
			user: &model.User{
				Username: "optionaluser",
				Email:    "optionaluser@example.com",
				Password: "password123",
				Bio:      "A user with optional fields",
				Image:    "optionaluser.png",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs(sqlmock.AnyArg(), "optionaluser", "optionaluser@example.com", "password123", "A user with optional fields", "optionaluser.png").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectedErr: nil,
		},
		{
			name: "Creating a User with an Invalid Data Type",
			user: &model.User{
				Username: "invaliduser",
				Email:    "invaliduser@example.com",
				Password: "password123",
				Bio:      "A user with invalid data type",
				Image:    "invaliduser.png",
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs(sqlmock.AnyArg(), "invaliduser", "invaliduser@example.com", "password123", "A user with invalid data type", "invaliduser.png").WillReturnError(errors.New("invalid input syntax for type"))
				mock.ExpectRollback()
			},
			expectedErr: errors.New("invalid input syntax for type"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
			}
			store := &UserStore{db: gormDB}

			tt.setupMock(mock)

			err = store.Create(tt.user)

			assert.Equal(t, tt.expectedErr, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

