package store

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
)









/*
ROOST_METHOD_HASH=GetArticles_6382a4fe7a
ROOST_METHOD_SIG_HASH=GetArticles_1a0b3b0e8b

FUNCTION_DEF=func (s *ArticleStore) GetArticles(tagName, username string, favoritedBy *model.User, limit, offset int64) ([]model.Article, error) 

 */
func TestArticleStoreGetArticles(t *testing.T) {

	type mockDB struct {
		findFunc    func(interface{}) *gorm.DB
		preloadFunc func(string, ...interface{}) *gorm.DB
		joinsFunc   func(string, ...interface{}) *gorm.DB
		whereFunc   func(interface{}, ...interface{}) *gorm.DB
		offsetFunc  func(interface{}) *gorm.DB
		limitFunc   func(interface{}) *gorm.DB
		selectFunc  func(interface{}, ...interface{}) *gorm.DB
		tableFunc   func(string) *gorm.DB
		rowsFunc    func() (*sql.Rows, error)
	}

	tests := []struct {
		name        string
		tagName     string
		username    string
		favoritedBy *model.User
		limit       int64
		offset      int64
		mockDB      mockDB
		want        []model.Article
		wantErr     bool
	}{
		{
			name:     "Retrieve Articles Without Filters",
			tagName:  "",
			username: "",
			limit:    10,
			offset:   0,
			mockDB: mockDB{
				preloadFunc: func(string, ...interface{}) *gorm.DB { return &gorm.DB{} },
				offsetFunc:  func(interface{}) *gorm.DB { return &gorm.DB{} },
				limitFunc:   func(interface{}) *gorm.DB { return &gorm.DB{} },
				findFunc: func(dest interface{}) *gorm.DB {
					reflect.ValueOf(dest).Elem().Set(reflect.ValueOf([]model.Article{
						{Model: gorm.Model{ID: 1}, Title: "Article 1"},
						{Model: gorm.Model{ID: 2}, Title: "Article 2"},
					}))
					return &gorm.DB{}
				},
			},
			want: []model.Article{
				{Model: gorm.Model{ID: 1}, Title: "Article 1"},
				{Model: gorm.Model{ID: 2}, Title: "Article 2"},
			},
		},
		{
			name:     "Filter Articles by Tag Name",
			tagName:  "golang",
			username: "",
			limit:    10,
			offset:   0,
			mockDB: mockDB{
				preloadFunc: func(string, ...interface{}) *gorm.DB { return &gorm.DB{} },
				joinsFunc:   func(string, ...interface{}) *gorm.DB { return &gorm.DB{} },
				whereFunc:   func(interface{}, ...interface{}) *gorm.DB { return &gorm.DB{} },
				offsetFunc:  func(interface{}) *gorm.DB { return &gorm.DB{} },
				limitFunc:   func(interface{}) *gorm.DB { return &gorm.DB{} },
				findFunc: func(dest interface{}) *gorm.DB {
					reflect.ValueOf(dest).Elem().Set(reflect.ValueOf([]model.Article{
						{Model: gorm.Model{ID: 1}, Title: "Golang Article"},
					}))
					return &gorm.DB{}
				},
			},
			want: []model.Article{
				{Model: gorm.Model{ID: 1}, Title: "Golang Article"},
			},
		},
		{
			name:     "Filter Articles by Author Username",
			tagName:  "",
			username: "johndoe",
			limit:    10,
			offset:   0,
			mockDB: mockDB{
				preloadFunc: func(string, ...interface{}) *gorm.DB { return &gorm.DB{} },
				joinsFunc:   func(string, ...interface{}) *gorm.DB { return &gorm.DB{} },
				whereFunc:   func(interface{}, ...interface{}) *gorm.DB { return &gorm.DB{} },
				offsetFunc:  func(interface{}) *gorm.DB { return &gorm.DB{} },
				limitFunc:   func(interface{}) *gorm.DB { return &gorm.DB{} },
				findFunc: func(dest interface{}) *gorm.DB {
					reflect.ValueOf(dest).Elem().Set(reflect.ValueOf([]model.Article{
						{Model: gorm.Model{ID: 1}, Title: "John's Article", Author: model.User{Username: "johndoe"}},
					}))
					return &gorm.DB{}
				},
			},
			want: []model.Article{
				{Model: gorm.Model{ID: 1}, Title: "John's Article", Author: model.User{Username: "johndoe"}},
			},
		},
		{
			name:        "Retrieve Favorited Articles",
			tagName:     "",
			username:    "",
			favoritedBy: &model.User{Model: gorm.Model{ID: 1}},
			limit:       10,
			offset:      0,
			mockDB: mockDB{
				preloadFunc: func(string, ...interface{}) *gorm.DB { return &gorm.DB{} },
				selectFunc:  func(interface{}, ...interface{}) *gorm.DB { return &gorm.DB{} },
				tableFunc:   func(string) *gorm.DB { return &gorm.DB{} },
				whereFunc:   func(interface{}, ...interface{}) *gorm.DB { return &gorm.DB{} },
				offsetFunc:  func(interface{}) *gorm.DB { return &gorm.DB{} },
				limitFunc:   func(interface{}) *gorm.DB { return &gorm.DB{} },
				rowsFunc: func() (*sql.Rows, error) {

					return nil, nil
				},
				findFunc: func(dest interface{}) *gorm.DB {
					reflect.ValueOf(dest).Elem().Set(reflect.ValueOf([]model.Article{
						{Model: gorm.Model{ID: 1}, Title: "Favorited Article"},
					}))
					return &gorm.DB{}
				},
			},
			want: []model.Article{
				{Model: gorm.Model{ID: 1}, Title: "Favorited Article"},
			},
		},
		{
			name:     "Handle Empty Result Set",
			tagName:  "nonexistent",
			username: "",
			limit:    10,
			offset:   0,
			mockDB: mockDB{
				preloadFunc: func(string, ...interface{}) *gorm.DB { return &gorm.DB{} },
				joinsFunc:   func(string, ...interface{}) *gorm.DB { return &gorm.DB{} },
				whereFunc:   func(interface{}, ...interface{}) *gorm.DB { return &gorm.DB{} },
				offsetFunc:  func(interface{}) *gorm.DB { return &gorm.DB{} },
				limitFunc:   func(interface{}) *gorm.DB { return &gorm.DB{} },
				findFunc:    func(interface{}) *gorm.DB { return &gorm.DB{} },
			},
			want: []model.Article{},
		},
		{
			name:     "Error Handling for Database Issues",
			tagName:  "",
			username: "",
			limit:    10,
			offset:   0,
			mockDB: mockDB{
				preloadFunc: func(string, ...interface{}) *gorm.DB { return &gorm.DB{} },
				offsetFunc:  func(interface{}) *gorm.DB { return &gorm.DB{} },
				limitFunc:   func(interface{}) *gorm.DB { return &gorm.DB{} },
				findFunc: func(interface{}) *gorm.DB {
					return &gorm.DB{Error: errors.New("database error")}
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ArticleStore{
				db: &gorm.DB{},
			}

			s.db.Callback().Query().Register("test_query", func(scope *gorm.Scope) {
				if tt.mockDB.findFunc != nil {
					scope.DB().AddError(tt.mockDB.findFunc(scope.Value).Error)
				}
			})

			got, err := s.GetArticles(tt.tagName, tt.username, tt.favoritedBy, tt.limit, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.GetArticles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ArticleStore.GetArticles() = %v, want %v", got, tt.want)
			}
		})
	}
}

