package db

import (
	"database/sql"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-txdb"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
)

// Mock struct for gorm.DB
type mockDB struct {
	*gorm.DB
}

func (m *mockDB) SetMaxIdleConns(n int) *gorm.DB {
	return m.DB
}

func (m *mockDB) LogMode(enable bool) *gorm.DB {
	return m.DB
}

func TestNewTestDb(t *testing.T) {
	// Mock functions
	originalGodotenvLoad := godotenv.Load
	originalGormOpen := gorm.Open
	originalSqlOpen := sql.Open
	originalAutoMigrate := AutoMigrate

	defer func() {
		godotenv.Load = originalGodotenvLoad
		gorm.Open = originalGormOpen
		sql.Open = originalSqlOpen
		AutoMigrate = originalAutoMigrate
	}()

	tests := []struct {
		name           string
		setupMock      func()
		expectedDB     bool
		expectedErrMsg string
	}{
		{
			name: "Successful Database Connection and Initialization",
			setupMock: func() {
				godotenv.Load = func(filenames ...string) error { return nil }
				gorm.Open = func(dialect string, args ...interface{}) (*gorm.DB, error) {
					return &gorm.DB{}, nil
				}
				sql.Open = func(driverName, dataSourceName string) (*sql.DB, error) {
					return &sql.DB{}, nil
				}
				AutoMigrate = func(db *gorm.DB) {}
				txdb.Register("txdb", "mysql", "mock_dsn")
			},
			expectedDB: true,
		},
		{
			name: "Environment File Not Found",
			setupMock: func() {
				godotenv.Load = func(filenames ...string) error {
					return errors.New("env file not found")
				}
			},
			expectedErrMsg: "env file not found",
		},
		{
			name: "Invalid Database Credentials",
			setupMock: func() {
				godotenv.Load = func(filenames ...string) error { return nil }
				gorm.Open = func(dialect string, args ...interface{}) (*gorm.DB, error) {
					return nil, errors.New("invalid database credentials")
				}
			},
			expectedErrMsg: "invalid database credentials",
		},
		{
			name: "Error in sql.Open",
			setupMock: func() {
				godotenv.Load = func(filenames ...string) error { return nil }
				gorm.Open = func(dialect string, args ...interface{}) (*gorm.DB, error) {
					return &gorm.DB{}, nil
				}
				sql.Open = func(driverName, dataSourceName string) (*sql.DB, error) {
					return nil, errors.New("sql open error")
				}
			},
			expectedErrMsg: "sql open error",
		},
		{
			name: "Error in gorm.Open with sql.DB",
			setupMock: func() {
				godotenv.Load = func(filenames ...string) error { return nil }
				gorm.Open = func(dialect string, args ...interface{}) (*gorm.DB, error) {
					if _, ok := args[0].(*sql.DB); ok {
						return nil, errors.New("gorm open error")
					}
					return &gorm.DB{}, nil
				}
				sql.Open = func(driverName, dataSourceName string) (*sql.DB, error) {
					return &sql.DB{}, nil
				}
			},
			expectedErrMsg: "gorm open error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			db, err := NewTestDB()

			if tt.expectedDB && db == nil {
				t.Error("Expected non-nil DB, got nil")
			}

			if tt.expectedErrMsg != "" {
				if err == nil {
					t.Errorf("Expected error with message '%s', got nil", tt.expectedErrMsg)
				} else if err.Error() != tt.expectedErrMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedErrMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestNewTestDbConcurrency(t *testing.T) {
	// Mock functions
	originalGodotenvLoad := godotenv.Load
	originalGormOpen := gorm.Open
	originalSqlOpen := sql.Open
	originalAutoMigrate := AutoMigrate

	defer func() {
		godotenv.Load = originalGodotenvLoad
		gorm.Open = originalGormOpen
		sql.Open = originalSqlOpen
		AutoMigrate = originalAutoMigrate
	}()

	godotenv.Load = func(filenames ...string) error { return nil }
	gorm.Open = func(dialect string, args ...interface{}) (*gorm.DB, error) {
		return &gorm.DB{}, nil
	}
	sql.Open = func(driverName, dataSourceName string) (*sql.DB, error) {
		return &sql.DB{}, nil
	}
	AutoMigrate = func(db *gorm.DB) {}
	txdb.Register("txdb", "mysql", "mock_dsn")

	var wg sync.WaitGroup
	concurrentCalls := 10

	for i := 0; i < concurrentCalls; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db, err := NewTestDB()
			if err != nil {
				t.Errorf("Unexpected error in concurrent call: %v", err)
			}
			if db == nil {
				t.Error("Expected non-nil DB in concurrent call, got nil")
			}
		}()
	}

	wg.Wait()
}

func TestNewTestDbConnectionPool(t *testing.T) {
	// Mock functions
	originalGodotenvLoad := godotenv.Load
	originalGormOpen := gorm.Open
	originalSqlOpen := sql.Open
	originalAutoMigrate := AutoMigrate

	defer func() {
		godotenv.Load = originalGodotenvLoad
		gorm.Open = originalGormOpen
		sql.Open = originalSqlOpen
		AutoMigrate = originalAutoMigrate
	}()

	godotenv.Load = func(filenames ...string) error { return nil }
	gorm.Open = func(dialect string, args ...interface{}) (*gorm.DB, error) {
		return &mockDB{&gorm.DB{}}, nil
	}
	sql.Open = func(driverName, dataSourceName string) (*sql.DB, error) {
		return &sql.DB{}, nil
	}
	AutoMigrate = func(db *gorm.DB) {}
	txdb.Register("txdb", "mysql", "mock_dsn")

	db, err := NewTestDB()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if db == nil {
		t.Fatal("Expected non-nil DB, got nil")
	}

	// Note: We can't actually test the SetMaxIdleConns and LogMode calls
	// without modifying the original function or using a more sophisticated mock.
	// The mock we've set up will just return without error, which is the best
	// we can do without changing the original code.
}

func TestNewTestDbUniqueConnections(t *testing.T) {
	// Mock functions
	originalGodotenvLoad := godotenv.Load
	originalGormOpen := gorm.Open
	originalSqlOpen := sql.Open
	originalAutoMigrate := AutoMigrate

	defer func() {
		godotenv.Load = originalGodotenvLoad
		gorm.Open = originalGormOpen
		sql.Open = originalSqlOpen
		AutoMigrate = originalAutoMigrate
	}()

	godotenv.Load = func(filenames ...string) error { return nil }
	gorm.Open = func(dialect string, args ...interface{}) (*gorm.DB, error) {
		return &gorm.DB{}, nil
	}

	connections := make([]*sql.DB, 0)
	sql.Open = func(driverName, dataSourceName string) (*sql.DB, error) {
		db := &sql.DB{}
		connections = append(connections, db)
		return db, nil
	}
	AutoMigrate = func(db *gorm.DB) {}
	txdb.Register("txdb", "mysql", "mock_dsn")

	db1, err := NewTestDB()
	if err != nil {
		t.Fatalf("Unexpected error in first call: %v", err)
	}

	db2, err := NewTestDB()
	if err != nil {
		t.Fatalf("Unexpected error in second call: %v", err)
	}

	if db1 == db2 {
		t.Error("Expected unique DB instances, got the same instance")
	}

	if len(connections) != 2 {
		t.Errorf("Expected 2 unique connections, got %d", len(connections))
	}

	if connections[0] == connections[1] {
		t.Error("Expected unique underlying sql.DB connections, got the same connection")
	}
}

// TODO: Add more specific tests for dsn() function if needed
// TODO: Add tests for AutoMigrate function if it's exported and needs testing
