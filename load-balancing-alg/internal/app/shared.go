package app

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

// GracefulShutdown handles OS signals and performs a graceful shutdown of the server.
func GracefulShutdown(logger zerolog.Logger, shutdownTasks ...func(ctx context.Context) error) {
	const shutdownTimeout = 5 * time.Second

	// Listen for SIGINT or SIGTERM
	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-shutdownCtx.Done()
	logger.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	cleanExit := true
	for _, task := range shutdownTasks {
		if err := task(ctx); err != nil {
			logger.Warn().Err(err).Msg("shutdown task error")
			cleanExit = false
		}
	}

	if cleanExit {
		logger.Info().Msg("server shut down cleanly")
	} else {
		logger.Warn().Msg("server encountered errors during shutdown")
	}
}
