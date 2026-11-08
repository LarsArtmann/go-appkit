package integration_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit/cqrs"
	"github.com/larsartmann/go-appkit/realtime"
	"github.com/larsartmann/go-sse/ssetest"
)

func TestZZDebugProbe2(t *testing.T) {
	store := newMarkerStore()
	folded := make(chan cqrs.FoldedEvent, 64)

	var snapshotCalls atomic.Int64

	h := newFoldedHarness(t, cqrs.NotifyFolded(&markerProjection{store: store}, folded))
	defer func() { _ = h.stream.Body.Close() }() //respdrain:ignore SSE stream — Close IS the disconnect

	t.Logf("subscribers after harness: %d", h.hub.SubscriberCount())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	realtime.FoldedSource(ctx, folded, func() any {
		snapshotCalls.Add(1)

		return store.snapshot()
	}, h.hub)

	if status := postStatus(t, h.base+"/things"); status != http.StatusNoContent {
		t.Fatalf("POST /things status = %d, want 204", status)
	}

	t.Log("POST done")

	time.Sleep(2 * time.Second)
	t.Logf("after 2s: snapshotCalls=%d subscribers=%d", snapshotCalls.Load(), h.hub.SubscriberCount())

	events := ssetest.MustReadNEvents(t, h.stream.Body, 1)
	t.Logf("SSE event: type=%s id=%s data=%s", events[0].Type, events[0].ID, events[0].Data())
}
