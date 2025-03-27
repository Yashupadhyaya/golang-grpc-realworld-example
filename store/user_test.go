package store

import (
	debug "runtime/debug"
	testing "testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	gorm "github.com/jinzhu/gorm"
	model "github.com/raahii/golang-grpc-realworld-example/model"
	require "github.com/stretchr/testify/require"
)

type mockDB struct {
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
				_, mock, err := sqlmock.New()
				if err != nil {
					return nil, err
				}
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO \"users\" (.+) VALUES (.+)").
					WithArgs("Any", "AnyWhere", "Any")
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
			require.NoError(t, err)
			db, err := gorm.Open("mysql", mock)
			require.NoError(t, err)
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
