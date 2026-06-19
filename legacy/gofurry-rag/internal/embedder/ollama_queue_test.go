package embedder

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAdmissionControllerLimitsConcurrency(t *testing.T) {
	ctrl := NewAdmissionController(2, 10, 10, time.Second)

	release1, err := ctrl.Acquire(context.Background(), PriorityQuery)
	if err != nil {
		t.Fatal(err)
	}
	release2, err := ctrl.Acquire(context.Background(), PriorityQuery)
	if err != nil {
		t.Fatal(err)
	}
	defer release1()
	defer release2()

	acquired := make(chan func(), 1)
	errCh := make(chan error, 1)
	go func() {
		release, err := ctrl.Acquire(context.Background(), PriorityQuery)
		if err != nil {
			errCh <- err
			return
		}
		acquired <- release
	}()

	select {
	case err := <-errCh:
		t.Fatalf("unexpected error: %v", err)
	case release := <-acquired:
		t.Fatalf("third request acquired too early: %v", release != nil)
	case <-time.After(75 * time.Millisecond):
	}

	status := ctrl.Status()
	if status.Active != 2 || status.QueuedQuery != 1 {
		t.Fatalf("status = %+v", status)
	}

	release1()

	select {
	case release := <-acquired:
		if release == nil {
			t.Fatal("expected release handle")
		}
		release()
	case err := <-errCh:
		t.Fatalf("unexpected error after release: %v", err)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for queued request")
	}
}

func TestAdmissionControllerPrioritizesQueryOverIngest(t *testing.T) {
	ctrl := NewAdmissionController(1, 4, 4, time.Second)

	heldRelease, err := ctrl.Acquire(context.Background(), PriorityIngest)
	if err != nil {
		t.Fatal(err)
	}
	defer heldRelease()

	ingestAcquired := make(chan struct{}, 1)
	ingestRelease := make(chan func(), 1)
	go func() {
		release, err := ctrl.Acquire(context.Background(), PriorityIngest)
		if err != nil {
			t.Errorf("ingest acquire failed: %v", err)
			return
		}
		ingestRelease <- release
		ingestAcquired <- struct{}{}
	}()

	waitForStatus(t, ctrl, func(status OllamaQueueStatus) bool {
		return status.QueuedIngest == 1
	})

	queryAcquired := make(chan struct{}, 1)
	queryRelease := make(chan func(), 1)
	go func() {
		release, err := ctrl.Acquire(context.Background(), PriorityQuery)
		if err != nil {
			t.Errorf("query acquire failed: %v", err)
			return
		}
		queryRelease <- release
		queryAcquired <- struct{}{}
	}()

	waitForStatus(t, ctrl, func(status OllamaQueueStatus) bool {
		return status.QueuedQuery == 1
	})

	heldRelease()

	select {
	case <-queryAcquired:
	case <-time.After(time.Second):
		t.Fatal("query did not acquire first")
	}

	select {
	case <-ingestAcquired:
		t.Fatal("ingest acquired before query was released")
	case <-time.After(75 * time.Millisecond):
	}

	release := <-queryRelease
	release()

	select {
	case <-ingestAcquired:
	case <-time.After(time.Second):
		t.Fatal("ingest did not acquire after query released")
	}
	(<-ingestRelease)()
}

func TestAdmissionControllerRejectsWhenQueueFull(t *testing.T) {
	ctrl := NewAdmissionController(1, 0, 0, 50*time.Millisecond)

	release, err := ctrl.Acquire(context.Background(), PriorityQuery)
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	_, err = ctrl.Acquire(context.Background(), PriorityQuery)
	var busy *BusyError
	if !errors.As(err, &busy) {
		t.Fatalf("expected BusyError, got %v", err)
	}
	if busy.HTTPStatus() != 503 {
		t.Fatalf("status = %d", busy.HTTPStatus())
	}
}

func TestAdmissionControllerTimesOutWhileWaiting(t *testing.T) {
	ctrl := NewAdmissionController(1, 2, 2, 25*time.Millisecond)

	release, err := ctrl.Acquire(context.Background(), PriorityQuery)
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	started := time.Now()
	_, err = ctrl.Acquire(context.Background(), PriorityQuery)
	var busy *BusyError
	if !errors.As(err, &busy) {
		t.Fatalf("expected BusyError, got %v", err)
	}
	if elapsed := time.Since(started); elapsed < 20*time.Millisecond {
		t.Fatalf("timeout returned too early: %s", elapsed)
	}
}

func waitForStatus(t *testing.T, ctrl *AdmissionController, check func(OllamaQueueStatus) bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if check(ctrl.Status()) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("status condition not met: %+v", ctrl.Status())
}
