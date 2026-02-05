package main

import (
	"log"
	"net"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"purpura.dev.br/study/grpc/server/auth"
	"purpura.dev.br/study/grpc/server/config"

	protocol "purpura.dev.br/study/grpc/protocol"
	"purpura.dev.br/study/grpc/server/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.LoadAuthConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	jwtValidator, err := auth.NewValidator(cfg.Domain, cfg.Audience)
	if err != nil {
		log.Fatalf("Failed to create validator: %v", err)
	}

	jwtValidatorInterceptor := service.NewJwtInterceptor(jwtValidator)

	svc := service.Service{}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(jwtValidatorInterceptor.Intercept),
		grpc.StreamInterceptor(jwtValidatorInterceptor.InterceptStream),
	)
	protocol.RegisterProtocolServer(server, &svc)

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	err = server.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}
