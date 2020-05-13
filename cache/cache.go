package cache

import (
	"sync"
)

type cacheKey int

const (
	// ProfileCacheKey is a key for caching profiles info.
	ProfileCacheKey cacheKey = iota
	// TrainDataCacheKey is a key for caching train data.
	TrainDataCacheKey
)

var (
	mu    sync.Mutex
	cache = make(map[cacheKey]interface{}, 2)
)

// Cached either returns data from cache or makes it via provided callback.
func Cached(key cacheKey, makev func() (interface{}, error)) (v interface{}, err error) {
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
	delete(cache, ProfileCacheKey)
}
