package main

import (
	"context"
	"gophkeeper/internal/client/app"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/tui"
	"gophkeeper/pkg/logging"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
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

	ctx, cancel := context.WithCancel(ctx)
	a, err := app.NewApp(cfg, conn)
	if err != nil {
		l.FatalCtx(ctx, "failed to create app", zap.Error(err))
	}
	p := tea.NewProgram(tui.NewModel(ctx, a), tea.WithContext(ctx), tea.WithAltScreen())
	go func() {
		time.Sleep(5 * time.Second)
		p.Send(tui.NewChangeSecretsMsg())
	}()
	if _, err := p.Run(); err != nil {
		l.FatalCtx(ctx, "error starting program", zap.Error(err))
	}
	cancel()
}
