package secret

import (
	"gophkeeper/internal/server/service"
	"gophkeeper/pkg/proto"
)

type ServiceServer struct {
	proto.UnimplementedSecretServiceServer
	findUserFunc     service.FindUserByIDFunc
	findSecretFunc   service.FindSecretFunc
	listSecretsFunc  service.ListSecretsFunc
	saveSecretFunc   service.SaveSecretFunc
	deleteSecretFunc service.DeleteSecretFunc
	sync             *service.SyncService
	timeWindow       int64
}

// defaultTimeWindow временное окно по умолчанию.
const defaultTimeWindow = 60

func NewServiceServer(opts ...Option) *ServiceServer {
	s := &ServiceServer{
		timeWindow: defaultTimeWindow,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option опция для конфигурации ServiceServer.
type Option func(*ServiceServer)

// WithTimeWindow устанавливает временное окно.
func WithTimeWindow(tw int64) Option {
	return func(s *ServiceServer) {
		s.timeWindow = tw
	}
}

// WithFindUserFunc устанавливает функцию поиска пользователя.
func WithFindUserFunc(fn service.FindUserByIDFunc) Option {
	return func(s *ServiceServer) {
		s.findUserFunc = fn
	}
}

// WithFindSecretFunc устанавливает функцию поиска секрета.
func WithFindSecretFunc(fn service.FindSecretFunc) Option {
	return func(s *ServiceServer) {
		s.findSecretFunc = fn
	}
}

// WithListSecretsFunc устанавливает функцию получения списка секретов.
func WithListSecretsFunc(fn service.ListSecretsFunc) Option {
	return func(s *ServiceServer) {
		s.listSecretsFunc = fn
	}
}

// WithSaveSecretFunc устанавливает функцию сохранения секрета.
func WithSaveSecretFunc(fn service.SaveSecretFunc) Option {
	return func(s *ServiceServer) {
		s.saveSecretFunc = fn
	}
}

// WithDeleteSecretFunc устанавливает функцию удаления секрета.
func WithDeleteSecretFunc(fn service.DeleteSecretFunc) Option {
	return func(s *ServiceServer) {
		s.deleteSecretFunc = fn
	}
}

// WithSyncService устанавливает сервис синхронизации.
func WithSyncService(ss *service.SyncService) Option {
	return func(s *ServiceServer) {
		s.sync = ss
	}
}
