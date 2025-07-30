package secret

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"gophkeeper/internal/server/service"
	"gophkeeper/pkg/proto"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userService interface {
	Find(ctx context.Context, id int64) (*service.User, error)
}

type secretService interface {
	Find(ctx context.Context, userID int64, id string) (*service.Secret, error)
	List(ctx context.Context, userID int64) ([]*service.Secret, error)
	Store(ctx context.Context, userID int64, id, newID string, secret, meta, data []byte, eventTime time.Time) (*service.Secret, error)
	Delete(ctx context.Context, userID int64, id string, eventTime time.Time) (*service.Secret, error)
}

type ServiceServer struct {
	proto.UnimplementedSecretServiceServer
	user       userService
	secret     secretService
	sync       *service.SyncService
	timeWindow int64
}

// defaultTimeWindow временное окно по умолчанию.
const defaultTimeWindow = 60

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

// WithUserService устанавливает user service.
func WithUserService(us userService) Option {
	return func(s *ServiceServer) {
		s.user = us
	}
}

// WithSyncService устанавливает sync service.
func WithSyncService(ss *service.SyncService) Option {
	return func(s *ServiceServer) {
		s.sync = ss
	}
}

// WithSecretService устанавливает secret service.
func WithSecretService(ss secretService) Option {
	return func(s *ServiceServer) {
		s.secret = ss
	}
}

// Store сохраняет секрет.
func (s *ServiceServer) Store(ctx context.Context, req *proto.SecretStoreRequest) (*proto.Secret, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(req.EventTime))
	user, err := s.findAndAuthUser(ctx, req.Auth, []byte(req.Auth.Id), []byte(req.Id),
		req.Crypt, req.Meta, req.Data, b)
	if err != nil {
		return nil, err
	}
	secret, err := s.secret.Store(ctx, user.ID, req.Auth.Id,
		req.Id, req.Crypt, req.Meta, req.Data, time.Unix(req.EventTime, 0))
	if err != nil {
		code := codes.Internal
		if errors.Is(err, service.ErrConflict) {
			code = codes.AlreadyExists
		}
		return nil, status.Error(code, fmt.Errorf("failed to store secret %s: %w", req.Auth.Id, err).Error())
	}
	go s.sync.Broadcast(int32(proto.SecretAction_STORED), req.Auth.Id, *secret)
	return &proto.Secret{
		Id:        secret.ID,
		Crypt:     secret.Crypt,
		Meta:      secret.Meta,
		Data:      secret.Data,
		CreatedAt: secret.CreatedAt.Unix(),
		UpdatedAt: secret.UpdatedAt.Unix(),
	}, nil
}

// Delete удаляет секрет.
func (s *ServiceServer) Delete(ctx context.Context, req *proto.SecretDeleteRequest) (*proto.Secret, error) {
	b := make([]byte, 8)
	binary.LittleEndian.AppendUint64(b, uint64(req.EventTime))
	user, err := s.findAndAuthUser(ctx, req.Auth, b)
	if err != nil {
		return nil, err
	}
	secret, err := s.secret.Delete(ctx, user.ID, req.Auth.Id, time.Unix(req.EventTime, 0))
	if err != nil {
		code := codes.Internal
		if errors.Is(err, service.ErrConflict) {
			code = codes.FailedPrecondition
		}
		return nil, status.Error(code, fmt.Errorf("failed to delete secret: %w", err).Error())
	}
	go s.sync.Broadcast(int32(proto.SecretAction_DELETED), req.Auth.Id, *secret)
	return &proto.Secret{
		Id:        secret.ID,
		Crypt:     secret.Crypt,
		Meta:      secret.Meta,
		Data:      secret.Data,
		CreatedAt: secret.CreatedAt.Unix(),
		UpdatedAt: secret.UpdatedAt.Unix(),
	}, nil
}
