package store

import (
	"reflect"
	"testing"
	"time"
	"github.com/jinzhu/gorm"
)



type T struct {
	common
	isEnvSet bool
	context  *testContext // For running tests and subtests.
}


type Time struct {
	// wall and ext encode the wall time seconds, wall time nanoseconds,
	// and optional monotonic clock reading in nanoseconds.
	//
	// From high to low bit position, wall encodes a 1-bit flag (hasMonotonic),
	// a 33-bit seconds field, and a 30-bit wall time nanoseconds field.
	// The nanoseconds field is in the range [0, 999999999].
	// If the hasMonotonic bit is 0, then the 33-bit field must be zero
	// and the full signed 64-bit wall seconds since Jan 1 year 1 is stored in ext.
	// If the hasMonotonic bit is 1, then the 33-bit field holds a 33-bit
	// unsigned wall seconds since Jan 1 year 1885, and ext holds a
	// signed 64-bit monotonic clock reading, nanoseconds since process start.
	wall uint64
	ext  int64

	// loc specifies the Location that should be used to
	// determine the minute, hour, month, day, and year
	// that correspond to this Time.
	// The nil location means UTC.
	// All UTC times are represented with loc==nil, never loc==&utcLoc.
	loc *Location
}
func TestNewArticleStore(t *testing.T) {
	tests := []struct {
		name string
		db   *gorm.DB
		want *ArticleStore
	}{
		{
			name: "Create ArticleStore with Valid DB Connection",
			db:   &gorm.DB{},
			want: &ArticleStore{db: &gorm.DB{}},
		},
		{
			name: "Create ArticleStore with Nil DB Connection",
			db:   nil,
			want: &ArticleStore{db: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewArticleStore(tt.db)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewArticleStore() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("Verify ArticleStore Immutability", func(t *testing.T) {
		db := &gorm.DB{}
		store1 := NewArticleStore(db)
		store2 := NewArticleStore(db)
		if store1 == store2 {
			t.Errorf("NewArticleStore() returned same instance for multiple calls")
		}
		if store1.db != store2.db {
			t.Errorf("NewArticleStore() returned different DB instances")
		}
	})

	t.Run("Check DB Field Accessibility", func(t *testing.T) {
		db := &gorm.DB{Value: "test"}
		store := NewArticleStore(db)
		if store.db != db {
			t.Errorf("NewArticleStore() did not set the correct DB field")
		}
	})

	t.Run("Performance Test for NewArticleStore", func(t *testing.T) {
		db := &gorm.DB{}
		start := time.Now()
		iterations := 1000
		for i := 0; i < iterations; i++ {
			NewArticleStore(db)
		}
		duration := time.Since(start)
		t.Logf("Time taken for %d iterations: %v", iterations, duration)

	})
}
