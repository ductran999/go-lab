package algorithms

import (
	"go-lab/load-balancing-alg/internal/errs"
	"go-lab/load-balancing-alg/pkg/backend"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type lowestLatencyAlg struct {
	backends   []*backend.SimpleHTTPServer
	proxyCache sync.Map
}

func NewLowestLatencyAlg(targets []*backend.SimpleHTTPServer) (*lowestLatencyAlg, error) {
	if len(targets) == 0 {
		return nil, errs.ErrNoTargetServersFound
	}

	// Validate backend URLs
	for _, target := range targets {
		if target.GetUrl() == nil {
			return nil, errs.ErrInvalidBackendUrl
		}
	}

	lr := &lowestLatencyAlg{
		backends:   targets,
		proxyCache: sync.Map{},
	}

	return lr, nil
}

func (lb *lowestLatencyAlg) ForwardRequest(w http.ResponseWriter, r *http.Request) {
	nextUrl := lb.getNextBackend()

	// Log the next URL to which the request will be forwarded
	log.Printf("[INFO] load balancer forwarding request to: %v\n", nextUrl.String())

	// Create a reverse proxy for the next backend
	proxy := lb.getOrCreateProxy(nextUrl)

	// Serve the request using the reverse proxy
	proxy.ServeHTTP(w, r)
}

func (lb *lowestLatencyAlg) getOrCreateProxy(target *url.URL) *httputil.ReverseProxy {
	key := target.String()
	if proxy, ok := lb.proxyCache.Load(key); ok {
		return proxy.(*httputil.ReverseProxy)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	lb.proxyCache.Store(key, proxy)

	return proxy
}

func (lb *lowestLatencyAlg) getNextBackend() *url.URL {
	if len(lb.backends) == 1 {
		return lb.backends[0].GetUrl()
	}

	minLatency := lb.backends[0].Latency()
	backendIdx := 0
	backendLatency := []time.Duration{minLatency}

	for idx := 1; idx < len(lb.backends); idx++ {
		backend := lb.backends[idx]
		backendLatency = append(backendLatency, backend.Latency())

		if minLatency > backend.Latency() {
			minLatency = backend.Latency()
			backendIdx = idx
		}
	}

	log.Println("--------------------------------------------------------")
	log.Printf(
		"[INFO] backend latency: %v, select: %d, latency: %v\n",
		backendLatency, backendIdx, minLatency,
	)

	return lb.backends[backendIdx].GetUrl()
}
