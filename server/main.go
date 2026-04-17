package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"purpura.dev.br/study/protocol"
)

type service struct {
	protocol.UnimplementedProtocolServer

	db map[string]string
}

func newService() *service {
	return &service{db: make(map[string]string)}
}

func (s *service) Get(_ context.Context, request *protocol.GetRequest) (*protocol.GetResponse, error) {
	value := s.db[request.GetName()]
	response := &protocol.GetResponse_builder{
		Value: proto.String(value),
	}
	slog.Info("get", slog.String("name", request.GetName()))
	return response.Build(), nil
}

func (s *service) Set(_ context.Context, request *protocol.SetRequest) (*protocol.SetResponse, error) {
	name := request.GetName()
	value := request.GetValue()
	s.db[name] = value
	response := &protocol.SetResponse_builder{}
	slog.Info("set", slog.String("name", request.GetName()))
	return response.Build(), nil
}

func (s *service) Clear(_ context.Context, request *protocol.ClearRequest) (*protocol.ClearResponse, error) {
	name := request.GetName()
	delete(s.db, name)
	response := &protocol.ClearResponse_builder{}
	slog.Info("clear", slog.String("name", request.GetName()))
	return response.Build(), nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: server (port)")
	}
	port := os.Args[1]

	service := newService()

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	protocol.RegisterProtocolServer(server, service)
	err = server.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("listen", slog.String("port", listener.Addr().String()))
}
