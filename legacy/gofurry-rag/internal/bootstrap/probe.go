package bootstrap

import (
	"context"
	"time"
)

func Live() bool {
	return true
}

func Started() bool {
	return started.Load()
}

func Ready() bool {
	if !Started() || runtime == nil || runtime.pool == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return runtime.pool.Ping(ctx) == nil
}
