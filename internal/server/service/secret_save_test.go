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

func TestSaveSecretFunc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMocksecretRepository(ctrl)
	saveFunc := NewSaveSecretFunc(mockRepo)
	userID := int64(1)
	id := "secret-123"
	newID := "secret-123"
	fixedTime := time.Now().Truncate(time.Second).UTC()
	cryptData := []byte("encrypted-data")
	metaData := []byte("metadata")
	data := []byte("secret-data")
	errDB := errors.New("database error")

	expectedSecret := &Secret{
		ID:        newID,
		UserID:    userID,
		Crypt:     cryptData,
		Meta:      metaData,
		Data:      data,
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}

	tests := []struct {
		name             string
		id               string
		newID            string
		setupMocks       func(context.Context)
		expectedSecret   *Secret
		expectedErr      error
		compareTextError bool
	}{
		{
			name:  "Create new secret",
			id:    id,
			newID: newID,
			setupMocks: func(ctx context.Context) {
				mockRepo.EXPECT().
					FindSecret(ctx, query.FindSecretParams{ID: id, UserID: userID}).
					Return(query.Secret{}, sql.ErrNoRows)
				mockRepo.EXPECT().
					CreateSecret(ctx, query.CreateSecretParams{
						ID:        newID,
						UserID:    userID,
						Crypt:     cryptData,
						Meta:      metaData,
						Data:      data,
						CreatedAt: fixedTime,
					}).
					Return(query.Secret{
						ID:        newID,
						UserID:    userID,
						Crypt:     cryptData,
						Meta:      metaData,
						Data:      data,
						CreatedAt: fixedTime,
						UpdatedAt: fixedTime,
					}, nil)
			},
			expectedSecret: expectedSecret,
			expectedErr:    nil,
		},
		{
			name:  "Update existing secret (same ID)",
			id:    id,
			newID: newID,
			setupMocks: func(ctx context.Context) {
				mockRepo.EXPECT().
					FindSecret(ctx, query.FindSecretParams{ID: id, UserID: userID}).
					Return(query.Secret{
						ID:        id,
						UserID:    userID,
						Crypt:     []byte("old-data"),
						Meta:      []byte("old-meta"),
						Data:      []byte("old-secret"),
						CreatedAt: fixedTime,
						UpdatedAt: fixedTime,
					}, nil)
				mockRepo.EXPECT().
					UpdateSecret(ctx, query.UpdateSecretParams{
						ID:        newID,
						ID_2:      id,
						UserID:    userID,
						Crypt:     cryptData,
						Meta:      metaData,
						Data:      data,
						UpdatedAt: fixedTime,
					}).
					Return(query.Secret{
						ID:        newID,
						UserID:    userID,
						Crypt:     cryptData,
						Meta:      metaData,
						Data:      data,
						CreatedAt: fixedTime,
						UpdatedAt: fixedTime,
					}, nil)
			},
			expectedSecret: expectedSecret,
			expectedErr:    nil,
		},
		{
			name:  "Update with different ID (newID exists, delete and update)",
			id:    "old-secret",
			newID: newID,
			setupMocks: func(ctx context.Context) {
				mockRepo.EXPECT().
					FindSecret(ctx, query.FindSecretParams{ID: "old-secret", UserID: userID}).
					Return(query.Secret{ID: "old-secret", UserID: userID}, nil)
				mockRepo.EXPECT().
					FindSecret(ctx, query.FindSecretParams{ID: newID, UserID: userID}).
					Return(query.Secret{ID: newID, UserID: userID}, nil)
				mockRepo.EXPECT().
					DeleteSecret(ctx, query.DeleteSecretParams{
						UserID:    userID,
						ID:        newID,
						UpdatedAt: fixedTime,
					}).
					Return(query.Secret{ID: newID, UserID: userID}, nil)
				mockRepo.EXPECT().
					UpdateSecret(ctx, query.UpdateSecretParams{
						ID:        newID,
						ID_2:      "old-secret",
						UserID:    userID,
						Crypt:     cryptData,
						Meta:      metaData,
						Data:      data,
						UpdatedAt: fixedTime,
					}).
					Return(query.Secret{
						ID:        newID,
						UserID:    userID,
						Crypt:     cryptData,
						Meta:      metaData,
						Data:      data,
						CreatedAt: fixedTime,
						UpdatedAt: fixedTime,
					}, nil)
			},
			expectedSecret: expectedSecret,
			expectedErr:    nil,
		},
		{
			name:  "FindSecret error",
			id:    id,
			newID: newID,
			setupMocks: func(ctx context.Context) {
				mockRepo.EXPECT().
					FindSecret(ctx, query.FindSecretParams{ID: id, UserID: userID}).
					Return(query.Secret{}, errors.New("database error"))
			},
			expectedSecret:   nil,
			expectedErr:      errors.New("failed to find: database error"),
			compareTextError: true,
		},
		{
			name:  "CreateSecret error",
			id:    id,
			newID: newID,
			setupMocks: func(ctx context.Context) {
				mockRepo.EXPECT().
					FindSecret(ctx, query.FindSecretParams{ID: id, UserID: userID}).
					Return(query.Secret{}, sql.ErrNoRows)
				mockRepo.EXPECT().
					CreateSecret(ctx, query.CreateSecretParams{
						ID:        newID,
						UserID:    userID,
						Crypt:     cryptData,
						Meta:      metaData,
						Data:      data,
						CreatedAt: fixedTime,
					}).
					Return(query.Secret{}, errors.New("database error"))
			},
			expectedSecret:   nil,
			expectedErr:      errors.New("failed to create: database error"),
			compareTextError: true,
		},
		{
			name:  "UpdateSecret conflict",
			id:    id,
			newID: newID,
			setupMocks: func(ctx context.Context) {
				mockRepo.EXPECT().
					FindSecret(ctx, query.FindSecretParams{ID: id, UserID: userID}).
					Return(query.Secret{ID: id, UserID: userID}, nil)
				mockRepo.EXPECT().
					UpdateSecret(ctx, query.UpdateSecretParams{
						ID:        newID,
						ID_2:      id,
						UserID:    userID,
						Crypt:     cryptData,
						Meta:      metaData,
						Data:      data,
						UpdatedAt: fixedTime,
					}).
					Return(query.Secret{}, sql.ErrNoRows)
			},
			expectedSecret: nil,
			expectedErr:    ErrConflict,
		},
		{
			name:  "DeleteSecret error when newID exists",
			id:    "old-secret",
			newID: newID,
			setupMocks: func(ctx context.Context) {
				mockRepo.EXPECT().
					FindSecret(ctx, query.FindSecretParams{ID: "old-secret", UserID: userID}).
					Return(query.Secret{ID: "old-secret", UserID: userID}, nil)
				mockRepo.EXPECT().
					FindSecret(ctx, query.FindSecretParams{ID: newID, UserID: userID}).
					Return(query.Secret{ID: newID, UserID: userID}, nil)
				mockRepo.EXPECT().
					DeleteSecret(ctx, query.DeleteSecretParams{
						UserID:    userID,
						ID:        newID,
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
			secret, err := saveFunc(t.Context(), userID, tt.id, tt.newID, cryptData, metaData, data, fixedTime)
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
