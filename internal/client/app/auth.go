package app

import (
	"context"
	"fmt"
	"gophkeeper/pkg/crypt"
	"gophkeeper/pkg/proto"
	"time"
)

type authRoute string

const (
	routeLogin    authRoute = "login"
	routeRegister authRoute = "register"
)

func (a *App) authByRoute(ctx context.Context, route authRoute) error {
	var f func(timestamp int64, signature []byte) (*proto.AuthResponse, error)
	switch route {
	case routeLogin:
		f = func(timestamp int64, signature []byte) (*proto.AuthResponse, error) {
			return a.auth.Login(ctx, &proto.LoginRequest{
				Fingerprint: a.key.Fingerprint,
				Timestamp:   timestamp,
				Signature:   signature,
			})
		}
	case routeRegister:
		f = func(timestamp int64, signature []byte) (*proto.AuthResponse, error) {
			return a.auth.Register(ctx, &proto.RegisterRequest{
				PublicKeyBytes: a.key.PublicKeyBytes,
				Timestamp:      timestamp,
				Signature:      signature,
			})
		}
	default:
		return fmt.Errorf("unknown route: %s", route)
	}
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
	go a.run(ctx)
	return nil
}

func (a *App) Login(ctx context.Context) error {
	if err := a.authByRoute(ctx, routeLogin); err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	return nil
}

func (a *App) Register(ctx context.Context) error {
	if err := a.authByRoute(ctx, routeRegister); err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}
	return nil
}
