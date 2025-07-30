package grpcclient

import (
	"context"
	"fmt"
	"gophkeeper/pkg/proto"
	"io"

	"google.golang.org/grpc"
)

// SecretClient реализует клиентскую логику работы с секретами.
type SecretClient struct {
	c      proto.SecretServiceClient
	online bool
}

func NewSecretClient(conn *grpc.ClientConn) *SecretClient {
	return &SecretClient{
		c: proto.NewSecretServiceClient(conn),
	}
}

func (s *SecretClient) Stream(ctx context.Context) error {
	stream, err := s.c.Stream(ctx, &proto.Auth{UserId: 1})
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
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				fmt.Println(err)
				break
			}
			return err
		}
		fmt.Println(msg.Action, msg.Secret)
	}
	return nil
}

func (s *SecretClient) Sync(ctx context.Context) error {
	//_, err := c.c.CreateSecret(ctx, secret)
	return nil
}

func (s *SecretClient) IsOnline() bool {
	return s.online
}
