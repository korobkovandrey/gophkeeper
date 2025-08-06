package grpcclient

import (
	"context"
	"fmt"
	"gophkeeper/internal/client/service"
	"gophkeeper/pkg/crypt"
	"gophkeeper/pkg/proto"
	"io"

	"google.golang.org/grpc"
)

// SecretClient реализует клиентскую логику работы с секретами.
type SecretClient struct {
	c      proto.SecretServiceClient
	key    *service.Key
	time   *service.Time
	online bool
}

func NewSecretClient(conn *grpc.ClientConn, key *service.Key, t *service.Time) *SecretClient {
	return &SecretClient{
		c:    proto.NewSecretServiceClient(conn),
		key:  key,
		time: t,
	}
}

func (s *SecretClient) Stream(ctx context.Context) error {
	auth, err := s.makeSignAuth("")
	if err != nil {
		return fmt.Errorf("failed to create signed Auth: %w", err)
	}
	stream, err := s.c.Stream(ctx, auth)
	defer func() {
		s.online = false
	}()
	if err != nil {
		return err
	}
	defer func() {
		_ = stream.CloseSend()
	}()
	s.online = true
	for {
		msg, errRecv := stream.Recv()
		if errRecv != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to receive message: %w", errRecv)
		}
		fmt.Println(msg.Action, msg.Id, msg.Secret)
	}
	return nil
}

func (s *SecretClient) Sync(ctx context.Context) error {
	//_, err := s.c.Store(ctx, secret)
	return nil
}

func (s *SecretClient) IsOnline() bool {
	return s.online
}

func (s *SecretClient) makeSignAuth(id string, data ...[]byte) (auth *proto.Auth, err error) {
	auth = &proto.Auth{
		Id:        id,
		UserId:    s.key.UserID,
		Timestamp: s.time.Current().Unix(),
	}
	auth.Signature, err = crypt.SignPSSWithTimestampAndUserID(s.key.PrivateKey, auth.UserId, auth.Timestamp,
		append(data, []byte(auth.Id), s.key.PublicKeyBytes)...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature: %w", err)
	}
	return auth, nil
}
