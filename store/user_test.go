package store

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
)

/*
ROOST_METHOD_HASH=UserStore_Create_9495ddb29d
ROOST_METHOD_SIG_HASH=UserStore_Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user
*/
func TestUserStoreCreate(t *testing.T) {
	tests := []struct {
		name    string
		user    *model.User
		dbMock  func() (*gorm.DB, sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "Successful User Creation",
			dbMock: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gdb, _ := gorm.Open("mysql", db)
				mock.ExpectExec("INSERT INTO").WillReturnResult(sqlmock.NewResult(1, 1))

				return gdb, mock
			},
			user:    &model.User{Username: "testuser", Email: "test@test.com"},
			wantErr: false,
		},
		{
			name: "Failure User Creation due to Invalid Input",
			dbMock: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gdb, _ := gorm.Open("mysql", db)
				mock.ExpectExec("INSERT INTO").WillReturnError(gorm.ErrRecordNotFound)

				return gdb, mock
			},
			user:    &model.User{Email: "test@test.com"},
			wantErr: true,
		},
		{
			name: "Failure due to Database Connection Error",
			dbMock: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gdb, _ := gorm.Open("mysql", db)
				mock.ExpectExec("INSERT INTO").WillReturnError(gorm.ErrRecordNotFound)

				return gdb, mock
			},
			user:    &model.User{Username: "testuser", Email: "test@test.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic encountered so failing test. %v", r)
					t.Fail()
				}
			}()

			gdb, _ := tt.dbMock()
			us := &UserStore{db: gdb}

			err := us.Create(tt.user)

			if (err != nil) != tt.wantErr {
				t.Errorf("UserStore.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
