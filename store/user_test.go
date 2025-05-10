package store

import (
	"errors"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
)


/*
ROOST_METHOD_HASH=Create_889fc0fc45
ROOST_METHOD_SIG_HASH=Create_4c48ec3920


 */
func TestCreate(t *testing.T) {

	tests := []struct {
		name      string
		user      model.User
		mockSetup func(sqlmock.Sqlmock)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "Successfully Create a New User",
			user: model.User{
				Username: "testuser",
				Email:    "testuser@example.com",
				Password: "password",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs(sqlmock.AnyArg(), "testuser", "testuser@example.com", "password", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Fail to Create User with Duplicate Email",
			user: model.User{
				Username: "testuser",
				Email:    "duplicate@example.com",
				Password: "password",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs(sqlmock.AnyArg(), "testuser", "duplicate@example.com", "password", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("duplicate email"))
				mock.ExpectRollback()
			},
			wantErr: true,
			errMsg:  "duplicate email",
		},
		{
			name: "Fail to Create User with Null Username",
			user: model.User{
				Username: "",
				Email:    "nullusername@example.com",
				Password: "password",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs(sqlmock.AnyArg(), "", "nullusername@example.com", "password", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("username cannot be null"))
				mock.ExpectRollback()
			},
			wantErr: true,
			errMsg:  "username cannot be null",
		},
		{
			name: "Database Connection Failure",
			user: model.User{
				Username: "testuser",
				Email:    "dbfailure@example.com",
				Password: "password",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("database connection error"))
			},
			wantErr: true,
			errMsg:  "database connection error",
		},
		{
			name: "Create User with Minimum Required Fields",
			user: model.User{
				Username: "minuser",
				Email:    "minuser@example.com",
				Password: "password",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs(sqlmock.AnyArg(), "minuser", "minuser@example.com", "password", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "Create User with Special Characters in Username",
			user: model.User{
				Username: "user!@#",
				Email:    "specialchar@example.com",
				Password: "password",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\"").WithArgs(sqlmock.AnyArg(), "user!@#", "specialchar@example.com", "password", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to open mock sql db: %v", err)
			}
			defer db.Close()

			gormDB, err := gorm.Open("postgres", db)
			if err != nil {
				t.Fatalf("failed to open gorm db: %v", err)
			}

			tt.mockSetup(mock)

			store := &UserStore{db: gormDB}

			err = store.Create(&tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %v", err)
			}
		})
	}
}

