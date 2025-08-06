package grpcclient

import (
	"context"
	"fmt"
	"gophkeeper/internal/client/service"
	"gophkeeper/pkg/crypt"
	"gophkeeper/pkg/proto"
	"time"

	"google.golang.org/grpc"
)

// AuthClient реализует клиентскую логику аутентификации.
type AuthClient struct {
	c    proto.AuthServiceClient
	key  *service.Key
	time *service.Time
}

// NewAuthClient создаёт новый клиент.
func NewAuthClient(conn *grpc.ClientConn, key *service.Key, t *service.Time) *AuthClient {
	return &AuthClient{
		c:    proto.NewAuthServiceClient(conn),
		key:  key,
		time: t,
	}
}

// Login получает userID.
func (s *AuthClient) Login(ctx context.Context) error {
	timestamp := s.time.Current().Unix()
	signature, err := crypt.SignPSSWithTimestamp(s.key.PrivateKey, timestamp, s.key.PublicKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to sign data: %w", err)
	}
	startTime := time.Now()
	r, err := s.c.Login(ctx, &proto.LoginRequest{
		Fingerprint: s.key.Fingerprint,
		Timestamp:   timestamp,
		Signature:   signature,
	})
	endTime := time.Now()
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	s.time.SetDiff((r.StartTime.AsTime().Sub(startTime) + r.EndTime.AsTime().Sub(endTime)) / 2)
	s.key.SetUserID(r.UserId)
	return nil
}

// Register регистрирует новый публичный ключ на сервере.
func (s *AuthClient) Register(ctx context.Context) error {
	timestamp := s.time.Current().Unix()
	signature, err := crypt.SignPSSWithTimestamp(s.key.PrivateKey, timestamp, s.key.PublicKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to sign data: %w", err)
	}
	startTime := time.Now()
	r, err := s.c.Register(ctx, &proto.RegisterRequest{
		PublicKeyBytes: s.key.PublicKeyBytes,
		Timestamp:      timestamp,
		Signature:      signature,
	})
	endTime := time.Now()
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	s.time.SetDiff((r.StartTime.AsTime().Sub(startTime) + r.EndTime.AsTime().Sub(endTime)) / 2)
	s.key.SetUserID(r.UserId)
	return nil
}
