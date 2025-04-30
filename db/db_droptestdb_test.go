package db

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
)

type mockDB struct {
	closeCallCount int
	closeError     error
	mu             sync.Mutex
}

func (m *mockDB) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeCallCount++
	return m.closeError
}

func TestDropTestDb(t *testing.T) {
	tests := []struct {
		name    string
		db      *gorm.DB
		wantErr bool
	}{
		{
			name: "Successfully Close Database Connection",
			db: &gorm.DB{
				db: &mockDB{},
			},
			wantErr: false,
		},
		{
			name:    "Handle Nil Database Instance",
			db:      nil,
			wantErr: false,
		},
		{
			name: "Error During Database Closure",
			db: &gorm.DB{
				db: &mockDB{closeError: errors.New("close error")},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DropTestDB(tt.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("DropTestDB() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.db != nil {
				mockDB := tt.db.db.(*mockDB)
				if mockDB.closeCallCount != 1 {
					t.Errorf("Expected Close() to be called once, got %d", mockDB.closeCallCount)
				}
			}
		})
	}
}

func TestDropTestDbConcurrent(t *testing.T) {
	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			db := &gorm.DB{db: &mockDB{}}
			err := DropTestDB(db)
			if err != nil {
				t.Errorf("Concurrent DropTestDB() error = %v", err)
			}
		}()
	}

	wg.Wait()
}

func TestDropTestDbPerformance(t *testing.T) {
	const numIterations = 10000
	start := time.Now()

	for i := 0; i < numIterations; i++ {
		db := &gorm.DB{db: &mockDB{}}
		err := DropTestDB(db)
		if err != nil {
			t.Errorf("Performance DropTestDB() error = %v", err)
		}
	}

	duration := time.Since(start)
	t.Logf("Time taken for %d iterations: %v", numIterations, duration)
}

func TestDropTestDbIdempotency(t *testing.T) {
	mockDB := &mockDB{}
	db := &gorm.DB{db: mockDB}

	// First call
	err := DropTestDB(db)
	if err != nil {
		t.Errorf("First DropTestDB() call error = %v", err)
	}

	// Second call
	err = DropTestDB(db)
	if err != nil {
		t.Errorf("Second DropTestDB() call error = %v", err)
	}

	if mockDB.closeCallCount != 2 {
		t.Errorf("Expected Close() to be called twice, got %d", mockDB.closeCallCount)
	}
}
