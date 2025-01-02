package store

import (
	"errors"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


type MockDB struct {
	mock.Mock
}





type Call struct {
	Parent *Mock

	// The name of the method that was or will be called.
	Method string

	// Holds the arguments of the method.
	Arguments Arguments

	// Holds the arguments that should be returned when
	// this method is called.
	ReturnArguments Arguments

	// Holds the caller info for the On() call
	callerInfo []string

	// The number of times to return the return arguments when setting
	// expectations. 0 means to always return the value.
	Repeatability int

	// Amount of times this call has been called
	totalCalls int

	// Call to this method can be optional
	optional bool

	// Holds a channel that will be used to block the Return until it either
	// receives a message or is closed. nil means it returns immediately.
	WaitFor <-chan time.Time

	waitTime time.Duration

	// Holds a handler used to manipulate arguments content that are passed by
	// reference. It's useful when mocking methods such as unmarshalers or
	// decoders.
	RunFn func(Arguments)
}

type T struct {
	common
	isEnvSet bool
	context  *testContext // For running tests and subtests.
}
func (m *MockDB) Find(out interface{}, where ...interface{}) *gorm.DB {
	args := m.Called(out, where)
	return args.Get(0).(*gorm.DB)
}
func TestUserStoreGetByID(t *testing.T) {
	tests := []struct {
		name     string
		setupDB  func() *gorm.DB
		userID   uint
		expected *model.User
		wantErr  error
	}{
		{
			name: "Successfully retrieve an existing user by ID",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				user := &model.User{
					Model:    gorm.Model{ID: 1},
					Username: "testuser",
					Email:    "test@example.com",
					Password: "password",
				}
				db.Create(user)
				return db
			},
			userID: 1,
			expected: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password",
			},
			wantErr: nil,
		},
		{
			name: "Attempt to retrieve a non-existent user",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				return db
			},
			userID:   999,
			expected: nil,
			wantErr:  gorm.ErrRecordNotFound,
		},
		{
			name: "Handle database connection error",
			setupDB: func() *gorm.DB {
				mockDB := new(MockDB)
				mockDB.On("Find", mock.Anything, mock.Anything).Return(&gorm.DB{Error: errors.New("database error")})
				return &gorm.DB{Error: errors.New("database error")}
			},
			userID:   1,
			expected: nil,
			wantErr:  errors.New("database error"),
		},
		{
			name: "Retrieve a user with minimum fields set",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				user := &model.User{
					Model:    gorm.Model{ID: 1},
					Username: "minuser",
					Email:    "min@example.com",
					Password: "minpass",
				}
				db.Create(user)
				return db
			},
			userID: 1,
			expected: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "minuser",
				Email:    "min@example.com",
				Password: "minpass",
			},
			wantErr: nil,
		},
		{
			name: "Retrieve a user with all fields populated",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				user := &model.User{
					Model:    gorm.Model{ID: 1},
					Username: "fulluser",
					Email:    "full@example.com",
					Password: "fullpass",
					Bio:      "Full bio",
					Image:    "full.jpg",
				}
				db.Create(user)
				return db
			},
			userID: 1,
			expected: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "fulluser",
				Email:    "full@example.com",
				Password: "fullpass",
				Bio:      "Full bio",
				Image:    "full.jpg",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setupDB()
			s := &UserStore{db: db}

			got, err := s.GetByID(tt.userID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
