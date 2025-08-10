package main

import (
	"context"
	"errors"
	"gophkeeper/internal/server/auth"
	"gophkeeper/internal/server/config"
	"gophkeeper/internal/server/infra/db"
	"gophkeeper/internal/server/interceptors/logger"
	"gophkeeper/internal/server/secret"
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

	userService := service.NewUserService(store)
	syncService := service.NewSyncService()
	secretService := service.NewSecretService(store)
	authServiceServer := auth.NewServiceServer(auth.WithUserService(userService))
	secretServiceServer := secret.NewServiceServer(
		secret.WithUserService(userService),
		secret.WithSyncService(syncService),
		secret.WithSecretService(secretService),
	)
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
		l.ErrorCtx(ctx, "failed to start gRPC server", zap.Error(err))
	}
}
