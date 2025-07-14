package handler

import model "github.com/raahii/golang-grpc-realworld-example/model"

func (m *MockArticleStore) CreateComment(comment *model.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockArticleStore) GetByID(id uint) (*model.Article, error) {
	args := m.Called(id)
	return args.Get(0).(*model.Article), args.Error(1)
}
