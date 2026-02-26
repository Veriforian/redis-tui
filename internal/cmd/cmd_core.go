package cmd

import (
	"sync"

	"github.com/davidbudnick/redis-tui/internal/db"
	"github.com/davidbudnick/redis-tui/internal/redis"
)

var (
	mu           sync.RWMutex
	config       *db.Config
	redisClient  *redis.Client
	scanSize     int64 = 1000
	includeTypes bool  = true
	version      string
)

func GetConfig() *db.Config {
	mu.RLock()
	defer mu.RUnlock()
	return config
}

func SetConfig(c *db.Config) {
	mu.Lock()
	defer mu.Unlock()
	config = c
}

func getRedisClient() *redis.Client {
	mu.RLock()
	defer mu.RUnlock()
	return redisClient
}

func setRedisClient(c *redis.Client) {
	mu.Lock()
	defer mu.Unlock()
	redisClient = c
}

func GetScanSize() int64 {
	mu.RLock()
	defer mu.RUnlock()
	return scanSize
}

func SetScanSize(s int64) {
	mu.Lock()
	defer mu.Unlock()
	scanSize = s
}

func getIncludeTypes() bool {
	mu.RLock()
	defer mu.RUnlock()
	return includeTypes
}

func SetIncludeTypes(v bool) {
	mu.Lock()
	defer mu.Unlock()
	includeTypes = v
}

func GetVersion() string {
	mu.RLock()
	defer mu.RUnlock()
	return version
}

func SetVersion(v string) {
	mu.Lock()
	defer mu.Unlock()
	version = v
}
