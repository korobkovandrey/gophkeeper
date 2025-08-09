package app

import (
	"context"
	"fmt"
	"gophkeeper/pkg/crypt"
	"gophkeeper/pkg/proto"
	"time"
)

func (a *App) authorize(f func(timestamp int64, signature []byte) (*proto.AuthResponse, error)) error {
	timestamp := time.Now().Unix()
	signature, err := crypt.SignPSSWithTimestamp(a.key.PrivateKey, timestamp, a.key.PublicKeyBytes)
	if err != nil {
		return fmt.Errorf("failed to sign data: %w", err)
	}
	startTime := time.Now()
	r, err := f(timestamp, signature)
	endTime := time.Now()
	if err != nil {
		return err
	}
	a.time.SetDiff((r.StartTime.AsTime().Sub(startTime) + r.EndTime.AsTime().Sub(endTime)) / 2)
	a.key.SetUserID(r.UserId)
	return nil
}

func (a *App) Login(ctx context.Context) error {
	err := a.authorize(func(timestamp int64, signature []byte) (*proto.AuthResponse, error) {
		return a.auth.Login(ctx, &proto.LoginRequest{
			Fingerprint: a.key.Fingerprint,
			Timestamp:   timestamp,
			Signature:   signature,
		})
	})
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	go a.run(ctx)
	return nil
}

func (a *App) Register(ctx context.Context) error {
	err := a.authorize(func(timestamp int64, signature []byte) (*proto.AuthResponse, error) {
		return a.auth.Register(ctx, &proto.RegisterRequest{
			PublicKeyBytes: a.key.PublicKeyBytes,
			Timestamp:      timestamp,
			Signature:      signature,
		})
	})
	if err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}
	go a.run(ctx)
	return nil
}
