package server

import (
	"gophkeeper/internal/server/auth"
	"gophkeeper/internal/server/infra/db"
	"gophkeeper/internal/server/secret"
	"gophkeeper/internal/server/service"
)

func MakeAuthServiceServer(store *db.Store) *auth.ServiceServer {
	return auth.NewServiceServer(
		auth.WithRegisterUserFunc(service.NewRegisterUserFunc(store)),
		auth.WithFindUserFunc(service.NewFindUserByFingerprintFunc(store)),
	)
}

func MakeSecretServiceServer(syncService *service.SyncService, store *db.Store) *secret.ServiceServer {
	return secret.NewServiceServer(
		secret.WithSyncService(syncService),
		secret.WithFindUserFunc(service.NewFindUserByIDFunc(store)),
		secret.WithFindSecretFunc(service.NewFindSecretFunc(store)),
		secret.WithListSecretsFunc(service.NewListSecretsFunc(store)),
		secret.WithSaveSecretFunc(service.NewSaveSecretFunc(store)),
		secret.WithDeleteSecretFunc(service.NewDeleteSecretFunc(store)),
	)
}
