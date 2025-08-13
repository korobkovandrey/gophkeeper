package service

import (
	"context"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"gophkeeper/internal/server/infra/db/query"
	"gophkeeper/internal/server/service/mocks"
	"gophkeeper/pkg/crypt"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFindUserByFingerprint(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFinder := mocks.NewMockuserFinderByFingerprint(ctrl)
	findFunc := NewFindUserByFingerprintFunc(mockFinder)
	_, publicKey := crypt.GenerateTestRSAKeyPair(t)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	assert.NoError(t, err)
	fingerprint := crypt.Fingerprint(publicKeyBytes)
	encodedPublicKey := crypt.EncodePublicKey(publicKeyBytes)
	fixedTime := time.Now().Truncate(time.Second)

	tests := []struct {
		name             string
		setupMocks       func(context.Context)
		expectedUser     *User
		expectedErr      error
		compareTextError bool
	}{
		{
			name: "Successful user retrieval",
			setupMocks: func(ctx context.Context) {
				mockFinder.EXPECT().
					GetUserByFingerprint(ctx, fingerprint).
					Return(query.User{
						ID:          1,
						Fingerprint: fingerprint,
						PublicKey:   encodedPublicKey,
						CreatedAt:   fixedTime,
						UpdatedAt:   fixedTime,
					}, nil)
			},
			expectedUser: &User{
				ID:             1,
				Fingerprint:    fingerprint,
				PublicKey:      encodedPublicKey,
				CreatedAt:      fixedTime,
				UpdatedAt:      fixedTime,
				PublicKeyBytes: publicKeyBytes,
			},
		},
		{
			name: "User not found",
			setupMocks: func(ctx context.Context) {
				mockFinder.EXPECT().
					GetUserByFingerprint(ctx, fingerprint).
					Return(query.User{}, sql.ErrNoRows)
			},
			expectedUser: nil,
			expectedErr:  ErrNotFound,
		},
		{
			name: "Unexpected database error",
			setupMocks: func(ctx context.Context) {
				mockFinder.EXPECT().
					GetUserByFingerprint(ctx, fingerprint).
					Return(query.User{}, errors.New("database error"))
			},
			expectedUser:     nil,
			compareTextError: true,
			expectedErr:      fmt.Errorf("failed to get user #%s: database error", fingerprint),
		},
		{
			name: "Invalid public key format",
			setupMocks: func(ctx context.Context) {
				mockFinder.EXPECT().
					GetUserByFingerprint(ctx, fingerprint).
					Return(query.User{
						ID:          1,
						Fingerprint: fingerprint,
						PublicKey:   "invalid-public-key", // Invalid PEM format
						CreatedAt:   fixedTime,
						UpdatedAt:   fixedTime,
					}, nil)
			},
			expectedUser:     nil,
			compareTextError: true,
			expectedErr:      errors.New("failed to new user: failed to decode public key: invalid public key format"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(t.Context())
			user, err := findFunc(t.Context(), fingerprint)
			assert.Equal(t, tt.expectedUser, user, "expected user %v, got %v", tt.expectedUser, user)
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
