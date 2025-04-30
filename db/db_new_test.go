package db

import (
	"errors"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
)

// Mock struct for gorm.DB
type mockDB struct {
	gorm.DB
	maxIdleConns int
	logMode      bool
}

func (m *mockDB) DB() *mockDB {
	return m
}

func (m *mockDB) SetMaxIdleConns(n int) {
	m.maxIdleConns = n
}

func (m *mockDB) LogMode(enable bool) *mockDB {
	m.logMode = enable
	return m
}

// Mock function for gorm.Open
var mockGormOpen func(dialect string, args ...interface{}) (*gorm.DB, error)

// Mock function for dsn
var mockDSN func() (string, error)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		dsnFunc        func() (string, error)
		gormOpenFunc   func(dialect string, args ...interface{}) (*gorm.DB, error)
		expectedDB     *gorm.DB
		expectedError  error
		maxRetries     int
		expectedRetries int
	}{
		{
			name: "Successful Database Connection",
			dsnFunc: func() (string, error) {
				return "valid_dsn", nil
			},
			gormOpenFunc: func(dialect string, args ...interface{}) (*gorm.DB, error) {
				return &gorm.DB{}, nil
			},
			expectedDB:    &gorm.DB{},
			expectedError: nil,
			maxRetries:    1,
			expectedRetries: 1,
		},
		{
			name: "Connection Retry Mechanism",
			dsnFunc: func() (string, error) {
				return "valid_dsn", nil
			},
			gormOpenFunc: func(dialect string, args ...interface{}) (*gorm.DB, error) {
				((ROOST_MOCK_STRUCT)).retryCount++
				if ((ROOST_MOCK_STRUCT)).retryCount < 3 {
					return nil, errors.New("connection failed")
				}
				return &gorm.DB{}, nil
			},
			expectedDB:    &gorm.DB{},
			expectedError: nil,
			maxRetries:    10,
			expectedRetries: 3,
		},
		{
			name: "Maximum Retry Limit Reached",
			dsnFunc: func() (string, error) {
				return "valid_dsn", nil
			},
			gormOpenFunc: func(dialect string, args ...interface{}) (*gorm.DB, error) {
				return nil, errors.New("connection failed")
			},
			expectedDB:    nil,
			expectedError: errors.New("connection failed"),
			maxRetries:    10,
			expectedRetries: 10,
		},
		{
			name: "Database Configuration Error",
			dsnFunc: func() (string, error) {
				return "", errors.New("dsn error")
			},
			gormOpenFunc: func(dialect string, args ...interface{}) (*gorm.DB, error) {
				t.Fatal("gorm.Open should not be called")
				return nil, nil
			},
			expectedDB:    nil,
			expectedError: errors.New("dsn error"),
			maxRetries:    0,
			expectedRetries: 0,
		},
		{
			name: "Connection Pool Configuration",
			dsnFunc: func() (string, error) {
				return "valid_dsn", nil
			},
			gormOpenFunc: func(dialect string, args ...interface{}) (*gorm.DB, error) {
				return &gorm.DB{}, nil
			},
			expectedDB:    &gorm.DB{},
			expectedError: nil,
			maxRetries:    1,
			expectedRetries: 1,
		},
		{
			name: "LogMode Configuration",
			dsnFunc: func() (string, error) {
				return "valid_dsn", nil
			},
			gormOpenFunc: func(dialect string, args ...interface{}) (*gorm.DB, error) {
				return &gorm.DB{}, nil
			},
			expectedDB:    &gorm.DB{},
			expectedError: nil,
			maxRetries:    1,
			expectedRetries: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the dsn function
			dsn = tt.dsnFunc

			// Mock gorm.Open
			gorm.Open = tt.gormOpenFunc

			// Reset retry count
			((ROOST_MOCK_STRUCT)).retryCount = 0

			// Create a mock DB
			mockDB := &mockDB{}

			// Override gorm.Open to return our mockDB
			gorm.Open = func(dialect string, args ...interface{}) (*gorm.DB, error) {
				db, err := tt.gormOpenFunc(dialect, args...)
				if err == nil {
					return &mockDB.DB, nil
				}
				return db, err
			}

			// Call the function under test
			db, err := New()

			// Check the error
			if (err != nil) != (tt.expectedError != nil) {
				t.Errorf("New() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			// Check if the returned DB matches the expected DB
			if (db == nil) != (tt.expectedDB == nil) {
				t.Errorf("New() returned DB = %v, expected %v", db, tt.expectedDB)
			}

			// Check the number of retries
			if ((ROOST_MOCK_STRUCT)).retryCount != tt.expectedRetries {
				t.Errorf("New() retries = %d, expected %d", ((ROOST_MOCK_STRUCT)).retryCount, tt.expectedRetries)
			}

			// Check MaxIdleConns configuration
			if mockDB.maxIdleConns != 3 {
				t.Errorf("New() MaxIdleConns = %d, expected 3", mockDB.maxIdleConns)
			}

			// Check LogMode configuration
			if mockDB.logMode != false {
				t.Errorf("New() LogMode = %v, expected false", mockDB.logMode)
			}
		})
	}
}

// TODO: Implement the actual dsn function
var dsn = func() (string, error) {
	// Implement the actual dsn function here
	return "", nil
}

// Mock struct to keep track of retry count
type mockStruct struct {
	retryCount int
}

var ((ROOST_MOCK_STRUCT)) = &mockStruct{}
