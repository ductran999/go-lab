package app

import (
	"log"

	loadbalancer "go-lab/network/load-balancing/internal/load_blancer"
	"go-lab/network/load-balancing/internal/tools"
	"go-lab/network/load-balancing/pkg/backend"

	"github.com/rs/zerolog"
)

func RunRoundRobinApp(logger zerolog.Logger) {
	log.Println("[INFO] running round-robin app")

	// Initialize the backend builder and configure number of backend servers
	backendBuilder := backend.NewBackendBuilder(logger)
	backendBuilder.SetNumberOfBackends(5)

	// Build the backend servers
	backends, err := backendBuilder.Build()
	if err != nil {
		logger.Fatal().Msgf("failed when build backends: %v", err)
	}

	// Create a new load balancer on localhost:8080 using the backends and round-robin algorithm
	lb, err := loadbalancer.NewLoadBalancer("localhost", 8080, backends, loadbalancer.RoundRobin)
	if err != nil {
		logger.Fatal().Msgf("failed to init loadbalancer: %v", err)
	}

	// Start the load balancer asynchronously
	err = lb.Start()
	if err != nil {
		logger.Fatal().Msgf("failed to start load balancer: %v", err)
	}

	// Initialize a request sender component and start sending requests asynchronously
	rs := tools.NewRequestSender(20)
	go rs.SendNow()

	// Wait for a graceful shutdown signal and stop the first backend cleanly
	GracefulShutdown(logger, backendBuilder.ShutdownAllBackends)
}
