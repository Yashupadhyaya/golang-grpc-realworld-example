package handler

import (
	model "github.com/raahii/golang-grpc-realworld-example/model"
	mock "github.com/stretchr/testify/mock"
)

type MockArticleStore struct {
	mock.Mock
}
type MockUserStore struct {
	mock.Mock
}

func (m *MockArticleStore) Create(a *model.Article) error {
	args := m.Called(a)
	return args.Error(0)
}

func (m *MockUserStore) GetByID(id uint) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserStore) IsFollowing(a *model.User, b *model.User) (bool, error) {
	args := m.Called(a, b)
	return args.Bool(0), args.Error(1)
}
