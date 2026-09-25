// Command server hosts the Todos gRPC service: unary Create/Get plus
// server-streaming Watch with after_id replay. Reflection is on for
// grpcurl debugging.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	pb "go-lab/network/grpc/internal/pb"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

type server struct {
	pb.UnimplementedTodosServer

	mu     sync.Mutex
	todos  map[int64]*pb.Todo
	order  []int64
	nextID int64
	subs   map[chan *pb.TodoEvent]struct{}
}

func newServer() *server {
	return &server{todos: make(map[int64]*pb.Todo), subs: make(map[chan *pb.TodoEvent]struct{})}
}

func (s *server) Create(_ context.Context, req *pb.CreateRequest) (*pb.Todo, error) {
	if req.GetTask() == "" {
		return nil, status.Error(codes.InvalidArgument, "task is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++

	todo := &pb.Todo{Id: s.nextID, Task: req.GetTask()}

	s.todos[todo.GetId()] = todo
	s.order = append(s.order, todo.GetId())
	s.broadcast(&pb.TodoEvent{Id: todo.GetId(), Type: "created", Todo: todo})

	return todo, nil
}

func (s *server) Get(_ context.Context, req *pb.GetRequest) (*pb.Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, ok := s.todos[req.GetId()]
	if !ok {
		return nil, status.Error(codes.NotFound, "todo not found")
	}

	return todo, nil
}

// Watch replays events after after_id, then tails live ones.
// Slow watchers drop (buffered chan, same rule as the SSE labs).
func (s *server) Watch(req *pb.WatchRequest, stream pb.Todos_WatchServer) error {
	s.mu.Lock()

	var missed []*pb.TodoEvent

	for _, id := range s.order {
		if id > req.GetAfterId() {
			missed = append(missed, &pb.TodoEvent{Id: id, Type: "created", Todo: s.todos[id]})
		}
	}

	ch := make(chan *pb.TodoEvent, 16)
	s.subs[ch] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.subs, ch)
		s.mu.Unlock()
	}()

	for _, ev := range missed {
		sendErr := stream.Send(ev)
		if sendErr != nil {
			return sendErr
		}
	}

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case ev := <-ch:
			if ev.GetId() <= req.GetAfterId() {
				continue
			}

			sendErr := stream.Send(ev)
			if sendErr != nil {
				return sendErr
			}
		}
	}
}

// Chat echoes every message back on the same stream: full-duplex on
// one stream ID. Run it alongside Watch and Wireshark shows two
// interleaved stream IDs on one TCP connection.
func (s *server) Chat(stream pb.Todos_ChatServer) error {
	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return err
		}

		sendErr := stream.Send(&pb.ChatMsg{From: "server", Text: "echo: " + msg.GetText()})
		if sendErr != nil {
			return sendErr
		}
	}
}

func (s *server) broadcast(ev *pb.TodoEvent) {
	for ch := range s.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}

func main() {
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

	srv := grpc.NewServer(srvOpts...)
	pb.RegisterTodosServer(srv, newServer())
	reflection.Register(srv)

	slog.Info("serving grpc lab", "addr", addr)

	serveErr := srv.Serve(lis)
	if serveErr != nil {
		fail(serveErr)
	}
}
