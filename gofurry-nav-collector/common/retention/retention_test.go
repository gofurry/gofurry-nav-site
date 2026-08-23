package retention

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDeleteBatchesPreservesBatchSemantics(t *testing.T) {
	calls := 0
	deleted, err := deleteBatches(2, time.Second, 0, func(context.Context, int) (int64, error) {
		calls++
		if calls == 1 {
			return 2, nil
		}
		return 1, nil
	})
	if err != nil || deleted != 3 || calls != 2 {
		t.Fatalf("deleted=%d calls=%d err=%v", deleted, calls, err)
	}
}

func TestDeleteBatchesReturnsPartialCount(t *testing.T) {
	want := errors.New("database failed")
	deleted, err := deleteBatches(2, time.Second, 0, func(context.Context, int) (int64, error) {
		return 1, want
	})
	if !errors.Is(err, want) || deleted != 1 {
		t.Fatalf("deleted=%d err=%v", deleted, err)
	}
}
