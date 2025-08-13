package secret

import (
	"context"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"testing"
	"time"

	"gophkeeper/internal/server/service"
	"gophkeeper/pkg/crypt"
	"gophkeeper/pkg/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// stubSaveSecretFunc creates a stub for SaveSecretFunc.
type stubSaveSecretFunc struct {
	Secret *service.Secret
	Err    error
}

func (s *stubSaveSecretFunc) SaveSecret(ctx context.Context, userID int64,
	id, newID string, crypt, meta, data []byte, storeTime time.Time) (*service.Secret, error) {
	return s.Secret, s.Err
}

func TestServiceServer_Store(t *testing.T) {
	fixedTime := time.Now().Truncate(time.Second).UTC()

	userID := int64(1)
	secretID := "secret-123"
	eventTime := fixedTime.Unix()
	errDB := errors.New("database error")

	privateKey, publicKey := crypt.GenerateTestRSAKeyPair(t)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	require.NoError(t, err)
	require.NoError(t, err, "failed to marshal public key")
	fingerprint := crypt.Fingerprint(publicKeyBytes)
	encodedPublicKey := crypt.EncodePublicKey(publicKeyBytes)
	data := [][]byte{
		[]byte(secretID),         // req.Id
		[]byte("encrypted-data"), // req.Crypt
		[]byte("metadata"),       // req.Meta
		[]byte("secret-data"),    // req.Data
	}
	b := make([]byte, crypt.Int64BytesLen)
	binary.LittleEndian.PutUint64(b, uint64(eventTime))
	signature, err := crypt.SignPSSWithTimestampAndUserID(privateKey, userID, eventTime, append(data, b, []byte(secretID), publicKeyBytes)...)
	require.NoError(t, err, "failed to sign data")

	user := &service.User{
		ID:             userID,
		Fingerprint:    fingerprint,
		PublicKey:      encodedPublicKey,
		PublicKeyBytes: publicKeyBytes,
		CreatedAt:      fixedTime,
		UpdatedAt:      fixedTime,
	}
	secret := &service.Secret{
		ID:        secretID,
		UserID:    userID,
		Crypt:     []byte("encrypted-data"),
		Meta:      []byte("metadata"),
		Data:      []byte("secret-data"),
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}

	tests := []struct {
		name               string
		req                *proto.SecretStoreRequest
		timeWindow         int64
		findUserStub       *stubFindUserByIDFunc
		saveSecretStub     *stubSaveSecretFunc
		expectedResponse   *proto.Secret
		expectedErr        error
		checkErrorContains bool
	}{
		{
			name: "Successful secret storage",
			req: &proto.SecretStoreRequest{
				Auth: &proto.Auth{
					UserId:    userID,
					Timestamp: eventTime,
					Signature: signature,
					Id:        secretID,
				},
				Id:        secretID,
				Crypt:     []byte("encrypted-data"),
				Meta:      []byte("metadata"),
				Data:      []byte("secret-data"),
				EventTime: eventTime,
			},
			timeWindow:     60,
			findUserStub:   &stubFindUserByIDFunc{User: user, Err: nil},
			saveSecretStub: &stubSaveSecretFunc{Secret: secret, Err: nil},
			expectedResponse: &proto.Secret{
				Id:        secretID,
				Crypt:     []byte("encrypted-data"),
				Meta:      []byte("metadata"),
				Data:      []byte("secret-data"),
				CreatedAt: fixedTime.Unix(),
				UpdatedAt: fixedTime.Unix(),
			},
			expectedErr: nil,
		},
		{
			name: "Nil Auth",
			req: &proto.SecretStoreRequest{
				Auth:      &proto.Auth{},
				Id:        secretID,
				Crypt:     []byte("encrypted-data"),
				Meta:      []byte("metadata"),
				Data:      []byte("secret-data"),
				EventTime: eventTime,
			},
			timeWindow:       60,
			findUserStub:     &stubFindUserByIDFunc{},
			saveSecretStub:   &stubSaveSecretFunc{},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, "Signature is required"),
		},
		{
			name: "Empty Auth.Id",
			req: &proto.SecretStoreRequest{
				Auth: &proto.Auth{
					UserId:    userID,
					Timestamp: eventTime,
					Signature: signature,
					Id:        "",
				},
				Id:        secretID,
				Crypt:     []byte("encrypted-data"),
				Meta:      []byte("metadata"),
				Data:      []byte("secret-data"),
				EventTime: eventTime,
			},
			timeWindow:       60,
			findUserStub:     &stubFindUserByIDFunc{User: user, Err: nil},
			saveSecretStub:   &stubSaveSecretFunc{},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.Unauthenticated, "invalid signature"),
		},
		{
			name: "Authentication failure (user not found)",
			req: &proto.SecretStoreRequest{
				Auth: &proto.Auth{
					UserId:    userID,
					Timestamp: eventTime,
					Signature: signature,
					Id:        secretID,
				},
				Id:        secretID,
				Crypt:     []byte("encrypted-data"),
				Meta:      []byte("metadata"),
				Data:      []byte("secret-data"),
				EventTime: eventTime,
			},
			timeWindow:       60,
			findUserStub:     &stubFindUserByIDFunc{User: nil, Err: service.ErrNotFound},
			saveSecretStub:   &stubSaveSecretFunc{},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.NotFound, "user not found"),
		},
		{
			name: "SaveSecret conflict",
			req: &proto.SecretStoreRequest{
				Auth: &proto.Auth{
					UserId:    userID,
					Timestamp: eventTime,
					Signature: signature,
					Id:        secretID,
				},
				Id:        secretID,
				Crypt:     []byte("encrypted-data"),
				Meta:      []byte("metadata"),
				Data:      []byte("secret-data"),
				EventTime: eventTime,
			},
			timeWindow:         60,
			findUserStub:       &stubFindUserByIDFunc{User: user, Err: nil},
			saveSecretStub:     &stubSaveSecretFunc{Secret: nil, Err: service.ErrConflict},
			expectedResponse:   nil,
			expectedErr:        status.Error(codes.AlreadyExists, fmt.Sprintf("failed to store secret %s", secretID)),
			checkErrorContains: true,
		},
		{
			name: "SaveSecret database error",
			req: &proto.SecretStoreRequest{
				Auth: &proto.Auth{
					UserId:    userID,
					Timestamp: eventTime,
					Signature: signature,
					Id:        secretID,
				},
				Id:        secretID,
				Crypt:     []byte("encrypted-data"),
				Meta:      []byte("metadata"),
				Data:      []byte("secret-data"),
				EventTime: eventTime,
			},
			timeWindow:         60,
			findUserStub:       &stubFindUserByIDFunc{User: user, Err: nil},
			saveSecretStub:     &stubSaveSecretFunc{Secret: nil, Err: errDB},
			expectedResponse:   nil,
			expectedErr:        status.Error(codes.Internal, fmt.Sprintf("failed to store secret %s", secretID)),
			checkErrorContains: true,
		},
		{
			name: "Empty req.Id",
			req: &proto.SecretStoreRequest{
				Auth: &proto.Auth{
					UserId:    userID,
					Timestamp: eventTime,
					Signature: signature,
					Id:        secretID,
				},
				Id:        "",
				Crypt:     []byte("encrypted-data"),
				Meta:      []byte("metadata"),
				Data:      []byte("secret-data"),
				EventTime: eventTime,
			},
			timeWindow:       60,
			findUserStub:     &stubFindUserByIDFunc{User: user, Err: nil},
			saveSecretStub:   &stubSaveSecretFunc{Secret: secret, Err: nil},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.Unauthenticated, "invalid signature"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncService := service.NewSyncService()
			s := NewServiceServer(
				WithFindUserFunc(tt.findUserStub.FindUserByID),
				WithSaveSecretFunc(tt.saveSecretStub.SaveSecret),
				WithSyncService(syncService),
				WithTimeWindow(tt.timeWindow),
			)
			resp, err := s.Store(t.Context(), tt.req)
			if tt.expectedErr == nil {
				require.NoError(t, err, "expected no error")
			} else {
				assert.Error(t, err, "expected an error")
				if tt.checkErrorContains {
					assert.Contains(t, err.Error(), tt.expectedErr.Error(), "expected error to contain %v, got %v", tt.expectedErr.Error(), err)
				} else {
					assert.Equal(t, tt.expectedErr, err, "expected error %v, got %v", tt.expectedErr, err)
				}
			}
			assert.Equal(t, tt.expectedResponse, resp, "expected response %v, got %v", tt.expectedResponse, resp)
		})
	}
}
