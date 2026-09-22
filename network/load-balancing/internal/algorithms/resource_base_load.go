package algorithms

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"go-lab/network/load-balancing/internal/errs"
	"go-lab/network/load-balancing/pkg/backend"
)

type resourceBaseLoadAlg struct {
	backends   []*backend.SimpleHTTPServer
	proxyCache sync.Map
}

func NewResourceBaseLoadAlg(targets []*backend.SimpleHTTPServer) (*resourceBaseLoadAlg, error) {
	if len(targets) == 0 {
		return nil, errs.ErrNoTargetServersFound
	}

	// Validate backend URLs
	for _, target := range targets {
		if target.GetUrl() == nil {
			return nil, errs.ErrInvalidBackendUrl
		}
	}

	rbl := &resourceBaseLoadAlg{
		backends:   targets,
		proxyCache: sync.Map{},
	}

	return rbl, nil
}

func (lb *resourceBaseLoadAlg) ForwardRequest(w http.ResponseWriter, r *http.Request) {
	nextUrl := lb.getNextBackend()

	// Log the next URL to which the request will be forwarded
	log.Printf("[INFO] load balancer forwarding request to: %v\n", nextUrl.String())

	// Create a reverse proxy for the next backend
	proxy := lb.getOrCreateProxy(nextUrl)

	// Serve the request using the reverse proxy
	proxy.ServeHTTP(w, r)
}

func (lb *resourceBaseLoadAlg) getOrCreateProxy(target *url.URL) *httputil.ReverseProxy {
	key := target.String()
	if proxy, ok := lb.proxyCache.Load(key); ok {
		return proxy.(*httputil.ReverseProxy)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	lb.proxyCache.Store(key, proxy)

	return proxy
}

func (lb *resourceBaseLoadAlg) getNextBackend() *url.URL {
	// Only one backend server return it intermediately
	if len(lb.backends) == 1 {
		return lb.backends[0].GetUrl()
	}

	// Lookup the backends got lowest cpu load
	minCPULoad := lb.backends[0].GetCPULoad()
	backendIdx := 0
	backendCPUs := []float64{minCPULoad}

	for idx := 1; idx < len(lb.backends); idx++ {
		backend := lb.backends[idx]
		backendCPUs = append(backendCPUs, backend.GetCPULoad())

		if minCPULoad > lb.backends[idx].GetCPULoad() {
			minCPULoad = backend.GetCPULoad()
			backendIdx = idx
		}
	}

	log.Println("----------------------------------------------------")
	log.Printf(
		"[INFO] backend connections: %v, select: %d, CPU load: %.2f \n",
		backendCPUs, backendIdx, minCPULoad,
	)

	return lb.backends[backendIdx].GetUrl()
}
