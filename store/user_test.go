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


/*
ROOST_METHOD_HASH=GetByID_bbf946112e
ROOST_METHOD_SIG_HASH=GetByID_728dd55ed1


 */
func TestUserStoreGetByID(t *testing.T) {
	tests := []struct {
		name     string
		setupDB  func() *gorm.DB
		userID   uint
		expected *model.User
		wantErr  bool
	}{
		{
			name: "Successfully retrieve a user by ID",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				user := &model.User{
					Model:    gorm.Model{ID: 1},
					Username: "testuser",
					Email:    "test@example.com",
					Password: "password",
					Bio:      "Test bio",
					Image:    "test.jpg",
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
				Bio:      "Test bio",
				Image:    "test.jpg",
			},
			wantErr: false,
		},
		{
			name: "Attempt to retrieve a non-existent user",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				return db
			},
			userID:   999,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Handle database connection error",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				db.Close()
				return db
			},
			userID:   1,
			expected: nil,
			wantErr:  true,
		},
		{
			name: "Retrieve a user with all fields populated",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				user := &model.User{
					Model:    gorm.Model{ID: 1},
					Username: "fulluser",
					Email:    "full@example.com",
					Password: "password",
					Bio:      "Full bio",
					Image:    "full.jpg",
					Follows:  []model.User{{Model: gorm.Model{ID: 2}, Username: "follower"}},
					FavoriteArticles: []model.Article{
						{Model: gorm.Model{ID: 1}, Title: "Favorite Article"},
					},
				}
				db.Create(user)
				return db
			},
			userID: 1,
			expected: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "fulluser",
				Email:    "full@example.com",
				Password: "password",
				Bio:      "Full bio",
				Image:    "full.jpg",
				Follows:  []model.User{{Model: gorm.Model{ID: 2}, Username: "follower"}},
				FavoriteArticles: []model.Article{
					{Model: gorm.Model{ID: 1}, Title: "Favorite Article"},
				},
			},
			wantErr: false,
		},
		{
			name: "Retrieve a user with minimal information",
			setupDB: func() *gorm.DB {
				db, _ := gorm.Open("sqlite3", ":memory:")
				user := &model.User{
					Model:    gorm.Model{ID: 1},
					Username: "minuser",
					Email:    "min@example.com",
					Password: "password",
				}
				db.Create(user)
				return db
			},
			userID: 1,
			expected: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "minuser",
				Email:    "min@example.com",
				Password: "password",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := tt.setupDB()
			s := &UserStore{db: db}

			user, err := s.GetByID(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.expected.ID, user.ID)
				assert.Equal(t, tt.expected.Username, user.Username)
				assert.Equal(t, tt.expected.Email, user.Email)
				assert.Equal(t, tt.expected.Password, user.Password)
				assert.Equal(t, tt.expected.Bio, user.Bio)
				assert.Equal(t, tt.expected.Image, user.Image)

				if tt.expected.Follows != nil {
					assert.Equal(t, len(tt.expected.Follows), len(user.Follows))
					for i, follow := range tt.expected.Follows {
						assert.Equal(t, follow.ID, user.Follows[i].ID)
						assert.Equal(t, follow.Username, user.Follows[i].Username)
					}
				}

				if tt.expected.FavoriteArticles != nil {
					assert.Equal(t, len(tt.expected.FavoriteArticles), len(user.FavoriteArticles))
					for i, article := range tt.expected.FavoriteArticles {
						assert.Equal(t, article.ID, user.FavoriteArticles[i].ID)
						assert.Equal(t, article.Title, user.FavoriteArticles[i].Title)
					}
				}
			}
		})
	}
}

