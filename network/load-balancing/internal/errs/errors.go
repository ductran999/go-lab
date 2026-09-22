package errs

import "errors"

var (
	ErrUnsupportedAlg       = errors.New("unsupported algorithm")
	ErrNoTargetServersFound = errors.New("no target servers found")

	ErrInvalidBackendUrl = errors.New("invalid backend url")
)
