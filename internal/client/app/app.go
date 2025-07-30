package app

import (
	"fmt"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/grpcclient"
	"gophkeeper/internal/client/service"

	"google.golang.org/grpc"
)

type App struct {
	conn *grpc.ClientConn
	cfg  *config.Config
	time *service.Time
	auth *service.Auth
	*grpcclient.AuthClient
	*grpcclient.SecretClient

	OnOnline  func()
	OnOffline func()
	OnMsg     func()
}

func NewApp(cfg *config.Config, conn *grpc.ClientConn) (*App, error) {
	a := &App{
		conn: conn,
		cfg:  cfg,
		time: service.NewTime(),
		auth: service.NewAuth(),
	}
	a.AuthClient = grpcclient.NewAuthClient(a.conn, a.auth, a.time)
	if a.cfg.PrivateKeyPath != "" {
		if err := a.auth.SetPrivateKeyFromPath(a.cfg.PrivateKeyPath); err != nil {
			return nil, fmt.Errorf("failed to set private key: %w", err)
		}
	}
	a.SecretClient = grpcclient.NewSecretClient(a.conn)
	return a, nil
}
