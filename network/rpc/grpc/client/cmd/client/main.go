// Command client dials the Todos service and dispatches one demo
// mode per run. Usage: go run ./cmd/client [create|get|chat|dual|slow|health|watch].
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "go-lab/network/rpc/grpc/api/gen"
	"go-lab/network/rpc/grpc/client/internal/modes"

	"github.com/ductran999/shared-pkg/environ"
	"github.com/google/uuid"
)

func fail(err error) {
	fmt.Fprintln(os.Stderr, "client:", err)

	os.Exit(1)
}

func main() {
	flag.Parse()

	target := "localhost:" + environ.Get("PORT", "8107")

	// GRPC_TLS_CA set → verify server cert (prod posture).
	// Unset → plaintext with an explicit insecure declaration.
	var creds grpc.DialOption

	ca := environ.Get("GRPC_TLS_CA", "")
	if ca != "" {
		tlsCreds, err := credentials.NewClientTLSFromFile(ca, "localhost")
		if err != nil {
			fail(fmt.Errorf("tls: %w", err))
		}

		creds = grpc.WithTransportCredentials(tlsCreds)
	} else {
		creds = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	conn, err := grpc.NewClient(target, creds)
	if err != nil {
		fail(err)
	}

	defer func() {
		_ = conn.Close()
	}()

	client := pb.NewTodosClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Every RPC carries the demo bearer (auth interceptor skips
	// health/reflection only) plus a request id both sides log.
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer demo-token")
	ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", uuid.NewString())

	switch flag.Arg(0) {
	case "create":
		modes.Create(ctx, client)
	case "get":
		modes.Get(ctx, client)
	case "chat":
		modes.Chat(ctx, client)
	case "dual":
		modes.Dual(ctx, client)
	case "slow":
		modes.Slow(ctx, client)
	case "health":
		modes.Health(ctx, conn)
	default:
		modes.Watch(ctx, client)
	}
}
