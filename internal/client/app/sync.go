package app

import (
	"context"
	"encoding/binary"
	"fmt"
	"gophkeeper/internal/client/model"
	"gophkeeper/pkg/crypt"
	"gophkeeper/pkg/proto"

	"go.uber.org/zap"
)

func (a *App) RunSync(ctx context.Context) {
	for range a.store.SyncCh {
		a.needSync.Store(true)
		if !a.IsLogged() {
			continue
		}
		a.runSync(ctx)
	}
}

func (a *App) runSync(ctx context.Context) {
	if !a.needSync.Load() {
		return
	}
	err := a.sync(ctx)
	if err == nil {
		a.needSync.Store(false)
		a.l.InfoCtx(ctx, "sync success")
	} else {
		a.l.ErrorCtx(ctx, "failed to sync", zap.Error(err))
	}
}

func (a *App) sync(ctx context.Context) error {
	secrets := a.store.ListForSync()
	if len(secrets) == 0 {
		return nil
	}
	for _, secret := range secrets {
		eventTime := a.time.Server(secret.UpdatedAt).Unix()
		b := make([]byte, crypt.Int64BytesLen)
		binary.LittleEndian.PutUint64(b, uint64(eventTime))
		if secret.Status == model.StatusDeleting {
			auth, err := a.makeSignAuth(secret.ID, b)
			if err != nil {
				return fmt.Errorf("failed to make auth: %w", err)
			}
			_, err = a.secret.Delete(ctx, &proto.SecretDeleteRequest{
				Auth:      auth,
				EventTime: eventTime,
			})
			if err != nil {
				return fmt.Errorf("failed to delete secret: %w", err)
			}
			a.store.Delete(secret.ID)
			a.updateEvent()
		} else {
			auth, err := a.makeSignAuth(secret.ID, []byte(secret.NewID), secret.Crypt, secret.Meta, secret.Data, b)
			if err != nil {
				return fmt.Errorf("failed to make auth: %w", err)
			}
			_, err = a.secret.Store(ctx, &proto.SecretStoreRequest{
				Auth:      auth,
				Id:        string(secret.NewID),
				Crypt:     secret.Crypt,
				Meta:      secret.Meta,
				Data:      secret.Data,
				EventTime: eventTime,
			})
			if err != nil {
				return fmt.Errorf("failed to store secret: %w", err)
			}
			secret.Status = model.StatusSynced
			err = a.store.Synced(secret)
			if err != nil {
				return fmt.Errorf("failed to sync secret: %w", err)
			}
			a.updateEvent()
		}
	}
	return nil
}
