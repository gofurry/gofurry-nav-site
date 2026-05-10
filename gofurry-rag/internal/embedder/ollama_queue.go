package embedder

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Priority int

const (
	PriorityQuery Priority = iota
	PriorityIngest
)

type priorityContextKey struct{}

func WithPriority(ctx context.Context, priority Priority) context.Context {
	return context.WithValue(ctx, priorityContextKey{}, normalizePriority(priority))
}

func priorityFromContext(ctx context.Context) Priority {
	if ctx == nil {
		return PriorityQuery
	}
	if value, ok := ctx.Value(priorityContextKey{}).(Priority); ok {
		return normalizePriority(value)
	}
	return PriorityQuery
}

type OllamaQueueStatus struct {
	MaxConcurrency     int   `json:"max_concurrency"`
	QueryQueueSize     int   `json:"query_queue_size"`
	IngestQueueSize    int   `json:"ingest_queue_size"`
	Active             int   `json:"active"`
	QueuedQuery        int   `json:"queued_query"`
	QueuedIngest       int   `json:"queued_ingest"`
	Rejected           int64 `json:"rejected"`
	OldestWaitMs       int64 `json:"oldest_wait_ms"`
	WaitTimeoutSeconds int   `json:"wait_timeout_seconds"`
}

type BusyError struct {
	Reason string
	Status OllamaQueueStatus
}

func (e *BusyError) Error() string {
	if e == nil {
		return "ollama busy"
	}
	parts := []string{
		e.Reason,
		fmt.Sprintf("active=%d", e.Status.Active),
		fmt.Sprintf("queued_query=%d", e.Status.QueuedQuery),
		fmt.Sprintf("queued_ingest=%d", e.Status.QueuedIngest),
		fmt.Sprintf("max_concurrency=%d", e.Status.MaxConcurrency),
	}
	return "ollama busy: " + strings.Join(parts, " ")
}

func (e *BusyError) HTTPStatus() int {
	return http.StatusServiceUnavailable
}

type AdmissionController struct {
	mu sync.Mutex

	maxConcurrency  int
	queryQueueSize  int
	ingestQueueSize int
	waitTimeout     time.Duration
	active          int
	queryWaiters    []*admissionWaiter
	ingestWaiters   []*admissionWaiter
	rejected        int64
}

type admissionWaiter struct {
	priority   Priority
	enqueuedAt time.Time
	ready      chan struct{}
	granted    bool
}

func NewAdmissionController(maxConcurrency, queryQueueSize, ingestQueueSize int, waitTimeout time.Duration) *AdmissionController {
	if maxConcurrency <= 0 {
		maxConcurrency = 4
	}
	if queryQueueSize < 0 {
		queryQueueSize = 0
	}
	if ingestQueueSize < 0 {
		ingestQueueSize = 0
	}
	if waitTimeout < 0 {
		waitTimeout = 0
	}
	return &AdmissionController{
		maxConcurrency:  maxConcurrency,
		queryQueueSize:  queryQueueSize,
		ingestQueueSize: ingestQueueSize,
		waitTimeout:     waitTimeout,
	}
}

