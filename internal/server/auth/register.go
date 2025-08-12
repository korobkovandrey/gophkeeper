package auth

import (
	"context"
	"errors"
	"fmt"
	"gophkeeper/internal/server/service"
	"gophkeeper/pkg/crypt"
	"gophkeeper/pkg/proto"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Register регистрирует нового пользователя по публичному ключу.
func (s *ServiceServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.AuthResponse, error) {
	startTime := timestamppb.Now()
	if len(req.PublicKeyBytes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "PublicKey is required")
	}
	if len(req.Signature) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Signature is required")
	}
	currentTime := time.Now().Unix()
	if req.Timestamp < currentTime-s.timeWindow || req.Timestamp > currentTime+s.timeWindow {
		return nil, status.Error(codes.InvalidArgument, "Timestamp is out of range")
	}
	rsaPublicKey, err := crypt.RSAPublicKeyFromBytes(req.PublicKeyBytes)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to parse public key: %w", err).Error())
	}
	if !crypt.VerifyPSSWithTimestamp(rsaPublicKey, req.Signature, req.Timestamp, req.PublicKeyBytes) {
		return nil, status.Error(codes.InvalidArgument, "invalid signature")
	}
	userID, err := s.registerUserFunc(ctx, req.PublicKeyBytes)
	if err != nil {
		if errors.Is(err, service.ErrConflict) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to register: %w", err).Error())
	}
	return &proto.AuthResponse{
		UserId:    userID,
		StartTime: startTime,
		EndTime:   timestamppb.Now(),
	}, nil
}
