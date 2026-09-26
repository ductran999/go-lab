// Package service implements the Todos gRPC service over a Store:
// unary Create/Get/Sleep, server-streaming Watch with replay, bidi Chat.
package service

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "go-lab/network/rpc/grpc/api/gen"
	"go-lab/network/rpc/grpc/server/internal/store"
)

// Service fans stored todos out to subscribers.
type Service struct {
	pb.UnimplementedTodosServer

	store *store.Store
	mu    sync.Mutex
	subs  map[chan *pb.TodoEvent]struct{}
}

// New builds a Service over a Store. It panics on nil store.
func New(st *store.Store) *Service {
	if st == nil {
		panic("service: nil store")
	}

	return &Service{store: st, subs: make(map[chan *pb.TodoEvent]struct{})}
}

func (s *Service) Create(_ context.Context, req *pb.CreateRequest) (*pb.Todo, error) {
	if req.GetTask() == "" {
		return nil, status.Error(codes.InvalidArgument, "task is required")
	}

	todo := s.store.Create(req.GetTask())
	s.broadcast(&pb.TodoEvent{Id: todo.GetId(), Type: "created", Todo: todo})

	return todo, nil
}

func (s *Service) Get(_ context.Context, req *pb.GetRequest) (*pb.Todo, error) {
	todo, ok := s.store.Get(req.GetId())
	if !ok {
		return nil, status.Error(codes.NotFound, "todo not found")
	}

	return todo, nil
}

// Watch replays events after after_id, then tails live ones.
// Slow watchers drop (buffered chan, same rule as the SSE labs).
func (s *Service) Watch(req *pb.WatchRequest, stream pb.Todos_WatchServer) error {
	ch := s.subscribe()
	defer s.unsubscribe(ch)

	maxSent := req.GetAfterId()

	for _, todo := range s.store.After(maxSent) {
		ev := &pb.TodoEvent{Id: todo.GetId(), Type: "created", Todo: todo}

		sendErr := stream.Send(ev)
		if sendErr != nil {
			return sendErr
		}

		maxSent = todo.GetId()
	}

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case ev := <-ch:
			if ev.GetId() <= maxSent {
				continue
			}

			sendErr := stream.Send(ev)
			if sendErr != nil {
				return sendErr
			}

			maxSent = ev.GetId()
		}
	}
}

// Chat echoes every message back on the same stream: full-duplex on
// one stream ID. Run it alongside Watch and Wireshark shows two
// interleaved stream IDs on one TCP connection.
func (s *Service) Chat(stream pb.Todos_ChatServer) error {
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

// Sleep waits ms milliseconds or until the client goes away: cancel
// propagation in one select. DeadlineExceeded on the client means the
// server stopped here, not after useless work.
func (s *Service) Sleep(ctx context.Context, req *pb.SleepRequest) (*pb.SleepReply, error) {
	timer := time.NewTimer(time.Duration(req.GetMs()) * time.Millisecond)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "client went away")
	case <-timer.C:
		return &pb.SleepReply{Woke: true}, nil
	}
}

func (s *Service) subscribe() chan *pb.TodoEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan *pb.TodoEvent, 16)
	s.subs[ch] = struct{}{}

	return ch
}

func (s *Service) unsubscribe(ch chan *pb.TodoEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.subs, ch)
}

func (s *Service) broadcast(ev *pb.TodoEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for ch := range s.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}
