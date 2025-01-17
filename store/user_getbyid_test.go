package store

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
)

func TestUserStoreGetById(t *testing.T) {
	tests := []struct {
		name    string
		id      uint
		wantUser *model.User
		wantErr error
		dbSetup func(*gorm.DB)
	}{
		{
			name: "Successfully retrieve an existing user",
			id:   1,
			wantUser: &model.User{
				Model:    gorm.Model{ID: 1},
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password",
			},
			wantErr: nil,
			dbSetup: func(db *gorm.DB) {
				db.Create(&model.User{
					Model:    gorm.Model{ID: 1},
					Username: "testuser",
					Email:    "test@example.com",
					Password: "password",
				})
			},
		},
		{
			name:     "Attempt to retrieve a non-existent user",
			id:       999,
			wantUser: nil,
			wantErr:  gorm.ErrRecordNotFound,
			dbSetup:  func(db *gorm.DB) {},
		},
		{
			name:     "Handle database connection error",
			id:       1,
			wantUser: nil,
			wantErr:  errors.New("database connection error"),
			dbSetup: func(db *gorm.DB) {
				db.AddError(errors.New("database connection error"))
			},
		},
		{
			name: "Retrieve a user with minimum field values",
			id:   2,
			wantUser: &model.User{
				Model:    gorm.Model{ID: 2},
				Username: "minuser",
				Email:    "min@example.com",
				Password: "minpass",
			},
			wantErr: nil,
			dbSetup: func(db *gorm.DB) {
				db.Create(&model.User{
					Model:    gorm.Model{ID: 2},
					Username: "minuser",
					Email:    "min@example.com",
					Password: "minpass",
				})
			},
		},
		{
			name: "Retrieve a user with all fields populated",
			id:   3,
			wantUser: &model.User{
				Model:    gorm.Model{ID: 3},
				Username: "fulluser",
				Email:    "full@example.com",
				Password: "fullpass",
				Bio:      "Full bio",
				Image:    "full.jpg",
				Follows:  []model.User{{Model: gorm.Model{ID: 1}}},
				FavoriteArticles: []model.Article{{Model: gorm.Model{ID: 1}}},
			},
			wantErr: nil,
			dbSetup: func(db *gorm.DB) {
				user := &model.User{
					Model:    gorm.Model{ID: 3},
					Username: "fulluser",
					Email:    "full@example.com",
					Password: "fullpass",
					Bio:      "Full bio",
					Image:    "full.jpg",
				}
				db.Create(user)
				db.Model(user).Association("Follows").Append(&model.User{Model: gorm.Model{ID: 1}})
				db.Model(user).Association("FavoriteArticles").Append(&model.Article{Model: gorm.Model{ID: 1}})
			},
		},
		{
			name:     "Handle zero ID input",
			id:       0,
			wantUser: nil,
			wantErr:  gorm.ErrRecordNotFound,
			dbSetup:  func(db *gorm.DB) {},
		},
		{
			name: "Performance test with a large user ID",
			id:   math.MaxUint32,
			wantUser: &model.User{
				Model:    gorm.Model{ID: math.MaxUint32},
				Username: "largeuser",
				Email:    "large@example.com",
				Password: "largepass",
			},
			wantErr: nil,
			dbSetup: func(db *gorm.DB) {
				db.Create(&model.User{
					Model:    gorm.Model{ID: math.MaxUint32},
					Username: "largeuser",
					Email:    "large@example.com",
					Password: "largepass",
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock database
			db, _ := gorm.Open("sqlite3", ":memory:")
			defer db.Close()
			db.AutoMigrate(&model.User{}, &model.Article{})
			tt.dbSetup(db)

			// Create UserStore with mock database
			s := &UserStore{db: db}

			// Call the function
			gotUser, err := s.GetByID(tt.id)

			// Check error
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("UserStore.GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("UserStore.GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check user
			if !reflect.DeepEqual(gotUser, tt.wantUser) {
				t.Errorf("UserStore.GetByID() = %v, want %v", gotUser, tt.wantUser)
			}
		})
	}
}
