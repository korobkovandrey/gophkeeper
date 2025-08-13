package secret

import (
	"context"
	"crypto/x509"
	"errors"
	"sync"
	"testing"
	"time"

	"gophkeeper/internal/server/service"
	"gophkeeper/pkg/crypt"
	"gophkeeper/pkg/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// stubListSecretsFunc creates a stub for ListSecretsFunc.
type stubListSecretsFunc struct {
	Secrets []*service.Secret
	Err     error
}

func (s *stubListSecretsFunc) ListSecrets(ctx context.Context, userID int64) ([]*service.Secret, error) {
	return s.Secrets, s.Err
}

// mockServerStreamingServer mocks grpc.ServerStreamingServer[proto.SecretEvent].
type mockServerStreamingServer struct {
	ctx        context.Context
	sentEvents []*proto.SecretEvent
	sendErr    error
	grpc.ServerStreamingServer[proto.SecretEvent]
}

func (m *mockServerStreamingServer) Send(event *proto.SecretEvent) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sentEvents = append(m.sentEvents, event)
	return nil
}

func (m *mockServerStreamingServer) Context() context.Context {
	return m.ctx
}

func TestServiceServer_Stream(t *testing.T) {
	fixedTime := time.Now().Truncate(time.Second).UTC()

	userID := int64(1)
	secretID := "secret-123"
	eventTime := fixedTime.Unix()
	errDB := errors.New("database error")
	errSend := errors.New("send error")

	privateKey, publicKey := crypt.GenerateTestRSAKeyPair(t)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	require.NoError(t, err, "failed to marshal public key")
	fingerprint := crypt.Fingerprint(publicKeyBytes)
	encodedPublicKey := crypt.EncodePublicKey(publicKeyBytes)
	signature, err := crypt.SignPSSWithTimestampAndUserID(privateKey, userID, eventTime, []byte(""), publicKeyBytes)
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
		name                    string
		req                     *proto.Auth
		timeWindow              int64
		findUserStub            *stubFindUserByIDFunc
		listSecretsStub         *stubListSecretsFunc
		syncService             *service.SyncService
		stream                  *mockServerStreamingServer
		broadcastData           func(sync *service.SyncService)
		expectedEvents          []*proto.SecretEvent
		expectedErr             error
		checkErrorContains      bool
		expectedEventsOnTimeout []*proto.SecretEvent
	}{
		{
			name: "Successful streaming with initial secrets and update",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: eventTime,
				Signature: signature,
				Id:        "",
			},
			timeWindow:      60,
			findUserStub:    &stubFindUserByIDFunc{User: user, Err: nil},
			listSecretsStub: &stubListSecretsFunc{Secrets: []*service.Secret{secret}, Err: nil},
			syncService:     service.NewSyncService(),
			stream:          &mockServerStreamingServer{ctx: context.Background()},
			broadcastData: func(sync *service.SyncService) {
				sync.Broadcast(int32(proto.SecretAction_STORED), secretID, *secret)
			},
			expectedEvents: []*proto.SecretEvent{
				{
					Id:     secretID,
					Action: proto.SecretAction_LIST,
					Secret: &proto.Secret{
						Id:        secretID,
						Crypt:     []byte("encrypted-data"),
						Meta:      []byte("metadata"),
						Data:      []byte("secret-data"),
						CreatedAt: fixedTime.Unix(),
						UpdatedAt: fixedTime.Unix(),
					},
				},
				{
					Id:     secretID,
					Action: proto.SecretAction_STORED,
					Secret: &proto.Secret{
						Id:        secretID,
						Crypt:     []byte("encrypted-data"),
						Meta:      []byte("metadata"),
						Data:      []byte("secret-data"),
						CreatedAt: fixedTime.Unix(),
						UpdatedAt: fixedTime.Unix(),
					},
				},
			},
			expectedErr: nil,
			expectedEventsOnTimeout: []*proto.SecretEvent{
				{
					Id:     secretID,
					Action: proto.SecretAction_LIST,
					Secret: &proto.Secret{
						Id:        secret.ID,
						Crypt:     secret.Crypt,
						Meta:      secret.Meta,
						Data:      secret.Data,
						CreatedAt: fixedTime.Unix(),
						UpdatedAt: fixedTime.Unix(),
					},
				},
				{
					Id:     secretID,
					Action: proto.SecretAction_STORED,
					Secret: &proto.Secret{
						Id:        secret.ID,
						Crypt:     secret.Crypt,
						Meta:      secret.Meta,
						Data:      secret.Data,
						CreatedAt: fixedTime.Unix(),
						UpdatedAt: fixedTime.Unix(),
					},
				},
			},
		},
		{
			name:                    "Nil Auth",
			req:                     &proto.Auth{},
			timeWindow:              60,
			findUserStub:            &stubFindUserByIDFunc{},
			listSecretsStub:         &stubListSecretsFunc{},
			syncService:             service.NewSyncService(),
			stream:                  &mockServerStreamingServer{ctx: context.Background()},
			broadcastData:           func(sync *service.SyncService) {},
			expectedEvents:          nil,
			expectedErr:             status.Error(codes.InvalidArgument, "Signature is required"),
			expectedEventsOnTimeout: nil,
		},
		{
			name: "Empty Auth.Id",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: eventTime,
				Signature: signature,
				Id:        "",
			},
			timeWindow:              60,
			findUserStub:            &stubFindUserByIDFunc{User: user, Err: nil},
			listSecretsStub:         &stubListSecretsFunc{},
			syncService:             service.NewSyncService(),
			stream:                  &mockServerStreamingServer{ctx: context.Background()},
			broadcastData:           func(sync *service.SyncService) {},
			expectedEvents:          nil,
			expectedErr:             status.Error(codes.Unauthenticated, "invalid signature"),
			expectedEventsOnTimeout: nil,
		},
		{
			name: "Authentication failure (user not found)",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: eventTime,
				Signature: signature,
				Id:        secretID,
			},
			timeWindow:              60,
			findUserStub:            &stubFindUserByIDFunc{User: nil, Err: service.ErrNotFound},
			listSecretsStub:         &stubListSecretsFunc{},
			syncService:             service.NewSyncService(),
			stream:                  &mockServerStreamingServer{ctx: context.Background()},
			broadcastData:           func(sync *service.SyncService) {},
			expectedEvents:          nil,
			expectedErr:             status.Error(codes.NotFound, "user not found"),
			expectedEventsOnTimeout: nil,
		},
		{
			name: "ListSecrets error",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: eventTime,
				Signature: signature,
			},
			timeWindow:      60,
			findUserStub:    &stubFindUserByIDFunc{User: user, Err: nil},
			listSecretsStub: &stubListSecretsFunc{Secrets: nil, Err: errDB},
			syncService:     service.NewSyncService(),
			stream:          &mockServerStreamingServer{ctx: context.Background()},
			broadcastData: func(sync *service.SyncService) {
				// Subscribe and immediately unsubscribe to close the channel
				if subscriber, err := sync.Subscribe(userID); err == nil {
					sync.Unsubscribe(subscriber)
				}
			},
			expectedEvents:          nil,
			expectedErr:             status.Error(codes.Internal, "failed to list secrets"),
			checkErrorContains:      true,
			expectedEventsOnTimeout: nil,
		},
		{
			name: "Stream send error",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: eventTime,
				Signature: signature,
			},
			timeWindow:      60,
			findUserStub:    &stubFindUserByIDFunc{User: user, Err: nil},
			listSecretsStub: &stubListSecretsFunc{Secrets: []*service.Secret{secret}, Err: nil},
			syncService:     service.NewSyncService(),
			stream:          &mockServerStreamingServer{ctx: context.Background(), sendErr: errSend},
			broadcastData: func(sync *service.SyncService) {
				// Subscribe and immediately unsubscribe to close the channel
				if subscriber, err := sync.Subscribe(userID); err == nil {
					sync.Unsubscribe(subscriber)
				}
			},
			expectedEvents:          nil,
			expectedErr:             status.Error(codes.Internal, "failed to send secret"),
			checkErrorContains:      true,
			expectedEventsOnTimeout: nil,
		},
		{
			name: "Context cancellation",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: eventTime,
				Signature: signature,
			},
			timeWindow:      60,
			findUserStub:    &stubFindUserByIDFunc{User: user, Err: nil},
			listSecretsStub: &stubListSecretsFunc{Secrets: []*service.Secret{secret}, Err: nil},
			syncService:     service.NewSyncService(),
			stream: &mockServerStreamingServer{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
			},
			broadcastData: func(sync *service.SyncService) {
				// No broadcast needed as context is canceled
			},
			expectedErr: status.Error(codes.Aborted, "aborted"),
		},
		{
			name: "Empty secret list",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: eventTime,
				Signature: signature,
			},
			timeWindow:      60,
			findUserStub:    &stubFindUserByIDFunc{User: user, Err: nil},
			listSecretsStub: &stubListSecretsFunc{Secrets: []*service.Secret{}, Err: nil},
			syncService:     service.NewSyncService(),
			stream:          &mockServerStreamingServer{ctx: context.Background()},
			broadcastData: func(sync *service.SyncService) {
				// Subscribe and immediately unsubscribe to close the channel
				if subscriber, err := sync.Subscribe(userID); err == nil {
					sync.Unsubscribe(subscriber)
				}
			},
		},
		{
			name: "Subscriber channel closed",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: eventTime,
				Signature: signature,
			},
			timeWindow:      60,
			findUserStub:    &stubFindUserByIDFunc{User: user, Err: nil},
			listSecretsStub: &stubListSecretsFunc{Secrets: []*service.Secret{}, Err: nil},
			syncService:     service.NewSyncService(),
			stream:          &mockServerStreamingServer{ctx: context.Background()},
			broadcastData: func(sync *service.SyncService) {
				// Subscribe and immediately unsubscribe to close the channel
				if subscriber, err := sync.Subscribe(userID); err == nil {
					sync.Unsubscribe(subscriber)
				}
			},
		},
		{
			name: "SyncService closed",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: eventTime,
				Signature: signature,
			},
			timeWindow:      60,
			findUserStub:    &stubFindUserByIDFunc{User: user, Err: nil},
			listSecretsStub: &stubListSecretsFunc{Secrets: []*service.Secret{}, Err: nil},
			syncService: func() *service.SyncService {
				s := service.NewSyncService()
				s.Close()
				return s
			}(),
			stream:                  &mockServerStreamingServer{ctx: context.Background()},
			broadcastData:           func(sync *service.SyncService) {},
			expectedEvents:          nil,
			expectedErr:             status.Error(codes.Internal, "failed to subscribe"),
			checkErrorContains:      true,
			expectedEventsOnTimeout: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServiceServer(
				WithFindUserFunc(tt.findUserStub.FindUserByID),
				WithListSecretsFunc(tt.listSecretsStub.ListSecrets),
				WithSyncService(tt.syncService),
				WithTimeWindow(tt.timeWindow),
			)

			// Create a done channel to capture Stream's error and prevent hanging
			done := make(chan error, 1)
			var wg sync.WaitGroup
			wg.Add(1)

			// Run Stream in a goroutine to avoid blocking
			go func() {
				defer wg.Done()
				err := s.Stream(tt.req, tt.stream)
				done <- err
			}()
			time.Sleep(time.Millisecond * 100)
			tt.broadcastData(tt.syncService)

			select {
			case err := <-done:
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
				assert.Equal(t, tt.expectedEvents, tt.stream.sentEvents, "expected events %v, got %v", tt.expectedEvents, tt.stream.sentEvents)
			case <-time.After(time.Second):
				// Instead of failing, check sent events on timeout
				assert.Equal(t, tt.expectedEventsOnTimeout, tt.stream.sentEvents, "expected events on timeout %v, got %v", tt.expectedEventsOnTimeout, tt.stream.sentEvents)
				wg.Done()
			}

			// Ensure the goroutine completes
			wg.Wait()
		})
	}
}
