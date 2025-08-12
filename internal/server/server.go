package server

import (
	"context"
	"errors"
	"fmt"
	"gophkeeper/internal/server/auth"
	"gophkeeper/internal/server/config"
	"gophkeeper/internal/server/infra/db"
	"gophkeeper/internal/server/interceptors/logger"
	"gophkeeper/internal/server/secret"
	"gophkeeper/internal/server/service"
	"gophkeeper/pkg/logging"
	pb "gophkeeper/pkg/proto"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func Serve(ctx context.Context, cfg *config.Config, store *db.Store, l *logging.ZapLogger) error {
	authServiceServer := auth.NewServiceServer(
		auth.WithRegisterUserFunc(service.NewRegisterUserFunc(store)),
		auth.WithFindUserFunc(service.NewFindUserByFingerprintFunc(store)),
	)
	syncService := service.NewSyncService()
	secretServiceServer := secret.NewServiceServer(
		secret.WithSyncService(syncService),
		secret.WithFindUserFunc(service.NewFindUserByIDFunc(store)),
		secret.WithFindSecretFunc(service.NewFindSecretFunc(store)),
		secret.WithListSecretsFunc(service.NewListSecretsFunc(store)),
		secret.WithSaveSecretFunc(service.NewSaveSecretFunc(store)),
		secret.WithDeleteSecretFunc(service.NewDeleteSecretFunc(store)),
	)

	var opts []grpc.ServerOption
	if cfg.CertPath != "" && cfg.KeyPath != "" {
		tlsCreds, err := credentials.NewServerTLSFromFile(cfg.CertPath, cfg.KeyPath)
		if err != nil {
			return fmt.Errorf("failed to create tls credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(tlsCreds))
	}
	opts = append(opts, grpc.UnaryInterceptor(logger.Interceptor(l)))
	s := grpc.NewServer(opts...)
	pb.RegisterAuthServiceServer(s, authServiceServer)
	pb.RegisterSecretServiceServer(s, secretServiceServer)

	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen gRPC server: %w", err)
	}
	go func() {
		<-ctx.Done()
		l.InfoCtx(ctx, "Shutting down the gRPC server...")
		syncService.Close()
		s.GracefulStop()
	}()
	if err = s.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("failed serve gRPC server: %w", err)
	}
	return nil
}
