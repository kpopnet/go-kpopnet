package kpopnet

import (
	"sync"
)

type cacheKey int

const (
	profileCacheKey cacheKey = iota
	trainDataCacheKey
)

var (
	mu    sync.Mutex
	cache = make(map[cacheKey]interface{}, 2)
)

func cached(key cacheKey, makev func() (interface{}, error)) (v interface{}, err error) {
	mu.Lock()
	defer mu.Unlock()

	v, ok := cache[key]
	if ok {
		return
	}

	if v, err = makev(); err != nil {
		return
	}
	cache[key] = v
	return
}

// ClearProfilesCache wipes cached profiles info.
// Should be called on DB update.
func ClearProfilesCache() {
	mu.Lock()
	defer mu.Unlock()
	delete(cache, profileCacheKey)
}
