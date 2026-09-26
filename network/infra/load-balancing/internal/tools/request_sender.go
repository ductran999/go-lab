package tools

import (
	"fmt"
	"go-lab/network/infra/load-balancing/pkg/requester"
	"io"
	"net/http"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/rs/zerolog/log"
)

type requestSender struct {
	sender requester.Requester
}

func NewRequestSender(numRequests int) *requestSender {
	cfg := requester.Config{
		NumOfRequest: numRequests,
		Mode:         requester.ParallelMode,
		Jitter:       time.Second,
	}

	return &requestSender{
		sender: requester.NewRequester(cfg),
	}
}

func (r *requestSender) SendNow() {
	r.sender.Start(r.sendRequest)
}

// sendRequest sends an HTTP GET request to the specified endpoint and logs the response.
// It closes the response body and handles errors appropriately.
func (r *requestSender) sendRequest(c http.Client, reqID int) {
	endpoint := fmt.Sprintf("http://localhost:8080/req/%d", reqID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to make new request")
		return
	}

	// Inject fake IP into common headers
	req.Header.Set("X-Forwarded-For", faker.IPv4())

	resp, err := c.Do(req)
	if err != nil {
		log.Error().Int("request_id", reqID).Err(err).Msg("failed to send request")
		return
	}
	defer resp.Body.Close() //nolint: errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Int("request_id", reqID).Err(err).Msg("failed to read response body")
		return
	}

	log.Printf("%s", body)
}
