package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"gophkeeper/internal/server/infra/db/query"
	"gophkeeper/internal/server/service/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDeleteSecretFunc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDeleter := mocks.NewMocksecretDeleter(ctrl)
	deleteFunc := NewDeleteSecretFunc(mockDeleter)
	userID := int64(1)
	secretID := "secret-123"
	fixedTime := time.Now().Truncate(time.Second).UTC()
	errDB := errors.New("database error")

	expectedSecret := &Secret{
		ID:        secretID,
		UserID:    userID,
		Crypt:     []byte("encrypted-data"),
		Meta:      []byte("metadata"),
		Data:      []byte("secret-data"),
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}

	tests := []struct {
		name             string
		setupMocks       func(ctx context.Context)
		expectedSecret   *Secret
		expectedErr      error
		compareTextError bool
	}{
		{
			name: "Successful secret deletion",
			setupMocks: func(ctx context.Context) {
				mockDeleter.EXPECT().
					DeleteSecret(ctx, query.DeleteSecretParams{
						ID:        secretID,
						UserID:    userID,
						UpdatedAt: fixedTime,
					}).
					Return(query.Secret{
						ID:        secretID,
						UserID:    userID,
						Crypt:     []byte("encrypted-data"),
						Meta:      []byte("metadata"),
						Data:      []byte("secret-data"),
						CreatedAt: fixedTime,
						UpdatedAt: fixedTime,
					}, nil)
			},
			expectedSecret: expectedSecret,
			expectedErr:    nil,
		},
		{
			name: "Secret not found (conflict)",
			setupMocks: func(ctx context.Context) {
				mockDeleter.EXPECT().
					DeleteSecret(ctx, query.DeleteSecretParams{
						ID:        secretID,
						UserID:    userID,
						UpdatedAt: fixedTime,
					}).
					Return(query.Secret{}, sql.ErrNoRows)
			},
			expectedSecret: nil,
			expectedErr:    ErrConflict,
		},
		{
			name: "Database error",
			setupMocks: func(ctx context.Context) {
				mockDeleter.EXPECT().
					DeleteSecret(ctx, query.DeleteSecretParams{
						ID:        secretID,
						UserID:    userID,
						UpdatedAt: fixedTime,
					}).
					Return(query.Secret{}, errDB)
			},
			expectedSecret:   nil,
			expectedErr:      fmt.Errorf("failed to delete: %w", errDB),
			compareTextError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(t.Context())
			secret, err := deleteFunc(t.Context(), userID, secretID, fixedTime)
			assert.Equal(t, tt.expectedSecret, secret, "expected secret %v, got %v", tt.expectedSecret, secret)
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
