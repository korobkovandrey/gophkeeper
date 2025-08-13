package app

import (
	"context"
	"fmt"
	"gophkeeper/internal/client/model"
	"gophkeeper/pkg/proto"

	"go.uber.org/zap"
)

func (a *App) RunEventProcessor(ctx context.Context) {
	for event := range a.eventCh {
		a.l.InfoCtx(ctx, "event", zap.Any("event", event))
		if err := a.processEvent(event); err != nil {
			a.l.WarnCtx(ctx, "failed to process event", zap.Error(err))
		} else {
			a.updateEvent()
		}
	}
}

func (a *App) processEvent(event *proto.SecretEvent) error {
	if event.Secret == nil {
		return fmt.Errorf("secret is nil, action %s id %s", event.Action, event.Id)
	}
	switch event.Action {
	case proto.SecretAction_LIST, proto.SecretAction_STORED:
		secret, err := model.MakeSecret(
			event.Id, event.Secret.Id, event.Secret.Crypt, event.Secret.Meta, event.Secret.Data,
			a.time.LocalFromUnix(event.Secret.CreatedAt), a.time.LocalFromUnix(event.Secret.UpdatedAt),
			a.key.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to make secret: %w", err)
		}
		err = a.store.SyncStore(secret)
		if err != nil {
			return fmt.Errorf("%w: action %s id %s", err, event.Action, secret.ID)
		}
	case proto.SecretAction_DELETED:
		err := a.store.Delete(model.ID(event.Id), a.time.LocalFromUnix(event.Secret.UpdatedAt))
		if err != nil {
			return fmt.Errorf("%w: deleting %s", err, event.Id)
		}
	default:
		return fmt.Errorf("unknown event action: %s %s", event.Action, event.Id)
	}
	return nil
}