func (a *AdmissionController) Acquire(ctx context.Context, priority Priority) (func(), error) {
	if a == nil {
		return func() {}, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	priority = normalizePriority(priority)
	waiter := &admissionWaiter{
		priority:   priority,
		enqueuedAt: time.Now(),
		ready:      make(chan struct{}),
	}
	releaseOnce := sync.Once{}
	release := func() {
		releaseOnce.Do(func() {
			a.release()
		})
	}

	a.mu.Lock()
	if a.canGrantImmediatelyLocked() {
		a.active++
		a.mu.Unlock()
		return release, nil
	}

	if a.queueLenLocked(priority) >= a.queueLimitLocked(priority) {
		a.rejected++
		status := a.snapshotLocked(time.Now())
		a.mu.Unlock()
		return nil, &BusyError{
			Reason: queueFullReason(priority),
			Status: status,
		}
	}

	a.enqueueLocked(waiter)
	a.dispatchLocked()
	if waiter.granted {
		a.mu.Unlock()
		return release, nil
	}
	a.mu.Unlock()

	var timer *time.Timer
	if a.waitTimeout > 0 {
		timer = time.NewTimer(a.waitTimeout)
		defer timer.Stop()
	}

	for {
		if timer != nil {
			select {
			case <-waiter.ready:
				return release, nil
			case <-ctx.Done():
			case <-timer.C:
				if a.cancelWaiter(waiter) {
					a.mu.Lock()
					a.rejected++
					a.mu.Unlock()
					return nil, &BusyError{Reason: "queue wait timeout", Status: a.Status()}
				}
				return release, nil
			}
		} else {
			select {
			case <-waiter.ready:
				return release, nil
			case <-ctx.Done():
			}
		}

		if a.cancelWaiter(waiter) {
			if ctx.Err() == context.DeadlineExceeded {
				a.mu.Lock()
				a.rejected++
				a.mu.Unlock()
				return nil, &BusyError{Reason: "queue wait timeout", Status: a.Status()}
			}
			return nil, ctx.Err()
		}
		return release, nil
	}
}

func (a *AdmissionController) Status() OllamaQueueStatus {
	if a == nil {
		return OllamaQueueStatus{}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.snapshotLocked(time.Now())
}

func (a *AdmissionController) release() {
	if a == nil {
		return
	}
	a.mu.Lock()
	if a.active > 0 {
		a.active--
	}
	a.dispatchLocked()
	a.mu.Unlock()
}

func (a *AdmissionController) canGrantImmediatelyLocked() bool {
	return a.active < a.maxConcurrency && len(a.queryWaiters) == 0 && len(a.ingestWaiters) == 0
}

func (a *AdmissionController) enqueueLocked(waiter *admissionWaiter) {
	switch waiter.priority {
	case PriorityIngest:
		a.ingestWaiters = append(a.ingestWaiters, waiter)
	default:
		a.queryWaiters = append(a.queryWaiters, waiter)
	}
}

func (a *AdmissionController) dispatchLocked() {
	for a.active < a.maxConcurrency {
		waiter := a.popNextWaiterLocked()
		if waiter == nil {
			return
		}
		a.active++
		waiter.granted = true
		close(waiter.ready)
	}
}

func (a *AdmissionController) popNextWaiterLocked() *admissionWaiter {
	if len(a.queryWaiters) > 0 {
		waiter := a.queryWaiters[0]
		a.queryWaiters = a.queryWaiters[1:]
		return waiter
	}
	if len(a.ingestWaiters) > 0 {
		waiter := a.ingestWaiters[0]
		a.ingestWaiters = a.ingestWaiters[1:]
		return waiter
	}
	return nil
}

func (a *AdmissionController) cancelWaiter(waiter *admissionWaiter) bool {
	if a == nil || waiter == nil {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if waiter.granted {
		return false
	}
	if !a.removeWaiterLocked(waiter) {
		return false
	}
	return true
}

func (a *AdmissionController) removeWaiterLocked(target *admissionWaiter) bool {
	remove := func(waiters []*admissionWaiter) ([]*admissionWaiter, bool) {
		for i, waiter := range waiters {
			if waiter == target {
				copy(waiters[i:], waiters[i+1:])
				waiters = waiters[:len(waiters)-1]
				return waiters, true
			}
		}
		return waiters, false
	}

	var removed bool
	a.queryWaiters, removed = remove(a.queryWaiters)
	if removed {
		return true
	}
	a.ingestWaiters, removed = remove(a.ingestWaiters)
	return removed
}

func (a *AdmissionController) queueLenLocked(priority Priority) int {
	switch priority {
	case PriorityIngest:
		return len(a.ingestWaiters)
	default:
		return len(a.queryWaiters)
	}
}

func (a *AdmissionController) queueLimitLocked(priority Priority) int {
	switch priority {
	case PriorityIngest:
		return a.ingestQueueSize
	default:
		return a.queryQueueSize
	}
}

func (a *AdmissionController) snapshotLocked(now time.Time) OllamaQueueStatus {
	status := OllamaQueueStatus{
		MaxConcurrency:     a.maxConcurrency,
		QueryQueueSize:     a.queryQueueSize,
		IngestQueueSize:    a.ingestQueueSize,
		Active:             a.active,
		QueuedQuery:        len(a.queryWaiters),
		QueuedIngest:       len(a.ingestWaiters),
		Rejected:           a.rejected,
		WaitTimeoutSeconds: int(a.waitTimeout.Seconds()),
	}
	oldest := time.Duration(0)
	for _, waiter := range a.queryWaiters {
		wait := now.Sub(waiter.enqueuedAt)
		if oldest == 0 || wait > oldest {
			oldest = wait
		}
	}
	for _, waiter := range a.ingestWaiters {
		wait := now.Sub(waiter.enqueuedAt)
		if oldest == 0 || wait > oldest {
			oldest = wait
		}
	}
	status.OldestWaitMs = oldest.Milliseconds()
	return status
}

func normalizePriority(priority Priority) Priority {
	if priority != PriorityIngest {
		return PriorityQuery
	}
	return PriorityIngest
}

func queueFullReason(priority Priority) string {
	if priority == PriorityIngest {
		return "ingest queue full"
	}
	return "query queue full"
}
