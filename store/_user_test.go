package store

import (
	fmt "fmt"
	net "net"
	os "os"
	testing "testing"
	gosqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	require "github.com/stretchr/testify/require"
)





type mockDB struct {
	Create func(user *model.User) error
}


/*
ROOST_METHOD_HASH=UserStore_Create_9495ddb29d
ROOST_METHOD_SIG_HASH=UserStore_Create_18451817fe

FUNCTION_DEF=func (s *UserStore) Create(m *model.User) error // Create create a user


*/
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
				mock.ExpectExec("INSERT INTO `users` (.+) VALUES (.+)").
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

