package secret

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
)

func (s *ServiceServer) findAndAuthUser(ctx context.Context, req *proto.Auth, data ...[]byte) (*service.User, error) {
	if len(req.Signature) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Signature is required")
	}
	currentTime := time.Now().Unix()
	if req.Timestamp < currentTime-s.timeWindow || req.Timestamp > currentTime+s.timeWindow {
		return nil, status.Error(codes.InvalidArgument, "Timestamp is out of range")
	}
	user, err := s.user.Find(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, fmt.Errorf("failed to find user: %w", err).Error())
	}
	publicKey, err := crypt.RSAPublicKeyFromBytes(user.PublicKeyBytes)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to parse public key: %w", err).Error())
	}
	if !crypt.VerifyPSSWithTimestampAndUserID(publicKey, req.Signature, req.UserId, req.Timestamp,
		append(data, []byte(req.Id), user.PublicKeyBytes)...) {
		return nil, status.Error(codes.Unauthenticated, "invalid signature")
	}
	return user, nil
}
