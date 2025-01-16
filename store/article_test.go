package store

import (
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
	type args struct {
		tagName     string
		username    string
		favoritedBy *model.User
		limit       int64
		offset      int64
	}

	tests := []struct {
		name    string
		args    args
		want    []model.Article
		wantErr bool
		mockDB  func() *gorm.DB
	}{
		{
			name: "Retrieve Articles Without Any Filters",
			args: args{
				tagName:     "",
				username:    "",
				favoritedBy: nil,
				limit:       10,
				offset:      0,
			},
			want: []model.Article{
				{Model: gorm.Model{ID: 1}, Title: "Article 1", Author: model.User{Model: gorm.Model{ID: 1}, Username: "user1"}},
				{Model: gorm.Model{ID: 2}, Title: "Article 2", Author: model.User{Model: gorm.Model{ID: 2}, Username: "user2"}},
			},
			wantErr: false,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(nil)
				return db
			},
		},
		{
			name: "Filter Articles by Tag Name",
			args: args{
				tagName:     "golang",
				username:    "",
				favoritedBy: nil,
				limit:       10,
				offset:      0,
			},
			want: []model.Article{
				{Model: gorm.Model{ID: 1}, Title: "Golang Article", Author: model.User{Model: gorm.Model{ID: 1}, Username: "user1"}},
			},
			wantErr: false,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(nil)
				return db
			},
		},
		{
			name: "Filter Articles by Author Username",
			args: args{
				tagName:     "",
				username:    "user1",
				favoritedBy: nil,
				limit:       10,
				offset:      0,
			},
			want: []model.Article{
				{Model: gorm.Model{ID: 1}, Title: "User1 Article", Author: model.User{Model: gorm.Model{ID: 1}, Username: "user1"}},
			},
			wantErr: false,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(nil)
				return db
			},
		},
		{
			name: "Retrieve Favorited Articles",
			args: args{
				tagName:     "",
				username:    "",
				favoritedBy: &model.User{Model: gorm.Model{ID: 1}},
				limit:       10,
				offset:      0,
			},
			want: []model.Article{
				{Model: gorm.Model{ID: 2}, Title: "Favorited Article", Author: model.User{Model: gorm.Model{ID: 2}, Username: "user2"}},
			},
			wantErr: false,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(nil)
				return db
			},
		},
		{
			name: "Combine Multiple Filters",
			args: args{
				tagName:     "golang",
				username:    "user1",
				favoritedBy: nil,
				limit:       10,
				offset:      0,
			},
			want: []model.Article{
				{Model: gorm.Model{ID: 1}, Title: "Golang Article by User1", Author: model.User{Model: gorm.Model{ID: 1}, Username: "user1"}},
			},
			wantErr: false,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(nil)
				return db
			},
		},
		{
			name: "Handle Empty Result Set",
			args: args{
				tagName:     "nonexistent",
				username:    "",
				favoritedBy: nil,
				limit:       10,
				offset:      0,
			},
			want:    []model.Article{},
			wantErr: false,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(nil)
				return db
			},
		},
		{
			name: "Test Pagination Limits",
			args: args{
				tagName:     "",
				username:    "",
				favoritedBy: nil,
				limit:       2,
				offset:      1,
			},
			want: []model.Article{
				{Model: gorm.Model{ID: 2}, Title: "Article 2", Author: model.User{Model: gorm.Model{ID: 2}, Username: "user2"}},
				{Model: gorm.Model{ID: 3}, Title: "Article 3", Author: model.User{Model: gorm.Model{ID: 3}, Username: "user3"}},
			},
			wantErr: false,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(nil)
				return db
			},
		},
		{
			name: "Error Handling for Database Issues",
			args: args{
				tagName:     "",
				username:    "",
				favoritedBy: nil,
				limit:       10,
				offset:      0,
			},
			want:    []model.Article{},
			wantErr: true,
			mockDB: func() *gorm.DB {
				db := &gorm.DB{}
				db.AddError(errors.New("database error"))
				return db
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ArticleStore{
				db: tt.mockDB(),
			}
			got, err := s.GetArticles(tt.args.tagName, tt.args.username, tt.args.favoritedBy, tt.args.limit, tt.args.offset)
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

