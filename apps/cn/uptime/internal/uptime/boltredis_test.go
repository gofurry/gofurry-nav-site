package uptime

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestBoltStorePersistsSupportedStructures(t *testing.T) {
	path := filepath.Join(t.TempDir(), "uptime.db")
	store, err := OpenBoltStore(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	client := store.client
	if err = client.HSet(ctx, "hash", "name", "value").Err(); err != nil {
		t.Fatal(err)
	}
	if err = client.SAdd(ctx, "set", "b", "a").Err(); err != nil {
		t.Fatal(err)
	}
	if err = client.ZAdd(ctx, "zset", redis.Z{Score: 2, Member: "b"}, redis.Z{Score: 1, Member: "a"}).Err(); err != nil {
		t.Fatal(err)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}

	store, err = OpenBoltStore(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if got, getErr := store.client.HGet(ctx, "hash", "name").Result(); getErr != nil || got != "value" {
		t.Fatalf("HGet() = %q, %v", got, getErr)
	}
	if got, getErr := store.client.SMembers(ctx, "set").Result(); getErr != nil || !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("SMembers() = %v, %v", got, getErr)
	}
	got, getErr := store.client.ZRangeByScore(ctx, "zset", &redis.ZRangeBy{Min: "-inf", Max: "+inf"}).Result()
	if getErr != nil || !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("ZRangeByScore() = %v, %v", got, getErr)
	}
}
