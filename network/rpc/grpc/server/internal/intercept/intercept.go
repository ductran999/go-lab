// Package intercept holds gRPC middleware: logging for unary and
// streams, plus bearer auth for unary calls (health/reflection skipped).
package intercept

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Logging records method, request id and code for every unary call.
// The request id comes from client metadata: H2 stream ids are
// transport-internal (numbered per direction), so correlation rides
// an explicit header both sides share.
func Logging(
	ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	dur := time.Since(start).Round(time.Millisecond)
	slog.Info("unary",
		"method", info.FullMethod,
		"correlationID", correlationID(ctx),
		"dur", dur, "code", status.Code(err))

	return resp, err
}

// LoggingStream records stream lifecycle (open/close only:
// per-message logging would drown the logs). StreamID from context
// (grpc.StreamContext -> grpc.ServerTransportStream) is best-effort.
func LoggingStream(
	srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler,
) error {
	slog.Info("stream open",
		"method", info.FullMethod,
		"correlationID", correlationID(stream.Context()))

	err := handler(srv, stream)

	slog.Info("stream close",
		"method", info.FullMethod,
		"correlationID", correlationID(stream.Context()),
		"code", status.Code(err))

	return err
}

// correlationID returns a correlation identifier from client metadata.
func correlationID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "-"
	}

	ids := md.Get("x-request-id")
	if len(ids) == 0 {
		return "-"
	}

	return ids[0]
}

// Auth requires metadata authorization: Bearer demo-token on every
// unary call except health and reflection probes.
func Auth(
	ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {
	if strings.HasPrefix(info.FullMethod, "/grpc.health.v1.Health/") ||
		info.FullMethod == "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no metadata")
	}

	auth := md.Get("authorization")
	if len(auth) == 0 || strings.TrimPrefix(auth[0], "Bearer ") != "demo-token" {
		return nil, status.Error(codes.Unauthenticated, "bad token")
	}

	return handler(ctx, req)
}
