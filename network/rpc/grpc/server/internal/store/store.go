// Package store holds todos in memory: map for lookup, ordered ids
// for replay. Single-binary demo state; swap for SQL on growth.
package store

import (
	"sync"

	pb "go-lab/network/rpc/grpc/api/gen"
)

// Store is a mutex-guarded todo registry.
type Store struct {
	mu     sync.Mutex
	todos  map[int64]*pb.Todo
	order  []int64
	nextID int64
}

// New builds an empty Store.
func New() *Store {
	return &Store{todos: make(map[int64]*pb.Todo)}
}

// Create stores a task and returns it with an id.
func (s *Store) Create(task string) *pb.Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++

	todo := &pb.Todo{Id: s.nextID, Task: task}
	s.todos[todo.GetId()] = todo
	s.order = append(s.order, todo.GetId())

	return todo
}

// Get returns one todo by id.
func (s *Store) Get(id int64) (*pb.Todo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, ok := s.todos[id]

	return todo, ok
}

// After returns todos with id greater than after, in insertion order.
func (s *Store) After(after int64) []*pb.Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	var out []*pb.Todo

	for _, id := range s.order {
		if id > after {
			out = append(out, s.todos[id])
		}
	}

	return out
}
