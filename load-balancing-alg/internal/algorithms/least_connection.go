package algorithms

import (
	"go-lab/load-balancing-alg/internal/errs"
	"go-lab/load-balancing-alg/pkg/backend"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type leastConnectionAlg struct {
	backends   []*backend.SimpleHTTPServer
	proxyCache sync.Map
}

func NewLeastConnectionAlg(targets []*backend.SimpleHTTPServer) (*leastConnectionAlg, error) {
	if len(targets) == 0 {
		return nil, errs.ErrNoTargetServersFound
	}

	// Validate backend URLs
	for _, target := range targets {
		if target.GetUrl() == nil {
			return nil, errs.ErrInvalidBackendUrl
		}
	}

	return &leastConnectionAlg{
		backends:   targets,
		proxyCache: sync.Map{},
	}, nil
}

func (lc *leastConnectionAlg) ForwardRequest(w http.ResponseWriter, r *http.Request) {
	url := lc.getNextBackend()

	log.Println("-----------------------------------------------------------------")
	// Log the next URL to which the request will be forwarded
	log.Printf("[INFO] load balancer forwarding request to: %v\n", url.String())

	proxy := lc.getOrCreateProxy(&url)

	proxy.ServeHTTP(w, r)
}

func (lc *leastConnectionAlg) getNextBackend() url.URL {
	// Only one backend server return it intermediately
	if len(lc.backends) == 1 {
		return *lc.backends[0].GetUrl()
	}

	// Lookup the backends got least connection
	minConnection := lc.backends[0].GetConnection()
	backendIdx := 0
	backendConnections := []int{minConnection}

	for idx := 1; idx < len(lc.backends); idx++ {
		backend := lc.backends[idx]
		backendConnections = append(backendConnections, backend.GetConnection())
		if minConnection > lc.backends[idx].GetConnection() {
			minConnection = backend.GetConnection()
			backendIdx = idx
		}
	}

	log.Printf(
		"[INFO] backend connections: %v, select: %d, connection: %d \n",
		backendConnections, backendIdx, minConnection,
	)

	return *lc.backends[backendIdx].GetUrl()
}

func (lb *leastConnectionAlg) getOrCreateProxy(target *url.URL) *httputil.ReverseProxy {
	key := target.String()
	if proxy, ok := lb.proxyCache.Load(key); ok {
		return proxy.(*httputil.ReverseProxy)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	lb.proxyCache.Store(key, proxy)

	return proxy
}
