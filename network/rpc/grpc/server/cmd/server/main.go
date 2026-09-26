// Command server wires the Todos service: store, interceptors, health,
// reflection. TLS when TLS_CERT/TLS_KEY are set, plaintext otherwise.
package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "go-lab/network/rpc/grpc/api/gen"
	"go-lab/network/rpc/grpc/server/internal/intercept"
	"go-lab/network/rpc/grpc/server/internal/service"
	"go-lab/network/rpc/grpc/server/internal/store"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

func main() {
	// JSON logs: greppable, ship-ready (method/req/dur/code as fields).
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	addr := ":" + environ.Get("PORT", "8107")

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fail(fmt.Errorf("listen: %w", err))
	}

	srvOpts := []grpc.ServerOption{}

	// TLS when certs are provided (prod posture); plaintext otherwise
	// (local demo). Never ship the else branch — insecure is loud here.
	if cert, key := environ.Get("TLS_CERT", ""), environ.Get("TLS_KEY", ""); cert != "" && key != "" {
		creds, err := credentials.NewServerTLSFromFile(cert, key)
		if err != nil {
			fail(fmt.Errorf("tls: %w", err))
		}

		srvOpts = append(srvOpts, grpc.Creds(creds))
	}

	srvOpts = append(srvOpts,
		grpc.ChainUnaryInterceptor(intercept.Logging, intercept.Auth),
		grpc.ChainStreamInterceptor(intercept.LoggingStream),
	)

	srv := grpc.NewServer(srvOpts...)
	pb.RegisterTodosServer(srv, service.New(store.New()))

	healthz := health.NewServer()
	healthz.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, healthz)

	reflection.Register(srv)

	slog.Info("serving grpc lab", "addr", addr)

	serveErr := srv.Serve(lis)
	if serveErr != nil {
		fail(serveErr)
	}
}
