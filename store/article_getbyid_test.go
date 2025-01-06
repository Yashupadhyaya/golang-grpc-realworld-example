package store

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"github.com/jinzhu/gorm"
	"github.com/raahii/golang-grpc-realworld-example/model"
)






type T struct {
	common
	isEnvSet bool
	context  *testContext // For running tests and subtests.
}
func TestArticleStoreGetByID(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.Article
		wantErr bool
	}{
		{
			name: "Successfully retrieve an existing article by ID",
			fields: fields{
				db: &gorm.DB{},
			},
			args: args{id: 1},
			want: &model.Article{
				Model: gorm.Model{ID: 1},
				Title: "Test Article",
				Tags:  []model.Tag{{Model: gorm.Model{ID: 1}, Name: "test"}},
				Author: model.User{
					Model: gorm.Model{ID: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "Attempt to retrieve a non-existent article",
			fields: fields{
				db: &gorm.DB{},
			},
			args:    args{id: 999},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Retrieve an article with no associated tags",
			fields: fields{
				db: &gorm.DB{},
			},
			args: args{id: 2},
			want: &model.Article{
				Model: gorm.Model{ID: 2},
				Title: "No Tags Article",
				Tags:  []model.Tag{},
				Author: model.User{
					Model: gorm.Model{ID: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "Database connection error",
			fields: fields{
				db: &gorm.DB{Error: errors.New("connection error")},
			},
			args:    args{id: 1},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Retrieve an article with maximum values",
			fields: fields{
				db: &gorm.DB{},
			},
			args: args{id: 3},
			want: &model.Article{
				Model:          gorm.Model{ID: 3},
				Title:          "Very Long Title",
				Description:    "Very Long Description",
				Body:           "Very Long Body",
				Tags:           []model.Tag{{Model: gorm.Model{ID: 1}, Name: "tag1"}, {Model: gorm.Model{ID: 2}, Name: "tag2"}},
				FavoritesCount: 9999999,
				Author: model.User{
					Model: gorm.Model{ID: 1},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ArticleStore{
				db: tt.fields.db,
			}
			got, err := s.GetByID(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleStore.GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ArticleStore.GetByID() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestArticleStoreGetByIDConcurrent(t *testing.T) {

	s := &ArticleStore{
		db: &gorm.DB{},
	}

	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id uint) {
			defer wg.Done()
			_, err := s.GetByID(id)
			if err != nil {
				t.Errorf("Concurrent ArticleStore.GetByID(%d) error = %v", id, err)
			}
		}(uint(i))
	}

	wg.Wait()
}
