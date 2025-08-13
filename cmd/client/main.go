package main

import (
	"context"
	"fmt"
	"gophkeeper/internal/client"
	"gophkeeper/internal/client/config"
	"gophkeeper/pkg/logging"
	"log"

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
	ctx := context.Background()
	l.InfoCtx(ctx, "", zap.Any("config", cfg))
	if err := client.Run(ctx, cfg, l); err != nil {
		l.ErrorCtx(ctx, err.Error())
	} else if cfg.IsWrite() {
		if err = cfg.WriteConfig(); err != nil {
			l.ErrorCtx(ctx, "failed to write config", zap.Error(err))
		} else {
			l.InfoCtx(ctx, "config written to "+cfg.ConfigPath())
		}
	}
}
