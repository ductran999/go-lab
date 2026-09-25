// Command client exercises the Todos service: create, get, watch.
// Usage: go run ./cmd/client [create|get|watch] (watch streams until Ctrl-C).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	pb "go-lab/network/grpc/internal/pb"

	"github.com/ductran999/shared-pkg/environ"
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

	switch flag.Arg(0) {
	case "create":
		runCreate(ctx, client)
	case "get":
		runGet(ctx, client)
	case "chat":
		runChat(ctx, client)
	case "dual":
		runDual(ctx, client)
	default:
		runWatch(ctx, client)
	}
}

func runCreate(ctx context.Context, client pb.TodosClient) {
	todo, err := client.Create(ctx, &pb.CreateRequest{Task: "write lab"})
	if err != nil {
		fail(err)
	}

	fmt.Printf("created: %+v\n", todo)
}

func runGet(ctx context.Context, client pb.TodosClient) {
	todo, err := client.Get(ctx, &pb.GetRequest{Id: 1})
	if err != nil {
		fail(err)
	}

	fmt.Printf("got: %+v\n", todo)
}

func runWatch(ctx context.Context, client pb.TodosClient) {
	stream, err := client.Watch(ctx, &pb.WatchRequest{AfterId: 0})
	if err != nil {
		fail(err)
	}

	for {
		ev, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return
		}

		if err != nil {
			fail(err)
		}

		fmt.Printf("event id=%d type=%s task=%q\n", ev.GetId(), ev.GetType(), ev.GetTodo().GetTask())
	}
}

func runChat(ctx context.Context, client pb.TodosClient) {
	chat, err := client.Chat(ctx)
	if err != nil {
		fail(err)
	}

	for _, text := range []string{"hello", "multiplex", "bye"} {
		sendErr := chat.Send(&pb.ChatMsg{From: "client", Text: text})
		if sendErr != nil {
			fail(sendErr)
		}

		echo, err := chat.Recv()
		if err != nil {
			fail(err)
		}

		fmt.Printf("echo: %s: %s\n", echo.GetFrom(), echo.GetText())
	}

	closeErr := chat.CloseSend()
	if closeErr != nil {
		fail(closeErr)
	}
}

// runDual opens Watch + Chat on ONE connection: two stream IDs
// interleaved on one TCP. Capture lo in Wireshark to see the frames.
func runDual(ctx context.Context, client pb.TodosClient) {
	go func() {
		watch, err := client.Watch(ctx, &pb.WatchRequest{AfterId: 0})
		if err != nil {
			return
		}

		for {
			ev, err := watch.Recv()
			if err != nil {
				return
			}

			fmt.Printf("[watch] event id=%d\n", ev.GetId())
		}
	}()

	chat, err := client.Chat(ctx)
	if err != nil {
		fail(err)
	}

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	for i := range 5 {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			sendErr := chat.Send(&pb.ChatMsg{From: "client", Text: fmt.Sprintf("ping %d", i)})
			if sendErr != nil {
				fail(sendErr)
			}

			echo, err := chat.Recv()
			if err != nil {
				fail(err)
			}

			fmt.Printf("[chat] echo: %s\n", echo.GetText())
		}
	}
}
