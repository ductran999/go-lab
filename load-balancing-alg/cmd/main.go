package main

import (
	"flag"
	"go-lab/load-balancing-alg/internal/app"
	"os"

	"github.com/rs/zerolog"
)

func main() {
	// Initialize zerolog with ConsoleWriter for pretty terminal output
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	appName := flag.String("app-name", "rr", "Load balance app to run")
	flag.Parse()

	switch *appName {
	case "rr":
		app.RunRoundRobinApp(logger)
	case "wr":
		app.RunWeightRoundRobinApp(logger)
	case "ih":
		app.RunSourceIPHashApp(logger)
	case "lc":
		app.RunLeastConnectionApp(logger)
	case "ll":
		app.RunLowestLatencyApp(logger)
	case "rb":
		app.RunResourceBaseApp(logger)
	default:
		logger.Fatal().Msg("[ERROR] app not available")
	}
}
