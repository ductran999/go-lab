package loadbalancer

import (
	"net/http"

	"go-lab/load-balancing-alg/internal/algorithms"
	"go-lab/load-balancing-alg/internal/errs"
	"go-lab/load-balancing-alg/pkg/backend"
)

type AlgorithmImplementer interface {
	ForwardRequest(w http.ResponseWriter, r *http.Request)
}

type loadBalanceHandler struct {
	targets       []*backend.SimpleHTTPServer
	algorithmImpl AlgorithmImplementer
}

func NewLoadBalancerHandler(
	alg Algorithm, targets []*backend.SimpleHTTPServer,
) (*loadBalanceHandler, error) {
	hdl := &loadBalanceHandler{
		targets: targets,
	}

	algorithmImpl, err := hdl.getAlgorithmImpl(alg)
	if err != nil {
		return nil, err
	}

	hdl.algorithmImpl = algorithmImpl

	err = hdl.validateConfig()
	if err != nil {
		return nil, err
	}

	return hdl, nil
}

func (lb *loadBalanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lb.algorithmImpl.ForwardRequest(w, r)
}

func (h *loadBalanceHandler) getAlgorithmImpl(alg Algorithm) (AlgorithmImplementer, error) {
	switch alg {
	case RoundRobin:
		return algorithms.NewRoundRobinAlg(h.targets)
	case WeightedRoundRobin:
		return algorithms.NewWeightedRoundRobinAlg(h.targets)
	case SourceIPHash:
		return algorithms.NewSourceIPHashAlgorithm(h.targets)
	case LowestLatency:
		return algorithms.NewLowestLatencyAlg(h.targets)
	case LeastConnection:
		return algorithms.NewLeastConnectionAlg(h.targets)
	case ResourceBase:
		return algorithms.NewResourceBaseLoadAlg(h.targets)
	default:
		return nil, errs.ErrUnsupportedAlg
	}
}

func (lb *loadBalanceHandler) validateConfig() error {
	if len(lb.targets) == 0 {
		return errs.ErrNoTargetServersFound
	}

	for _, s := range lb.targets {
		if s == nil {
			return errs.ErrNoTargetServersFound
		}
	}

	return nil
}
