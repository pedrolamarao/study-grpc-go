package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"purpura.dev.br/study/grpc/client/auth"

	protocol "purpura.dev.br/study/grpc/protocol"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	token, err := auth.RequestAuthorization(http.DefaultClient, "br.dev.purpura.study.query br.dev.purpura.study.update")
	if err != nil {
		log.Fatal(err)
	}

	connection, err := grpc.NewClient("127.0.0.1:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer connection.Close()

	requestor := protocol.NewProtocolClient(connection)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

	request := protocol.Request_builder{
		Message: proto.String("Hello!"),
	}
	response, err := requestor.Operation(ctx, request.Build())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(response.GetMessage())
}
