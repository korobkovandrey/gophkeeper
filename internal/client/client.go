package client

import (
	"context"
	"gophkeeper/internal/client/app"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/service"
	"gophkeeper/internal/client/tui"
	"gophkeeper/pkg/logging"

	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func Run(ctx context.Context, cfg *config.Config, l *logging.ZapLogger) error {
	key := service.NewKey()
	if cfg.PrivateKeyPath != "" {
		if err := key.SetPrivateKeyFromPath(cfg.PrivateKeyPath); err != nil {
			l.FatalCtx(ctx, "failed to set private key", zap.Error(err))
		}
	}

	var dialOpts []grpc.DialOption
	if cfg.CAPath == "" {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsCreds, err := credentials.NewClientTLSFromFile(cfg.CAPath, "localhost")
		if err != nil {
			l.FatalCtx(ctx, "failed to create tls credentials", zap.Error(err))
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(tlsCreds))
	}
	conn, err := grpc.NewClient(
		cfg.Addr,
		dialOpts...,
	)
	if err != nil {
		l.FatalCtx(ctx, "failed to create grpc client", zap.Error(err))
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			l.ErrorCtx(ctx, "failed to close grpc client", zap.Error(closeErr))
		}
	}()

	t := service.NewTime()
	keyManager := app.NewKeyManager(cfg, key)
	store := service.NewMemStore()
	defer store.Close()
	a := app.NewApp(conn, app.WithLogger(l), app.WithTimeService(t), app.WithKeyService(key), app.WithStore(store))
	go a.RunSync(ctx)
	storage := service.NewStorage(key, store)

	modelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	p := tea.NewProgram(tui.NewModel(modelCtx, keyManager, a, storage), tea.WithContext(ctx), tea.WithAltScreen())
	_, err = p.Run()
	return err
}
