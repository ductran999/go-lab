package requester

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type requester struct {
	config Config
}

func NewRequester(config Config) *requester {
	rs := &requester{
		config: config,
	}
	rs.applyConfig()

	return rs
}

func (r *requester) Start(fn DoRequestCallback) {
	if r.config.Mode == ParallelMode {
		r.sendParallel(fn)
	} else {
		r.sendSequential(fn)
	}

	log.Println("[INFO] all request sent")
	log.Println("[INFO] press Ctrl + C to stop the app")
}

// applyConfig will set default value for config when it missing.
func (r *requester) applyConfig() {
	if r.config.NumOfRequest == 0 {
		r.config.NumOfRequest = 10
	}

	if r.config.Jitter == 0 {
		r.config.Jitter = time.Second
	}
}

func (r *requester) sendParallel(fn DoRequestCallback) {
	wg := sync.WaitGroup{}

	c := http.Client{Timeout: 10 * time.Second}

	for i := range r.config.NumOfRequest {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			fn(c, idx)
		}(i)

		// Wait a while before send new request
		time.Sleep(r.config.Jitter)
	}

	wg.Wait()
}

func (r *requester) sendSequential(fn DoRequestCallback) {
	c := http.Client{Timeout: 10 * time.Second}

	for i := range r.config.NumOfRequest {
		fn(c, i)

		// Wait a while before send new request
		time.Sleep(r.config.Jitter)
	}
}
