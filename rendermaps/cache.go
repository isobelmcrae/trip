package rendermaps

import (
	"os"
	"path/filepath"
)

var (
	localCache = createCache()
)

func createCache() string {
	cache := filepath.Join(os.Getenv("HOME"), ".cache", "trip")
	if err := os.MkdirAll(cache, 0755); err != nil {
		panic(err)
	}
	return cache
}

func cacheInsertKey(key string, value []byte) {
	cacheFile := filepath.Join(localCache, key)
	if err := os.WriteFile(cacheFile, value, 0644); err != nil {
		panic(err)
	}
}

func cacheGetKey(key string) ([]byte, error) {
	cacheFile := filepath.Join(localCache, key)
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}
	return data, nil
}
