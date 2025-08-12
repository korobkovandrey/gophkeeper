package main

import (
	"context"
	"errors"
	"fmt"
	"gophkeeper/internal/server"
	"gophkeeper/internal/server/config"
	"gophkeeper/internal/server/infra/db"
	"gophkeeper/internal/server/interceptors/logger"
	"gophkeeper/internal/server/service"
	"gophkeeper/pkg/logging"
	pb "gophkeeper/pkg/proto"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

	if err := db.Migrate(cfg.DSN); err != nil {
		l.FatalCtx(ctx, "failed to migrate database", zap.Error(err))
	}
	dbConnect, err := db.Connect(ctx, cfg.DSN)
	if err != nil {
		l.FatalCtx(ctx, "failed to connect to database", zap.Error(err))
	}
	defer func() {
		if closeErr := dbConnect.Close(); closeErr != nil {
			l.ErrorCtx(ctx, "failed to close database connection", zap.Error(closeErr))
		}
	}()
	store := db.NewStore(dbConnect)

	authServiceServer := server.MakeAuthServiceServer(store)
	syncService := service.NewSyncService()
	secretServiceServer := server.MakeSecretServiceServer(syncService, store)

	var opts []grpc.ServerOption
	if cfg.CertPath != "" && cfg.KeyPath != "" {
		tlsCreds, err := credentials.NewServerTLSFromFile(cfg.CertPath, cfg.KeyPath)
		if err != nil {
			l.FatalCtx(ctx, "failed to create tls credentials", zap.Error(err))
		}
		opts = append(opts, grpc.Creds(tlsCreds))
	}
	serv, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		l.FatalCtx(ctx, "failed to listen gRPC server", zap.Error(err))
	}
	opts = append(opts, grpc.UnaryInterceptor(logger.Interceptor(l)))
	s := grpc.NewServer(opts...)
	pb.RegisterAuthServiceServer(s, authServiceServer)
	pb.RegisterSecretServiceServer(s, secretServiceServer)
	go func() {
		<-ctx.Done()
		l.InfoCtx(ctx, "Shutting down the gRPC server...")
		syncService.Close()
		s.GracefulStop()
	}()

	if err = s.Serve(serv); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		l.ErrorCtx(ctx, "failed serve gRPC server", zap.Error(err))
	} else if cfg.IsWrite() {
		if err = cfg.WriteConfig(); err != nil {
			l.ErrorCtx(ctx, "failed to write config", zap.Error(err))
		} else {
			l.InfoCtx(ctx, "config written to "+cfg.ConfigPath())
		}
	}
}
