package secret

import (
	"context"
	"crypto/x509"
	"encoding/binary"
	"errors"
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

// stubDeleteSecretFunc creates a stub for DeleteSecretFunc.
type stubDeleteSecretFunc struct {
	Secret *service.Secret
	Err    error
}

func (s *stubDeleteSecretFunc) DeleteSecret(ctx context.Context, userID int64, id string, updatedAt time.Time) (*service.Secret, error) {
	return s.Secret, s.Err
}

func TestServiceServer_Delete(t *testing.T) {
	fixedTime := time.Now().Truncate(time.Second).UTC()

	userID := int64(1)
	secretID := "secret-123"
	eventTime := fixedTime.Unix()
	errDB := errors.New("database error")

	privateKey, publicKey := crypt.GenerateTestRSAKeyPair(t)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	require.NoError(t, err, "failed to marshal public key")
	fingerprint := crypt.Fingerprint(publicKeyBytes)
	encodedPublicKey := crypt.EncodePublicKey(publicKeyBytes)
	data := [][]byte{make([]byte, crypt.Int64BytesLen)}
	binary.LittleEndian.PutUint64(data[0], uint64(eventTime))
	signature, err := crypt.SignPSSWithTimestampAndUserID(privateKey, userID, eventTime, append(data, []byte(secretID), publicKeyBytes)...)
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
		req                *proto.SecretDeleteRequest
		timeWindow         int64
		findUserStub       *stubFindUserByIDFunc
		deleteSecretStub   *stubDeleteSecretFunc
		expectedResponse   *proto.Secret
		expectedErr        error
		checkErrorContains bool
	}{
		{
			name: "Successful secret deletion",
			req: &proto.SecretDeleteRequest{
				Auth: &proto.Auth{
					UserId:    userID,
					Timestamp: eventTime,
					Signature: signature,
					Id:        secretID,
				},
				EventTime: eventTime,
			},
			timeWindow:       60,
			findUserStub:     &stubFindUserByIDFunc{User: user, Err: nil},
			deleteSecretStub: &stubDeleteSecretFunc{Secret: secret, Err: nil},
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
			req: &proto.SecretDeleteRequest{
				Auth:      &proto.Auth{},
				EventTime: eventTime,
			},
			timeWindow:       60,
			findUserStub:     &stubFindUserByIDFunc{User: user, Err: nil},
			deleteSecretStub: &stubDeleteSecretFunc{},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, "Signature is required"),
		},
		{
			name: "Empty Auth.Id",
			req: &proto.SecretDeleteRequest{
				Auth: &proto.Auth{
					UserId:    userID,
					Timestamp: eventTime,
					Signature: signature,
					Id:        "",
				},
				EventTime: eventTime,
			},
			timeWindow:       60,
			findUserStub:     &stubFindUserByIDFunc{User: user, Err: nil},
			deleteSecretStub: &stubDeleteSecretFunc{},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.Unauthenticated, "invalid signature"),
		},
		{
			name: "Authentication failure (user not found)",
			req: &proto.SecretDeleteRequest{
				Auth: &proto.Auth{
					UserId:    userID,
					Timestamp: eventTime,
					Signature: signature,
					Id:        secretID,
				},
				EventTime: eventTime,
			},
			timeWindow:       60,
			findUserStub:     &stubFindUserByIDFunc{User: nil, Err: service.ErrNotFound},
			deleteSecretStub: &stubDeleteSecretFunc{},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.NotFound, "user not found"),
		},
		{
			name: "DeleteSecret conflict",
			req: &proto.SecretDeleteRequest{
				Auth: &proto.Auth{
					UserId:    userID,
					Timestamp: eventTime,
					Signature: signature,
					Id:        secretID,
				},
				EventTime: eventTime,
			},
			timeWindow:         60,
			findUserStub:       &stubFindUserByIDFunc{User: user, Err: nil},
			deleteSecretStub:   &stubDeleteSecretFunc{Secret: nil, Err: service.ErrConflict},
			expectedResponse:   nil,
			expectedErr:        status.Error(codes.FailedPrecondition, "failed to delete secret"),
			checkErrorContains: true,
		},
		{
			name: "DeleteSecret database error",
			req: &proto.SecretDeleteRequest{
				Auth: &proto.Auth{
					UserId:    userID,
					Timestamp: eventTime,
					Signature: signature,
					Id:        secretID,
				},
				EventTime: eventTime,
			},
			timeWindow:         60,
			findUserStub:       &stubFindUserByIDFunc{User: user, Err: nil},
			deleteSecretStub:   &stubDeleteSecretFunc{Secret: nil, Err: errDB},
			expectedResponse:   nil,
			expectedErr:        status.Error(codes.Internal, "failed to delete secret"),
			checkErrorContains: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncService := service.NewSyncService()
			s := NewServiceServer(
				WithFindUserFunc(tt.findUserStub.FindUserByID),
				WithDeleteSecretFunc(tt.deleteSecretStub.DeleteSecret),
				WithSyncService(syncService),
				WithTimeWindow(tt.timeWindow),
			)
			resp, err := s.Delete(t.Context(), tt.req)
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
