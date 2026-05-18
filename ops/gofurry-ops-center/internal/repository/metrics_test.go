package repository

import (
	"testing"
	"time"
)

func TestAvgPointsBucketsSamples(t *testing.T) {
	since := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	buckets := map[int64]*avgBucket{}
	addAvg(buckets, since, time.Minute, since.Add(10*time.Second), 10)
	addAvg(buckets, since, time.Minute, since.Add(40*time.Second), 30)
	addAvg(buckets, since, time.Minute, since.Add(70*time.Second), 50)

	points := avgPoints(buckets)
	if len(points) != 2 {
		t.Fatalf("expected 2 points, got %d", len(points))
	}
	if !points[0].Timestamp.Equal(since) || points[0].Value != 20 {
		t.Fatalf("unexpected first point: %#v", points[0])
	}
	if !points[1].Timestamp.Equal(since.Add(time.Minute)) || points[1].Value != 50 {
		t.Fatalf("unexpected second point: %#v", points[1])
	}
}

func TestPositiveRateHandlesCounterReset(t *testing.T) {
	if got := positiveRate(-100, 10); got != 0 {
		t.Fatalf("expected reset rate 0, got %f", got)
	}
	if got := positiveRate(300, 3); got != 100 {
		t.Fatalf("expected 100 bytes/s, got %f", got)
	}
}
