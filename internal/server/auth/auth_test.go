package auth

import (
	"context"
	"gophkeeper/internal/server/service"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServiceServer(t *testing.T) {
	t.Run("Default initialization", func(t *testing.T) {
		s := NewServiceServer()
		assert.NotNil(t, s, "ServiceServer should not be nil")
		assert.Equal(t, int64(defaultTimeWindow), s.timeWindow, "timeWindow should be default value")
		assert.Nil(t, s.registerUserFunc, "registerUserFunc should be nil")
		assert.Nil(t, s.findUserFunc, "findUserFunc should be nil")
	})

	t.Run("With custom options", func(t *testing.T) {
		tw := int64(300)
		registerFunc := func(ctx context.Context, pubKey []byte) (int64, error) { return 0, nil }
		findFunc := func(ctx context.Context, fingerprint string) (*service.User, error) { return nil, nil }
		s := NewServiceServer(
			WithTimeWindow(tw),
			WithRegisterUserFunc(registerFunc),
			WithFindUserFunc(findFunc),
		)
		assert.Equal(t, tw, s.timeWindow, "timeWindow should match custom value")
		assert.NotNil(t, s.registerUserFunc, "registerUserFunc should not be nil")
		assert.NotNil(t, s.findUserFunc, "findUserFunc should not be nil")
	})
}
