package service

import (
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestHSetMapSkipsEmptyMap(t *testing.T) {
	if err := HSetMap("test:empty", map[string]string{}); err != nil {
		t.Fatalf("HSetMap() should skip empty maps without Redis command, got %v", err)
	}

	if err := HSetMap("test:nil", nil); err != nil {
		t.Fatalf("HSetMap() should skip nil maps without Redis command, got %v", err)
	}
}

func TestRedisCommandErrorsReturnGFError(t *testing.T) {
	oldClient := client
	client = redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	t.Cleanup(func() {
		_ = client.Close()
		client = oldClient
	})

	if _, err := HDel("test:hash", "field"); err == nil {
		t.Fatal("HDel() should return GFError when Redis command fails")
	}
	if _, err := Get("test:key"); err == nil {
		t.Fatal("Get() should return GFError when Redis command fails")
	}
	if err := Incr("test:counter"); err == nil {
		t.Fatal("Incr() should return GFError when Redis command fails")
	}
}
