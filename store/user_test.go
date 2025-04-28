package store

import (
		"testing"
		"github.com/jinzhu/gorm"
		"github.com/raahii/golang-grpc-realworld-example/model"
		"github.com/stretchr/testify/assert"
)

type T struct {
	common
	isEnvSet bool
	context  *testContext
}
type DB struct {
	sync.RWMutex
	Value        interface{}
	Error        error
	RowsAffected int64

	// single db
	db                SQLCommon
	blockGlobalUpdate bool
	logMode           logModeValue
	logger            logger
	search            *search
	values            sync.Map

	// global db
	parent        *DB
	callbacks     *Callback
	dialect       Dialect
	singularTable bool

	// function to be used to override the creating of a new timestamp
	nowFuncOverride func() time.Time
}

type User struct {
	gorm.Model
	Username         string    `gorm:"unique_index;not null"`
	Email            string    `gorm:"unique_index;not null"`
	Password         string    `gorm:"not null"`
	Bio              string    `gorm:"not null"`
	Image            string    `gorm:"not null"`
	Follows          []User    `gorm:"many2many:follows;jointable_foreignkey:from_user_id;association_jointable_foreignkey:to_user_id"`
	FavoriteArticles []Article `gorm:"many2many:favorite_articles;"`
}

type UserStore struct {
	db *gorm.DB
}

type T struct {
	common
	isEnvSet bool
	context  *testContext // For running tests and subtests.
}

/*
ROOST_METHOD_HASH=GetByID_bbf946112e
ROOST_METHOD_SIG_HASH=GetByID_728dd55ed1


 */
func TestUserStoreGetByID(t *testing.T) {
	tests := []struct {
		name     string
		userID   uint
		mockDB   func() *gorm.DB
		expected *model.User
		wantErr  bool
	}{
		{
			name:   "Successfully retrieve a user by ID",
			userID: 1,
			mockDB: func() *gorm.DB {
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
			expected: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password",
			},
			wantErr: false,
		},
		{
			name:   "Attempt to retrieve a non-existent user",
			userID: 999,
			mockDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				return db
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name:   "Handle database connection error",
			userID: 1,
			mockDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				db.Close()
				return db
			},
			expected: nil,
			wantErr:  true,
		},
		{
			name:   "Retrieve a user with minimum fields populated",
			userID: 2,
			mockDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				user := &model.User{
					Model:    gorm.Model{ID: 2},
					Username: "minuser",
					Email:    "min@example.com",
					Password: "minpass",
				}
				db.Create(user)
				return db
			},
			expected: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "minuser",
				Email:    "min@example.com",
				Password: "minpass",
			},
			wantErr: false,
		},
		{
			name:   "Verify retrieval of user with associated data",
			userID: 3,
			mockDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				user := &model.User{
					Model:    gorm.Model{ID: 3},
					Username: "assocuser",
					Email:    "assoc@example.com",
					Password: "assocpass",
					Follows: []model.User{
						{Model: gorm.Model{ID: 4}, Username: "follower1"},
					},
					FavoriteArticles: []model.Article{
						{Model: gorm.Model{ID: 1}, Title: "Favorite Article"},
					},
				}
				db.Create(user)
				return db
			},
			expected: &model.User{
				Model:    gorm.Model{ID: 3},
				Username: "assocuser",
				Email:    "assoc@example.com",
				Password: "assocpass",
				Follows: []model.User{
					{Model: gorm.Model{ID: 4}, Username: "follower1"},
				},
				FavoriteArticles: []model.Article{
					{Model: gorm.Model{ID: 1}, Title: "Favorite Article"},
				},
			},
			wantErr: false,
		},
		{
			name:   "Performance test with a large user ID",
			userID: 4294967295,
			mockDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				return db
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &UserStore{db: tt.mockDB()}
			user, err := store.GetByID(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, user)
			}

			if tt.name == "Verify retrieval of user with associated data" {
				assert.Len(t, user.Follows, 1)
				assert.Equal(t, "follower1", user.Follows[0].Username)
				assert.Len(t, user.FavoriteArticles, 1)
				assert.Equal(t, "Favorite Article", user.FavoriteArticles[0].Title)
			}
		})
	}
}

func (s *ArticleStore) GetByID(id uint) (*model.Article, error) {
	var m model.Article
	err := s.db.Preload("Tags").Preload("Author").Find(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

