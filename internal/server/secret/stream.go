package secret

import (
	"fmt"
	"gophkeeper/pkg/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ServiceServer) Stream(req *proto.Auth, stream grpc.ServerStreamingServer[proto.SecretEvent]) error {
	user, err := s.findAndAuthUser(stream.Context(), req)
	if err != nil {
		return err
	}
	subscriber, err := s.sync.Subscribe(user.ID)
	if err != nil {
		return status.Error(codes.Internal, fmt.Errorf("failed to subscribe: %w", err).Error())
	}
	defer s.sync.Unsubscribe(subscriber)
	secrets, err := s.secret.List(stream.Context(), user.ID)
	if err != nil {
		return status.Error(codes.Internal, fmt.Errorf("failed to list secrets: %w", err).Error())
	}
	for _, secret := range secrets {
		if stream.Context().Err() != nil {
			return status.Error(codes.Aborted, "aborted")
		}
		if sendErr := stream.Send(&proto.SecretEvent{
			Id:     secret.ID,
			Action: proto.SecretAction_LIST,
			Secret: &proto.Secret{
				Id:        secret.ID,
				Crypt:     secret.Crypt,
				Meta:      secret.Meta,
				Data:      secret.Data,
				CreatedAt: secret.CreatedAt.Unix(),
				UpdatedAt: secret.UpdatedAt.Unix(),
			},
		}); sendErr != nil {
			return status.Error(codes.Internal, fmt.Errorf("failed to send secret: %w", sendErr).Error())
		}
	}
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case data, ok := <-subscriber.C:
			if !ok {
				return nil
			}
			if sendErr := stream.Send(&proto.SecretEvent{
				Id:     data.ID,
				Action: proto.SecretAction(data.Action),
				Secret: &proto.Secret{
					Id:        data.Secret.ID,
					Crypt:     data.Secret.Crypt,
					Meta:      data.Secret.Meta,
					Data:      data.Secret.Data,
					CreatedAt: data.Secret.CreatedAt.Unix(),
					UpdatedAt: data.Secret.UpdatedAt.Unix(),
				},
			}); sendErr != nil {
				return status.Error(codes.Internal, fmt.Errorf("failed to send secret: %w", sendErr).Error())
			}
		}
	}
}
