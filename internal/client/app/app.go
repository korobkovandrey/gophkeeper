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
}

func NewApp(l *logging.ZapLogger, conn *grpc.ClientConn, t *service.Time, key *service.Key, store *service.MemStore) *App {
	a := &App{
		auth:     proto.NewAuthServiceClient(conn),
		secret:   proto.NewSecretServiceClient(conn),
		l:        l,
		time:     t,
		key:      key,
		store:    store,
		OnlineCh: make(chan struct{}, 1),
		UpdateCh: make(chan struct{}, 1),
	}
	return a
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
			event, err := stream.Recv()
			if err != nil {
				a.onlineEvent(false)
				a.l.InfoCtx(ctx, "stream error", zap.Error(err))
				if status.Code(err) == codes.Canceled {
					return
				}
				break
			}
			// @todo
			a.l.InfoCtx(ctx, "event", zap.Any("event", event))
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
