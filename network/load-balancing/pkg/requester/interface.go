package requester

import (
	"net/http"
)

type DoRequestCallback = func(c http.Client, reqID int)

type RequestSenderMode int

const (
	ParallelMode RequestSenderMode = iota
	SequentialMode
)

type Requester interface {
	Start(fn DoRequestCallback)
}
