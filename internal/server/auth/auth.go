package auth

import (
	"gophkeeper/internal/server/service"
	"gophkeeper/pkg/proto"
)

// ServiceServer реализует gRPC AuthServiceServer.
type ServiceServer struct {
	proto.UnimplementedAuthServiceServer
	registerUserFunc service.RegisterUserFunc
	findUserFunc     service.FindUserByFingerprintFunc
	// timeWindow временное окно в секундах для предотвращения атак повторного воспроизведения
	timeWindow int64
}

// defaultTimeWindow временное окно по умолчанию.
const defaultTimeWindow = 600

// NewServiceServer создаёт новый ServiceServer.
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

// WithRegisterUserFunc устанавливает функцию регистрации пользователя.
func WithRegisterUserFunc(fn service.RegisterUserFunc) Option {
	return func(s *ServiceServer) {
		s.registerUserFunc = fn
	}
}

// WithFindUserFunc устанавливает функцию поиска пользователя.
func WithFindUserFunc(fn service.FindUserByFingerprintFunc) Option {
	return func(s *ServiceServer) {
		s.findUserFunc = fn
	}
}
