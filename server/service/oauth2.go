package service

import (
	"context"
	"log/slog"
	"strings"

	"github.com/auth0/go-jwt-middleware/v3/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errMissingToken    = status.Errorf(codes.Unauthenticated, "missing token")
	errInvalidToken    = status.Errorf(codes.PermissionDenied, "invalid token")
)

type JwtInterceptor struct {
	validator *validator.Validator
}

func NewJwtInterceptor(validator *validator.Validator) *JwtInterceptor {
	return &JwtInterceptor{validator: validator}
}

func (interceptor *JwtInterceptor) valid(ctx context.Context, authorization []string) (any, error) {
	if len(authorization) < 1 {
		return nil, errMissingToken
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")
	claims, err := interceptor.validator.ValidateToken(ctx, token)
	if err != nil {
		return nil, errInvalidToken
	}
	return claims, nil
}

func (interceptor *JwtInterceptor) Intercept(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		slog.Error("", errMissingMetadata)
		return nil, errMissingMetadata
	}
	_, err := interceptor.valid(ctx, md["authorization"])
	if err != nil {
		slog.Error("", err)
		return nil, err
	}
	// TODO: store claims
	return handler(ctx, req)
}

func (interceptor *JwtInterceptor) InterceptStream(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		slog.Error("", errMissingMetadata)
		return errMissingMetadata
	}
	_, err := interceptor.valid(ctx, md["authorization"])
	if err != nil {
		slog.Error("", err)
		return err
	}
	// TODO: store claims
	return handler(srv, ss)
}
