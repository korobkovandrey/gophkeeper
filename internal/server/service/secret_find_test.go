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

func TestFindSecretFunc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFinder := mocks.NewMocksecretFinder(ctrl)
	findFunc := NewFindSecretFunc(mockFinder)
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
			name: "Successful secret retrieval",
			setupMocks: func(ctx context.Context) {
				mockFinder.EXPECT().
					FindSecret(ctx, query.FindSecretParams{ID: secretID, UserID: userID}).
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
			name: "Secret not found",
			setupMocks: func(ctx context.Context) {
				mockFinder.EXPECT().
					FindSecret(ctx, query.FindSecretParams{ID: secretID, UserID: userID}).
					Return(query.Secret{}, sql.ErrNoRows)
			},
			expectedSecret: nil,
			expectedErr:    ErrNotFound,
		},
		{
			name: "Database error",
			setupMocks: func(ctx context.Context) {
				mockFinder.EXPECT().
					FindSecret(ctx, query.FindSecretParams{ID: secretID, UserID: userID}).
					Return(query.Secret{}, errDB)
			},
			expectedSecret:   nil,
			expectedErr:      fmt.Errorf("failed to find secret %s: %w", secretID, errDB),
			compareTextError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(t.Context())
			secret, err := findFunc(t.Context(), userID, secretID)
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
