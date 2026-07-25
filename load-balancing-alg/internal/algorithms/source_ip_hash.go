package algorithms

import (
	"hash/fnv"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"go-lab/load-balancing-alg/internal/errs"
	"go-lab/load-balancing-alg/pkg/backend"
)

type sourceIPHash struct {
	backends   []*backend.SimpleHTTPServer
	proxyCache sync.Map
}

func NewSourceIPHashAlgorithm(targets []*backend.SimpleHTTPServer) (*sourceIPHash, error) {
	if len(targets) == 0 {
		return nil, errs.ErrNoTargetServersFound
	}

	// Validate backend URLs
	for _, target := range targets {
		if target.GetUrl() == nil {
			return nil, errs.ErrInvalidBackendUrl
		}
	}

	sih := &sourceIPHash{
		backends:   targets,
		proxyCache: sync.Map{},
	}

	return sih, nil
}

func (lb *sourceIPHash) ForwardRequest(w http.ResponseWriter, r *http.Request) {
	ip := lb.getClientIP(r)

	nextUrl := lb.getNextBackend(ip)

	// Log the next URL to which the request will be forwarded
	log.Println("---------------------------------------------------------")
	log.Printf(
		"[INFO] source ip: %s -> load balancer forwarding request to: %v\n",
		ip, nextUrl.String(),
	)

	// Create a reverse proxy for the next backend
	proxy := lb.getOrCreateProxy(&nextUrl)

	// Serve the request using the reverse proxy
	proxy.ServeHTTP(w, r)
}

func (lb *sourceIPHash) getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("[ERROR] failed to get client ip")
		host = r.RemoteAddr // keep real IP
	}

	return host
}

func (lb *sourceIPHash) getNextBackend(sourceIP string) url.URL {
	idx := lb.simpleHash(sourceIP, len(lb.backends))
	return *lb.backends[idx].GetUrl()
}

func (lb *sourceIPHash) simpleHash(s string, buckets int) int {
	h := fnv.New32a()
	h.Write([]byte(s)) //nolint:gosec

	return int(h.Sum32()) % buckets
}

func (lb *sourceIPHash) getOrCreateProxy(target *url.URL) *httputil.ReverseProxy {
	key := target.String()
	if proxy, ok := lb.proxyCache.Load(key); ok {
		return proxy.(*httputil.ReverseProxy)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	lb.proxyCache.Store(key, proxy)

	return proxy
}
