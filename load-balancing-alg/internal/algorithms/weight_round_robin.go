package algorithms

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"sync"

	"go-lab/load-balancing-alg/internal/errs"
	"go-lab/load-balancing-alg/pkg/backend"
)

type weightedRoundRobin struct {
	backends      []*backend.SimpleHTTPServer
	currentWeight int
	currentIndex  int
	proxyCache    sync.Map
	mutex         sync.Mutex
}

func NewWeightedRoundRobinAlg(targets []*backend.SimpleHTTPServer) (*weightedRoundRobin, error) {
	if len(targets) == 0 {
		return nil, errs.ErrNoTargetServersFound
	}

	// Validate backend URLs
	for _, target := range targets {
		if target.GetUrl() == nil {
			return nil, errs.ErrInvalidBackendUrl
		}
	}

	wrr := &weightedRoundRobin{
		backends:   targets,
		proxyCache: sync.Map{},
		mutex:      sync.Mutex{},
	}

	wrr.electInitialBackend()

	return wrr, nil
}

func (lb *weightedRoundRobin) ForwardRequest(w http.ResponseWriter, r *http.Request) {
	nextUrl := lb.getNextBackend()

	log.Println("-----------------------------------------------------------------")

	// Log the next URL to which the request will be forwarded
	log.Printf("[INFO] load balancer forwarding request to: %v\n", nextUrl.String())

	// Create a reverse proxy for the next backend
	proxy := lb.getOrCreateProxy(&nextUrl)

	// Serve the request using the reverse proxy
	proxy.ServeHTTP(w, r)
}

func (lb *weightedRoundRobin) getOrCreateProxy(target *url.URL) *httputil.ReverseProxy {
	key := target.String()
	if proxy, ok := lb.proxyCache.Load(key); ok {
		return proxy.(*httputil.ReverseProxy)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	lb.proxyCache.Store(key, proxy)

	return proxy
}

func (lb *weightedRoundRobin) getNextBackend() url.URL {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	if lb.currentWeight == 0 {
		lb.currentIndex = lb.calculateNextIndex()
		lb.currentWeight = lb.backends[lb.currentIndex].GetWeight()
	}

	lb.currentWeight--
	nextBackend := lb.backends[lb.currentIndex]

	return *nextBackend.GetUrl()
}

func (lb *weightedRoundRobin) electInitialBackend() {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	// Sort backends by weight in descending order
	sort.SliceStable(lb.backends, func(i, j int) bool {
		return lb.backends[i].GetWeight() > lb.backends[j].GetWeight()
	})

	// Initialize currentWeight and currentIndex
	if len(lb.backends) > 0 {
		lb.currentWeight = lb.backends[0].GetWeight()
		lb.currentIndex = 0
	}
}

func (lb *weightedRoundRobin) calculateNextIndex() int {
	current := lb.currentIndex + 1
	if current > len(lb.backends)-1 {
		current = 0
	}

	return current
}
