package secret

import (
	"context"
	"gophkeeper/internal/server/service"
	"testing"

	"github.com/stretchr/testify/assert"
)

// stubFindSecretFunc creates a stub for FindSecretFunc.
type stubFindSecretFunc struct{}

func (s *stubFindSecretFunc) FindSecret(ctx context.Context, userID int64, id string) (*service.Secret, error) {
	return nil, nil
}

func TestNewServiceServer(t *testing.T) {
	t.Run("Default initialization", func(t *testing.T) {
		s := NewServiceServer()
		assert.NotNil(t, s, "ServiceServer should not be nil")
		assert.Equal(t, int64(defaultTimeWindow), s.timeWindow, "timeWindow should be default value")
		assert.Nil(t, s.findUserFunc, "findUserFunc should be nil")
		assert.Nil(t, s.findSecretFunc, "findSecretFunc should be nil")
		assert.Nil(t, s.listSecretsFunc, "listSecretsFunc should be nil")
		assert.Nil(t, s.saveSecretFunc, "saveSecretFunc should be nil")
		assert.Nil(t, s.deleteSecretFunc, "deleteSecretFunc should be nil")
		assert.Nil(t, s.sync, "sync service should be nil")
	})

	t.Run("With custom options", func(t *testing.T) {
		tw := int64(300)
		findUserFunc := &stubFindUserByIDFunc{}
		findSecretFunc := &stubFindSecretFunc{}
		listSecretsFunc := &stubListSecretsFunc{}
		saveSecretFunc := &stubSaveSecretFunc{}
		deleteSecretFunc := &stubDeleteSecretFunc{}
		syncService := &service.SyncService{}

		s := NewServiceServer(
			WithTimeWindow(tw),
			WithFindUserFunc(findUserFunc.FindUserByID),
			WithFindSecretFunc(findSecretFunc.FindSecret),
			WithListSecretsFunc(listSecretsFunc.ListSecrets),
			WithSaveSecretFunc(saveSecretFunc.SaveSecret),
			WithDeleteSecretFunc(deleteSecretFunc.DeleteSecret),
			WithSyncService(syncService),
		)

		assert.Equal(t, tw, s.timeWindow, "timeWindow should match custom value")
		assert.NotNil(t, s.findUserFunc, "findUserFunc should not be nil")
		assert.NotNil(t, s.findSecretFunc, "findSecretFunc should not be nil")
		assert.NotNil(t, s.listSecretsFunc, "listSecretsFunc should not be nil")
		assert.NotNil(t, s.saveSecretFunc, "saveSecretFunc should not be nil")
		assert.NotNil(t, s.deleteSecretFunc, "deleteSecretFunc should not be nil")
		assert.Equal(t, syncService, s.sync, "sync service should match provided value")
	})
}
