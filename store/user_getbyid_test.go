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


type Article struct {
	gorm.Model
	Title          string `gorm:"not null"`
	Description    string `gorm:"not null"`
	Body           string `gorm:"not null"`
	Tags           []Tag  `gorm:"many2many:article_tags"`
	Author         User   `gorm:"foreignkey:UserID"`
	UserID         uint   `gorm:"not null"`
	FavoritesCount int32  `gorm:"not null;default=0"`
	FavoritedUsers []User `gorm:"many2many:favorite_articles"`
	Comments       []Comment
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
func TestGetByID(t *testing.T) {
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
