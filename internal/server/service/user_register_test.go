package service

import (
	"context"
	"crypto/x509"
	"database/sql"
	"errors"
	"gophkeeper/internal/server/infra/db/query"
	"gophkeeper/internal/server/service/mocks"
	"gophkeeper/pkg/crypt"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegisterUserFunc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockregisterRepository(ctrl)
	registerFunc := NewRegisterUserFunc(mockRepo)
	_, publicKey := crypt.GenerateTestRSAKeyPair(t)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	assert.NoError(t, err)
	fingerprint := crypt.Fingerprint(publicKeyBytes)

	tests := []struct {
		name             string
		setupMocks       func(context.Context)
		expectedID       int64
		expectedErr      error
		compareTextError bool
	}{
		{
			name: "Successful registration",
			setupMocks: func(ctx context.Context) {
				mockRepo.EXPECT().
					GetUserByFingerprint(ctx, fingerprint).
					Return(query.User{}, sql.ErrNoRows)
				mockRepo.EXPECT().
					CreateUser(ctx, gomock.Any()).
					Return(int64(1), nil)
			},
			expectedID:  1,
			expectedErr: nil,
		},
		{
			name: "Fingerprint already exists",
			setupMocks: func(ctx context.Context) {
				mockRepo.EXPECT().
					GetUserByFingerprint(ctx, fingerprint).
					Return(query.User{ID: 1, Fingerprint: fingerprint}, nil)
			},
			expectedID:  0,
			expectedErr: ErrConflict,
		},
		{
			name: "Database integrity constraint violation",
			setupMocks: func(ctx context.Context) {
				mockRepo.EXPECT().
					GetUserByFingerprint(ctx, fingerprint).
					Return(query.User{}, sql.ErrNoRows)
				mockRepo.EXPECT().
					CreateUser(ctx, gomock.Any()).
					Return(int64(0), &pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
			expectedID:  0,
			expectedErr: ErrConflict,
		},
		{
			name: "Unexpected database error",
			setupMocks: func(ctx context.Context) {
				mockRepo.EXPECT().
					GetUserByFingerprint(ctx, fingerprint).
					Return(query.User{}, sql.ErrNoRows)
				mockRepo.EXPECT().
					CreateUser(ctx, gomock.Any()).
					Return(int64(0), errors.New("database error"))
			},
			expectedID:       0,
			expectedErr:      errors.New("failed to create user: database error"),
			compareTextError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(t.Context())
			id, err := registerFunc(t.Context(), publicKeyBytes)
			assert.Equal(t, tt.expectedID, id, "expected ID %d, got %d", tt.expectedID, id)
			if tt.expectedErr == nil {
				assert.NoError(t, err, "expected no error")
			} else {
				assert.Error(t, err, "expected an error")
				if tt.compareTextError {
					assert.EqualError(t, err, tt.expectedErr.Error(), "expected error %v, got %v", tt.expectedErr, err)
				} else {
					assert.ErrorIs(t, err, tt.expectedErr, "expected error %v, got %v", tt.expectedErr, err)
				}
			}
		})
	}
}
