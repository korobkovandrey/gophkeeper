package main

import (
	"context"
	"fmt"
	"gophkeeper/internal/client/app"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/service"
	"gophkeeper/internal/client/tui"
	"gophkeeper/pkg/logging"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version=%v, date=%v, commit=%v\n", buildVersion, buildDate, buildCommit)
	_ = godotenv.Load()
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	l, err := logging.NewZapLogger(zapcore.Level(cfg.LogLevel), cfg.LogOutputs)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Sync()
	ctx := context.Background()
	l.InfoCtx(ctx, "", zap.Any("config", cfg))

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
	a := app.NewApp(l, conn, t, key, store)
	go a.RunSync(ctx)
	storage := service.NewStorage(key, store)

	modelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	p := tea.NewProgram(tui.NewModel(modelCtx, keyManager, a, storage), tea.WithContext(ctx), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		l.ErrorCtx(ctx, "error program", zap.Error(err))
	} else if cfg.IsWrite() {
		if err = cfg.WriteConfig(); err != nil {
			l.ErrorCtx(ctx, "failed to write config", zap.Error(err))
		} else {
			l.InfoCtx(ctx, "config written to "+cfg.ConfigPath())
		}
	}
}
