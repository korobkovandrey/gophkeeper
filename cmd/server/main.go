package main

import (
	"context"
	"fmt"
	"gophkeeper/internal/server"
	"gophkeeper/internal/server/config"
	"gophkeeper/internal/server/infra/db"
	"gophkeeper/pkg/logging"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	l.InfoCtx(ctx, "", zap.Any("config", cfg))
	store, err := db.MakeStoreConnectAndMigrate(ctx, cfg.DSN)
	if err != nil {
		l.FatalCtx(ctx, "failed to connect to database", zap.Error(err))
	}
	defer func() {
		if closeErr := store.Close(); closeErr != nil {
			l.ErrorCtx(ctx, "failed to close store", zap.Error(closeErr))
		}
	}()
	if err = server.Serve(ctx, cfg, store, l); err != nil {
		l.ErrorCtx(ctx, "failed to serve gRPC server", zap.Error(err))
	} else if cfg.IsWrite() {
		if err = cfg.WriteConfig(); err != nil {
			l.ErrorCtx(ctx, "failed to write config", zap.Error(err))
		} else {
			l.InfoCtx(ctx, "config written to "+cfg.ConfigPath())
		}
	}
}
