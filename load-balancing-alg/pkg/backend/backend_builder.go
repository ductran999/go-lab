package backend

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	DefaultNumberBackends = 3
	DefaultBackendWeight  = 2
)

var (
	ErrBuildBackend = errors.New("setup backend failed after max retries")
)

type backendBuilder struct {
	numOfBackends int
	backends      []*SimpleHTTPServer
	logger        zerolog.Logger
	randomWeight  bool
}

func NewBackendBuilder(logger zerolog.Logger) *backendBuilder {
	return &backendBuilder{
		numOfBackends: DefaultNumberBackends,
		logger:        logger,
	}
}

// SetNumberOfBackends sets the number of backends for the builder.
// It uses the maximum of num and DefaultNumberBackends, logging a warning if num is invalid.
func (b *backendBuilder) SetNumberOfBackends(num int) {
	b.numOfBackends = max(num, DefaultNumberBackends)
	if num <= 0 {
		// Log warning when an invalid (non-positive) number is provided
		b.logger.Warn().
			Int("numberOfBackends", num).
			Msg("invalid number of backends, using default")
	}
}

func (b *backendBuilder) EnableRandomWeight() {
	b.randomWeight = true
	b.logger.Info().Msg("random weight enabled for backends")
}

func (b *backendBuilder) Build() ([]*SimpleHTTPServer, error) {
	var err error
	b.logger.Info().Msg("building backends...")

	b.backends = make([]*SimpleHTTPServer, b.numOfBackends)
	for i := range b.numOfBackends {
		if b.backends[i], err = b.setupBackend(i); err != nil {
			return nil, err
		}
	}

	b.logger.Info().Msg("all backends are ready")
	return b.backends, nil
}

func (b *backendBuilder) ShutdownAllBackends(ctx context.Context) error {
	b.logger.Info().Msg("shutdown backends ...")
	wg := sync.WaitGroup{}

	for i := range b.backends {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			err := b.backends[idx].Stop(ctx)
			if err != nil {
				b.logger.Error().Err(err).Msgf("failed to shutdown server %d\n", idx)
			}
		}(i)
	}

	// Wait for all server shutdown completely
	wg.Wait()
	b.logger.Info().Msg("all backends shutdown")
	return nil
}

func (b *backendBuilder) setupBackend(id int) (*SimpleHTTPServer, error) {
	const maxRetries = 10

	for i := range maxRetries {
		port := b.getRandomPort()
		if !b.isPortAvailable("localhost", port) {
			b.logger.Warn().Msgf("retry %d/%d: port %d not available", i+1, maxRetries, port)
			continue
		}

		be := NewSimpleHTTPServer("localhost", port, id, b.createBackendWeight())
		errChan := make(chan error, 1)

		go func() {
			errChan <- be.Start()
		}()

		if b.waitForServerReady(port, errChan, id) {
			return be, nil
		}
	}

	return nil, ErrBuildBackend
}

// waitForServerReady checks if a server on the given port is ready within 5 seconds.
// It returns true if the server is ready or gracefully closed, false on timeout or error.
func (b *backendBuilder) waitForServerReady(port int, errChan <-chan error, id int) bool {
	address := fmt.Sprintf("localhost:%d", port)
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case err := <-errChan:
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				b.logger.Error().
					Msgf("server %d o %s failed: %v", id, address, err)
				return false
			}
			return true

		case <-timeout:
			b.logger.Error().
				Msgf("timeout waiting for server %d on %s", id, address)
			return false

		case <-ticker.C:
			if conn, err := net.DialTimeout("tcp", address, time.Second); err == nil {
				if err := conn.Close(); err != nil {
					b.logger.Warn().
						Msgf("failed to close TCP connection on %s: %v", address, err)
				}
				return true
			}
		}
	}
}

// Helper function to check port availability
func (b *backendBuilder) isPortAvailable(host string, port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return false
	}

	if err := ln.Close(); err != nil {
		b.logger.Warn().
			Msgf("closing probe listener on %s:%d failed: %v", host, port, err)
	}

	return true
}

// random port in 49152–65535
func (b *backendBuilder) getRandomPort() int {
	return rand.IntN(65535-49152+1) + 49152 //nolint:gosec
}

func (b *backendBuilder) createBackendWeight() int {
	if b.randomWeight {
		return rand.IntN(5) + 1 //nolint:gosec
	}
	return DefaultBackendWeight
}
