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
	auth *service.Auth
	time *service.Time
}

// NewAuthClient создаёт новый клиент.
func NewAuthClient(conn *grpc.ClientConn, s *service.Auth, t *service.Time) *AuthClient {
	return &AuthClient{
		c:    proto.NewAuthServiceClient(conn),
		auth: s,
		time: t,
	}
}

// Login получает userID.
func (c *AuthClient) Login(ctx context.Context) error {
	timestamp := c.time.Current().Unix()
	signature, err := crypt.SignPSSWithTimestamp(c.auth.PrivateKey, timestamp, c.auth.PublicKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to sign data: %w", err)
	}
	startTime := time.Now()
	r, err := c.c.Login(ctx, &proto.LoginRequest{
		Fingerprint: c.auth.Fingerprint,
		Timestamp:   timestamp,
		Signature:   signature,
	})
	endTime := time.Now()
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	c.time.SetDiff((r.StartTime.AsTime().Sub(startTime) + r.EndTime.AsTime().Sub(endTime)) / 2)
	c.auth.SetUserID(r.UserId)
	return nil
}

// Register регистрирует новый публичный ключ на сервере.
func (c *AuthClient) Register(ctx context.Context) error {
	timestamp := c.time.Current().Unix()
	signature, err := crypt.SignPSSWithTimestamp(c.auth.PrivateKey, timestamp, c.auth.PublicKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to sign data: %w", err)
	}
	startTime := time.Now()
	r, err := c.c.Register(ctx, &proto.RegisterRequest{
		PublicKeyBytes: c.auth.PublicKeyBytes,
		Timestamp:      timestamp,
		Signature:      signature,
	})
	endTime := time.Now()
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	c.time.SetDiff((r.StartTime.AsTime().Sub(startTime) + r.EndTime.AsTime().Sub(endTime)) / 2)
	c.auth.SetUserID(r.UserId)
	return nil
}
