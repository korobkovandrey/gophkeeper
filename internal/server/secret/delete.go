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

// Delete удаляет секрет.
func (s *ServiceServer) Delete(ctx context.Context, req *proto.SecretDeleteRequest) (*proto.Secret, error) {
	b := make([]byte, crypt.Int64BytesLen)
	binary.LittleEndian.PutUint64(b, uint64(req.EventTime))
	user, err := s.findAndAuthUser(ctx, req.Auth, b)
	if err != nil {
		return nil, err
	}
	secret, err := s.deleteSecretFunc(ctx, user.ID, req.Auth.Id, time.Unix(req.EventTime, 0))
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
