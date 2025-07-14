package handler

import model "github.com/raahii/golang-grpc-realworld-example/model"

/*
ROOST_METHOD_HASH=Handler_CreateComment_c9eeba8015
ROOST_METHOD_SIG_HASH=Handler_CreateComment_9348ec6f77

FUNCTION_DEF=func (h *Handler) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CommentResponse, error) // CreateComment create a comment for an article
*/
func (m *MockArticleStore) CreateComment(comment *model.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockArticleStore) GetByID(id uint) (*model.Article, error) {
	args := m.Called(id)
	return args.Get(0).(*model.Article), args.Error(1)
}
