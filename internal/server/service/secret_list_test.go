package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"gophkeeper/internal/server/infra/db/query"
	"gophkeeper/internal/server/service/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestListSecretsFunc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := mocks.NewMocksecretLister(ctrl)
	listFunc := NewListSecretsFunc(mockLister)
	userID := int64(1)
	fixedTime := time.Now().Truncate(time.Second).UTC()
	errDB := errors.New("database error")

	secret1 := query.Secret{
		ID:        "secret-123",
		UserID:    userID,
		Crypt:     []byte("encrypted-data-1"),
		Meta:      []byte("metadata-1"),
		Data:      []byte("secret-data-1"),
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}
	secret2 := query.Secret{
		ID:        "secret-456",
		UserID:    userID,
		Crypt:     []byte("encrypted-data-2"),
		Meta:      []byte("metadata-2"),
		Data:      []byte("secret-data-2"),
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}

	expectedSecrets := []*Secret{
		{
			ID:        "secret-123",
			UserID:    userID,
			Crypt:     []byte("encrypted-data-1"),
			Meta:      []byte("metadata-1"),
			Data:      []byte("secret-data-1"),
			CreatedAt: fixedTime,
			UpdatedAt: fixedTime,
		},
		{
			ID:        "secret-456",
			UserID:    userID,
			Crypt:     []byte("encrypted-data-2"),
			Meta:      []byte("metadata-2"),
			Data:      []byte("secret-data-2"),
			CreatedAt: fixedTime,
			UpdatedAt: fixedTime,
		},
	}

	tests := []struct {
		name             string
		setupMocks       func(ctx context.Context)
		expectedSecrets  []*Secret
		expectedErr      error
		compareTextError bool
	}{
		{
			name: "Successful list with secrets",
			setupMocks: func(ctx context.Context) {
				mockLister.EXPECT().
					ListSecrets(ctx, userID).
					Return([]query.Secret{secret1, secret2}, nil)
			},
			expectedSecrets: expectedSecrets,
			expectedErr:     nil,
		},
		{
			name: "Empty list of secrets",
			setupMocks: func(ctx context.Context) {
				mockLister.EXPECT().
					ListSecrets(ctx, userID).
					Return([]query.Secret{}, nil)
			},
			expectedSecrets: []*Secret{},
			expectedErr:     nil,
		},
		{
			name: "Database error",
			setupMocks: func(ctx context.Context) {
				mockLister.EXPECT().
					ListSecrets(ctx, userID).
					Return(nil, errDB)
			},
			expectedSecrets:  nil,
			expectedErr:      fmt.Errorf("failed to list secrets: %w", errDB),
			compareTextError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(t.Context())
			secrets, err := listFunc(t.Context(), userID)
			assert.Equal(t, tt.expectedSecrets, secrets, "expected secrets %v, got %v", tt.expectedSecrets, secrets)
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
