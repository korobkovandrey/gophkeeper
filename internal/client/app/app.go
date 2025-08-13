package app

import (
	"context"
	"gophkeeper/internal/client/service"
	"gophkeeper/pkg/logging"
	"gophkeeper/pkg/proto"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type App struct {
	auth     proto.AuthServiceClient
	secret   proto.SecretServiceClient
	l        *logging.ZapLogger
	time     *service.Time
	key      *service.Key
	store    *service.MemStore
	needSync atomic.Bool
	online   atomic.Bool
	OnlineCh chan struct{}
	UpdateCh chan struct{}
	eventCh  chan *proto.SecretEvent
}

// NewApp создает новое приложение.
func NewApp(conn *grpc.ClientConn, opts ...Option) *App {
	a := &App{
		auth:     proto.NewAuthServiceClient(conn),
		secret:   proto.NewSecretServiceClient(conn),
		OnlineCh: make(chan struct{}, 1),
		UpdateCh: make(chan struct{}, 1),
		eventCh:  make(chan *proto.SecretEvent),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Option опция для конфигурации App.
type Option func(*App)

// WithLogger устанавливает логгер.
func WithLogger(l *logging.ZapLogger) Option {
	return func(a *App) {
		a.l = l
	}
}

// WithTimeService устанавливает сервис времени.
func WithTimeService(t *service.Time) Option {
	return func(a *App) {
		a.time = t
	}
}

// WithKeyService устанавливает сервис ключей.
func WithKeyService(key *service.Key) Option {
	return func(a *App) {
		a.key = key
	}
}

// WithStore устанавливает хранилище.
func WithStore(store *service.MemStore) Option {
	return func(a *App) {
		a.store = store
	}
}

func (a *App) IsLogged() bool {
	return a.key.UserID > 0
}

func (a *App) run(ctx context.Context) {
	const maxBackoff = 5 * time.Second
	backoff := time.Second
	a.runSync(ctx)
	started := false
	defer func() {
		close(a.OnlineCh)
		close(a.UpdateCh)
		close(a.eventCh)
	}()
	for {
		var stream grpc.ServerStreamingClient[proto.SecretEvent]
		auth, err := a.makeSignAuth("")
		if err != nil {
			a.l.ErrorCtx(ctx, "failed to create signed Auth", zap.Error(err))
		} else {
			stream, err = a.secret.Stream(ctx, auth)
			if err != nil {
				a.l.ErrorCtx(ctx, "failed to create stream", zap.Error(err))
			}
		}
		if err != nil {
			a.onlineEvent(false)
			time.Sleep(backoff)
			backoff = min(backoff+time.Second, maxBackoff)
			continue
		}
		a.onlineEvent(true)
		backoff = time.Second
		if started {
			go a.runSync(ctx)
		} else {
			started = true
		}
		for {
			if event, err := stream.Recv(); err != nil {
				a.onlineEvent(false)
				a.l.InfoCtx(ctx, "stream error", zap.Error(err))
				if status.Code(err) == codes.Canceled {
					return
				}
				break
			} else {
				a.l.InfoCtx(ctx, "stream", zap.Any("event", event))
				a.eventCh <- event
			}
		}
	}
}

func (a *App) IsOnline() bool {
	return a.online.Load()
}

func (a *App) onlineEvent(online bool) {
	a.online.Store(online)
	select {
	case a.OnlineCh <- struct{}{}:
	default:
	}
}

func (a *App) updateEvent() {
	select {
	case a.UpdateCh <- struct{}{}:
	default:
	}
}
