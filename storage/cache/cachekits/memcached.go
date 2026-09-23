package cachekits

import "github.com/bradfitz/gomemcache/memcache"

func NewMemcachedClient() (*memcache.Client, error) {
	mc := memcache.New("localhost:11211")

	if err := mc.Ping(); err != nil {
		return nil, err
	}

	return mc, nil
}
