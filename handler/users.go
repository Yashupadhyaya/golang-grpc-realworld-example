package handler

import (
	"context"
	"fmt"

	"github.com/raahii/golang-grpc-realworld-example/model"
	pb "github.com/raahii/golang-grpc-realworld-example/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) ShowProfile(ctx context.Context, req *pb.ShowProfileRequest) (*pb.ShowProfileResponse, error) {
	h.logger.Info().Msgf("Show profile | req: %+v\n", req)

	user := model.User{}
	err := h.db.Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		h.logger.Fatal().Err(fmt.Errorf("user not found: %w", err))
		return nil, status.Error(codes.NotFound, "user was not found")
	}

	var bio string
	if user.Bio != nil {
		bio = *user.Bio
	}

	var image string
	if user.Image != nil {
		image = *user.Image
	}

	p := pb.Profile{
		Username: req.Username,
		Bio:      bio,
		Image:    image,
	}

	return &pb.ShowProfileResponse{Profile: &p}, nil
}