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

// userService представляет интерфейс для регистрации и поиска пользователей.
type userService interface {
	Register(ctx context.Context, publicKeyBytes []byte) (id int64, err error)
	FindByFingerprint(ctx context.Context, fingerprint string) (user *service.User, err error)
}

// ServiceServer реализует gRPC AuthServiceServer.
type ServiceServer struct {
	proto.UnimplementedAuthServiceServer
	user userService
	// timeWindow временное окно в секундах для предотвращения атак повторного воспроизведения
	timeWindow int64
}

// defaultTimeWindow временное окно по умолчанию.
const defaultTimeWindow = 60

// NewServiceServer создаёт новый ServiceServer.
func NewServiceServer(opts ...Option) *ServiceServer {
	s := &ServiceServer{
		timeWindow: defaultTimeWindow,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option опция для конфигурации ServiceServer.
type Option func(*ServiceServer)

// WithTimeWindow устанавливает временное окно.
func WithTimeWindow(tw int64) Option {
	return func(s *ServiceServer) {
		s.timeWindow = tw
	}
}

// WithUserService устанавливает user.
func WithUserService(us userService) Option {
	return func(s *ServiceServer) {
		s.user = us
	}
}

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
	userID, err := s.user.Register(ctx, req.PublicKeyBytes)
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

// Login аутентифицирует пользователя по подписи.
func (s *ServiceServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.AuthResponse, error) {
	startTime := timestamppb.Now()
	if len(req.Fingerprint) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Fingerprint is required")
	}
	if len(req.Signature) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Signature is required")
	}
	currentTime := time.Now().Unix()
	if req.Timestamp < currentTime-s.timeWindow || req.Timestamp > currentTime+s.timeWindow {
		return nil, status.Error(codes.InvalidArgument, "Timestamp is out of range")
	}
	user, err := s.user.FindByFingerprint(ctx, req.Fingerprint)
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
