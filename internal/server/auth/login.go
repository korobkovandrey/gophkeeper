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

// Login проверяет подпись пользователя.
func (s *ServiceServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.AuthResponse, error) {
	startTime := timestamppb.Now()
	if req.Fingerprint == "" {
		return nil, status.Error(codes.InvalidArgument, "Fingerprint is required")
	}
	if len(req.Signature) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Signature is required")
	}
	currentTime := time.Now().Unix()
	if req.Timestamp < currentTime-s.timeWindow || req.Timestamp > currentTime+s.timeWindow {
		return nil, status.Error(codes.InvalidArgument, "Timestamp is out of range")
	}
	user, err := s.findUserFunc(ctx, req.Fingerprint)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to get user by fingerprint: %w", err).Error())
	}
	rsaPublicKey, err := crypt.RSAPublicKeyFromBytes(user.PublicKeyBytes)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to parse public key: %w", err).Error())
	}
	if !crypt.VerifyPSSWithTimestamp(rsaPublicKey, req.Signature, req.Timestamp, user.PublicKeyBytes) {
		return nil, status.Error(codes.InvalidArgument, "invalid signature")
	}
	return &proto.AuthResponse{
		UserId:    user.ID,
		StartTime: startTime,
		EndTime:   timestamppb.Now(),
	}, nil
}
