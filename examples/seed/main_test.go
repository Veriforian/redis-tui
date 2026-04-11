package main

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupMiniRedis(t *testing.T) (*miniredis.Miniredis, redis.Cmdable) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return mr, rdb
}

func TestSeedStrings(t *testing.T) {
	mr, rdb := setupMiniRedis(t)
	seedStrings(context.Background(), rdb)

	val, err := mr.Get("app:name")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if val != "redis-tui" {
		t.Errorf("app:name = %q, want %q", val, "redis-tui")
	}
}

func TestSeedLists(t *testing.T) {
	mr, rdb := setupMiniRedis(t)
	seedLists(context.Background(), rdb)

	vals, err := mr.List("queue:emails")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(vals) != 5 {
		t.Errorf("queue:emails len = %d, want 5", len(vals))
	}
}

func TestSeedSets(t *testing.T) {
	mr, rdb := setupMiniRedis(t)
	seedSets(context.Background(), rdb)

	members, err := mr.Members("tags:popular")
	if err != nil {
		t.Fatalf("members failed: %v", err)
	}
	if len(members) != 7 {
		t.Errorf("tags:popular len = %d, want 7", len(members))
	}
}

func TestSeedSortedSets(t *testing.T) {
	mr, rdb := setupMiniRedis(t)
	seedSortedSets(context.Background(), rdb)

	members, err := mr.ZMembers("leaderboard:weekly")
	if err != nil {
		t.Fatalf("zmembers failed: %v", err)
	}
	if len(members) != 8 {
		t.Errorf("leaderboard:weekly len = %d, want 8", len(members))
	}
}

func TestSeedHashes(t *testing.T) {
	mr, rdb := setupMiniRedis(t)
	seedHashes(context.Background(), rdb)

	vals, _ := mr.HKeys("user:1001")
	if len(vals) == 0 {
		t.Error("user:1001 should have fields")
	}
}

func TestSeedStreams(t *testing.T) {
	_, rdb := setupMiniRedis(t)
	// miniredis supports streams.
	seedStreams(context.Background(), rdb)
}

func TestSeedHyperLogLog(t *testing.T) {
	_, rdb := setupMiniRedis(t)
	seedHyperLogLog(context.Background(), rdb)
}

func TestSeedBitmaps(t *testing.T) {
	mr, rdb := setupMiniRedis(t)
	seedBitmaps(context.Background(), rdb)

	if !mr.Exists("bitmap:user-activity:2024-01-15") {
		t.Error("bitmap key should exist")
	}
}

func TestSeedGeo(t *testing.T) {
	_, rdb := setupMiniRedis(t)
	seedGeo(context.Background(), rdb)
}

func TestSeedTTLKeys(t *testing.T) {
	mr, rdb := setupMiniRedis(t)
	seedTTLKeys(context.Background(), rdb)

	if !mr.Exists("cache:homepage") {
		t.Error("cache:homepage should exist")
	}
	if mr.TTL("cache:homepage") == 0 {
		t.Error("cache:homepage should have TTL")
	}
}

func TestSeedNestedKeys(t *testing.T) {
	mr, rdb := setupMiniRedis(t)
	seedNestedKeys(context.Background(), rdb)

	if !mr.Exists("api:v1:auth:token") {
		t.Error("api:v1:auth:token should exist")
	}
}

func TestSeedJSONStrings(t *testing.T) {
	mr, rdb := setupMiniRedis(t)
	seedJSONStrings(context.Background(), rdb)

	if !mr.Exists("json:user-profile") {
		t.Error("json:user-profile should exist")
	}
}

func TestHasJSONModule(t *testing.T) {
	_, rdb := setupMiniRedis(t)
	// miniredis doesn't support JSON.SET, so this should return false.
	if hasJSONModule(context.Background(), rdb) {
		t.Error("expected false — miniredis doesn't support RedisJSON")
	}
}

func TestSeedJSON_ErrorHandling(t *testing.T) {
	_, rdb := setupMiniRedis(t)
	// seedJSON uses JSON.SET which miniredis doesn't support —
	// but the function handles errors gracefully (logs, doesn't crash).
	seedJSON(context.Background(), rdb)
}

func TestMust_Success(t *testing.T) {
	_, rdb := setupMiniRedis(t)
	must(rdb.Set(context.Background(), "test-must", "val", 0))
}

type mustPanic struct{}

func TestMust_Error(t *testing.T) {
	orig := logFatalf
	logFatalf = func(string, ...any) { panic(mustPanic{}) }
	t.Cleanup(func() { logFatalf = orig })

	mr2, _ := setupMiniRedis(t)
	addr := mr2.Addr()
	mr2.Close() // close to make commands fail

	rdb2 := redis.NewClient(&redis.Options{Addr: addr})
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(mustPanic); !ok {
				panic(r)
			}
		}
	}()
	must(rdb2.Set(context.Background(), "key", "val", 0))
	t.Fatal("expected panic from must()")
}

func TestNewClusterClient_Error(t *testing.T) {
	orig := logFatalf
	logFatalf = func(string, ...any) { panic(mustPanic{}) }
	t.Cleanup(func() { logFatalf = orig })

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(mustPanic); !ok {
				panic(r)
			}
		}
	}()

	// Use a closed miniredis so CLUSTER SLOTS fails.
	mr, _ := setupMiniRedis(t)
	addr := mr.Addr()
	mr.Close()
	_ = newClusterClient(context.Background(), addr)
	t.Fatal("expected panic from newClusterClient")
}

func TestRunSeeds(t *testing.T) {
	_, rdb := setupMiniRedis(t)
	// runSeeds calls all seed functions — exercises the full flow.
	runSeeds(context.Background(), rdb)
}

func TestFlushAll(t *testing.T) {
	mr, rdb := setupMiniRedis(t)
	// Seed some data first.
	mr.Set("key1", "val1")

	flushAll(context.Background(), rdb)

	if mr.Exists("key1") {
		t.Error("key1 should have been flushed")
	}
}
