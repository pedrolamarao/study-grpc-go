// Copyright (c) 2025 Pedro Lamarão. All rights reserved.

package main

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	"purpura.dev.br/study/protocol"
)

func closeOrPanic(closeable *grpc.ClientConn) {
	if err := closeable.Close(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: client (port) get|put|clear (name) [value]")
	}
	port := os.Args[1]
	command := os.Args[2]

	connection, err := grpc.NewClient(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer closeOrPanic(connection)

	requestor := protocol.NewProtocolClient(connection)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	switch command {
	case "get":
		request := protocol.GetRequest_builder{
			Name: proto.String(os.Args[3]),
		}
		response, err := requestor.Get(ctx, request.Build())
		if err != nil {
			log.Fatal(err)
		}
		log.Print(response.GetValue())
	case "put":
		if len(os.Args) < 5 {
			log.Fatal("usage: client (port) put (name) (value)")
		}
		value := os.Args[4]
		request := protocol.SetRequest_builder{
			Name:  proto.String(os.Args[3]),
			Value: proto.String(value),
		}
		_, err := requestor.Set(ctx, request.Build())
		if err != nil {
			log.Fatal(err)
		}
	case "clear":
		request := protocol.ClearRequest_builder{
			Name: proto.String(os.Args[3]),
		}
		_, err := requestor.Clear(ctx, request.Build())
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("invalid command: ", command)
	}
}
