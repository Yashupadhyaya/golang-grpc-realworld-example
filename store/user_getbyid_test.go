package store

import (
	"errors"
	"testing"
	"time"
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

type Mock struct {
	// Represents the calls that are expected of
	// an object.
	ExpectedCalls []*Call

	// Holds the calls that were made to this mocked object.
	Calls []Call

	// test is An optional variable that holds the test struct, to be used when an
	// invalid mock call was made.
	test TestingT

	// TestData holds any data that might be useful for testing.  Testify ignores
	// this data completely allowing you to do whatever you like with it.
	testData objx.Map

	mutex sync.Mutex
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
		name          string
		userID        uint
		mockSetup     func(*MockDB)
		expectedUser  *model.User
		expectedError error
	}{
		{
			name:   "Successfully retrieve a user by ID",
			userID: 1,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Find", mock.AnythingOfType("*model.User"), uint(1)).Return(&gorm.DB{Error: nil})
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				Username: "testuser",
				Email:    "test@example.com",
				Bio:      "Test bio",
				Image:    "test.jpg",
			},
			expectedError: nil,
		},
		{
			name:   "Attempt to retrieve a non-existent user",
			userID: 999,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Find", mock.AnythingOfType("*model.User"), uint(999)).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:   "Handle database connection error",
			userID: 2,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Find", mock.AnythingOfType("*model.User"), uint(2)).Return(&gorm.DB{Error: errors.New("database connection error")})
			},
			expectedUser:  nil,
			expectedError: errors.New("database connection error"),
		},
		{
			name:   "Retrieve a user with minimum field values",
			userID: 3,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Find", mock.AnythingOfType("*model.User"), uint(3)).Return(&gorm.DB{Error: nil})
			},
			expectedUser: &model.User{
				Model:    gorm.Model{ID: 3, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				Username: "minuser",
				Email:    "min@example.com",
			},
			expectedError: nil,
		},
		{
			name:   "Retrieve a user with all fields populated",
			userID: 4,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Find", mock.AnythingOfType("*model.User"), uint(4)).Return(&gorm.DB{Error: nil})
			},
			expectedUser: &model.User{
				Model:            gorm.Model{ID: 4, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				Username:         "fulluser",
				Email:            "full@example.com",
				Bio:              "Full bio",
				Image:            "full.jpg",
				Follows:          []model.User{{Model: gorm.Model{ID: 5}}},
				FavoriteArticles: []model.Article{{Model: gorm.Model{ID: 1}}},
			},
			expectedError: nil,
		},
		{
			name:   "Handle zero value ID",
			userID: 0,
			mockSetup: func(mockDB *MockDB) {
				mockDB.On("Find", mock.AnythingOfType("*model.User"), uint(0)).Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			expectedUser:  nil,
			expectedError: gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := new(MockDB)
			tt.mockSetup(mockDB)

			mockGormDB := &MockGormDB{MockDB: mockDB}

			store := &UserStore{db: mockGormDB}

			user, err := store.GetByID(tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedUser, user)

			mockDB.AssertExpectations(t)
		})
	}
}
