package grpcadapter

import (
	"context"
	"errors"

	apperror "be-modami-user-service/internal/apperror"
	"be-modami-user-service/internal/service"

	"github.com/google/uuid"
	pb "gitlab.com/lifegoeson-libs/pkg-techinsights-grpc-client/go/modami/user"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

type UserGRPCServer struct {
	pb.UnimplementedUserInternalServiceServer
	userService   *service.UserService
	sellerService *service.SellerService
}

func NewUserGRPCServer(userService *service.UserService, sellerService *service.SellerService) *UserGRPCServer {
	return &UserGRPCServer{
		userService:   userService,
		sellerService: sellerService,
	}
}

func (s *UserGRPCServer) GetUserBasic(ctx context.Context, req *pb.GetUserBasicRequest) (*pb.UserBasicResponse, error) {
	id, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, grpcstatus.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	user, err := s.userService.GetProfile(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, grpcstatus.Errorf(codes.NotFound, "user not found")
		}
		return nil, grpcstatus.Errorf(codes.Internal, "internal error: %v", err)
	}

	return &pb.UserBasicResponse{
		Id:         user.ID.String(),
		FullName:   user.FullName,
		AvatarUrl:  user.AvatarURL,
		Role:       string(user.Role),
		Status:     string(user.Status),
		TrustScore: user.TrustScore,
	}, nil
}

func (s *UserGRPCServer) GetUsersByIDs(ctx context.Context, req *pb.GetUsersByIDsRequest) (*pb.UsersResponse, error) {
	var users []*pb.UserBasicResponse
	for _, idStr := range req.GetUserIds() {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		user, err := s.userService.GetProfile(ctx, id)
		if err != nil {
			continue
		}
		users = append(users, &pb.UserBasicResponse{
			Id:         user.ID.String(),
			FullName:   user.FullName,
			AvatarUrl:  user.AvatarURL,
			Role:       string(user.Role),
			Status:     string(user.Status),
			TrustScore: user.TrustScore,
		})
	}
	return &pb.UsersResponse{Users: users}, nil
}

func (s *UserGRPCServer) CheckUserStatus(ctx context.Context, req *pb.CheckUserStatusRequest) (*pb.UserStatusResponse, error) {
	id, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, grpcstatus.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	user, err := s.userService.GetProfile(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, grpcstatus.Errorf(codes.NotFound, "user not found")
		}
		return nil, grpcstatus.Errorf(codes.Internal, "internal error: %v", err)
	}

	return &pb.UserStatusResponse{
		Status:   string(user.Status),
		IsActive: user.Status == "active",
	}, nil
}

func (s *UserGRPCServer) GetSellerInfo(ctx context.Context, req *pb.GetSellerInfoRequest) (*pb.SellerInfoResponse, error) {
	id, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, grpcstatus.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	profile, err := s.sellerService.GetShopProfile(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrSellerNotFound) || errors.Is(err, apperror.ErrNotFound) {
			return nil, grpcstatus.Errorf(codes.NotFound, "seller not found")
		}
		return nil, grpcstatus.Errorf(codes.Internal, "internal error: %v", err)
	}

	return &pb.SellerInfoResponse{
		UserId:       profile.UserID.String(),
		ShopName:     profile.ShopName,
		ShopSlug:     profile.ShopSlug,
		ShopLogoUrl:  profile.ShopLogoURL,
		KycStatus:    string(profile.KYCStatus),
		AvgRating:    profile.AvgRating,
		TotalReviews: int32(profile.TotalReviews),
	}, nil
}
