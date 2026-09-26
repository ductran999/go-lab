package requester

import "time"

type Config struct {
	NumOfRequest int
	Mode         RequestSenderMode
	Jitter       time.Duration
}
