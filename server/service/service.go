package service

import (
	"context"

	"google.golang.org/protobuf/proto"
	"purpura.dev.br/study/grpc/protocol"
)

type Service struct {
	protocol.UnimplementedProtocolServer
}

func (_ *Service) Operation(_ context.Context, request *protocol.Request) (*protocol.Response, error) {
	response := &protocol.Response_builder{
		Message: proto.String(request.GetMessage()),
	}
	return response.Build(), nil
}
