package secret

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"gophkeeper/internal/server/service"
	"gophkeeper/pkg/crypt"
	"gophkeeper/pkg/proto"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Store сохраняет секрет.
func (s *ServiceServer) Store(ctx context.Context, req *proto.SecretStoreRequest) (*proto.Secret, error) {
	b := make([]byte, crypt.Int64BytesLen)
	binary.LittleEndian.PutUint64(b, uint64(req.EventTime))
	user, err := s.findAndAuthUser(ctx, req.Auth, []byte(req.Id), req.Crypt, req.Meta, req.Data, b)
	if err != nil {
		return nil, err
	}
	secret, err := s.saveSecretFunc(ctx, user.ID, req.Auth.Id,
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
