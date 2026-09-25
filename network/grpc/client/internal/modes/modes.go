// Package modes holds one function per client demo mode: unary,
// streams, deadline, health. main only dials and dispatches.
package modes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	pb "go-lab/network/grpc/api/gen"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func fail(err error) {
	fmt.Fprintln(os.Stderr, "client:", err)

	os.Exit(1)
}

// Create makes one todo.
func Create(ctx context.Context, client pb.TodosClient) {
	todo, err := client.Create(ctx, &pb.CreateRequest{Task: "write lab"})
	if err != nil {
		fail(err)
	}

	fmt.Printf("created: %+v\n", todo)
}

// Get fetches todo 1.
func Get(ctx context.Context, client pb.TodosClient) {
	todo, err := client.Get(ctx, &pb.GetRequest{Id: 1})
	if err != nil {
		fail(err)
	}

	fmt.Printf("got: %+v\n", todo)
}

// Watch tails the event stream until Ctrl-C.
func Watch(ctx context.Context, client pb.TodosClient) {
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

// Chat exchanges three messages on one bidi stream.
func Chat(ctx context.Context, client pb.TodosClient) {
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

// Dual opens Watch + Chat on ONE connection: two stream IDs
// interleaved on one TCP. Capture lo in Wireshark to see the frames.
func Dual(ctx context.Context, client pb.TodosClient) {
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
			msg := &pb.ChatMsg{From: "client", Text: fmt.Sprintf("ping %d", i)}
			sendErr := chat.Send(msg)
			if sendErr != nil {
				fail(sendErr)
			}
			fmt.Println("[send]", msg.GetText())

			echo, err := chat.Recv()
			if err != nil {
				fail(err)
			}

			fmt.Println("[recv]:", echo.GetText())
		}
	}
}

// Slow asks for 2s of sleep with a 500ms deadline: the server
// stops on ctx.Done, the client reports DeadlineExceeded.
func Slow(ctx context.Context, client pb.TodosClient) {
	short, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	reply, err := client.Sleep(short, &pb.SleepRequest{Ms: 2000})
	if err != nil {
		fmt.Printf("slow: %s (server stopped waiting too)\n", status.Code(err))

		return
	}

	fmt.Printf("slow: woke=%t (unexpected)\n", reply.GetWoke())
}

// Health asks the standard health endpoint (no token needed).
func Health(ctx context.Context, conn *grpc.ClientConn) {
	check, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{})
	if err != nil {
		fail(err)
	}

	fmt.Printf("health: %s\n", check.GetStatus())
}
