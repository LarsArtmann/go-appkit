package integration_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	transport "github.com/larsartmann/cqrs-htmx/v4/transport"
	appkit "github.com/larsartmann/go-appkit"
	"github.com/larsartmann/go-appkit/realtime"
	"github.com/larsartmann/go-appkit/testkit"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/storage/memory/v4"
	"github.com/larsartmann/go-sse"
	"github.com/larsartmann/go-sse/ssetest"
	"github.com/oklog/ulid/v2"
)

// newSSEService starts an appkit.Service with the default middleware stack and
// the SSE-safe composition the reference consumer (cqrs-htmx setup) uses:
// NoTimeout on read/write so the server never caps stream lifetime. The
// integration module pins PUBLISHED tags (what consumers resolve); since the
// core v0.5.0 pin the NoDrainDelay sentinel IS available here, but the 1ms
// explicit drain keeps the two SSE tests exercising the real drain wait.
// The hub is mounted at /sse via realtime.Mount. Start-wait and teardown
// (graceful shutdown, server-error drain, goroutine-baseline leak assert)
// go through testkit.Serve's own cleanup — the same harness the other
// integration tests use.
func newSSEService(t *testing.T, hub *realtime.Hub) *appkit.Service {
	t.Helper()

	svc, err := appkit.NewService(appkit.ServiceConfig{
		Addr:         freeAddr(t),
		ReadTimeout:  appkit.NoTimeout,
		WriteTimeout: appkit.NoTimeout,
		DrainDelay:   time.Millisecond,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	realtime.Mount(svc.Mux, "/sse", hub)

	testkit.Serve(t, svc)

	return svc
}

// freeAddr picks a loopback address for a Service to bind. Two hardenings
// against the 2026-10-24 runner flake (TestVersionAndMetricsCompose...:
// listen_failed EADDRINUSE in testkit.Serve):
//
//  1. a process-wide issued-registry — the OS reuses a just-closed probe
//     port for the NEXT :0 pick, so two parallel tests could draw the same
//     addr before either Service bound; a repeat addr is now redrawn.
//  2. an instant rebind probe — the port must be bindable at a SECOND
//     moment, so a steal between close and rebind redraws instead of
//     failing the test minutes later inside Service start.
var (
	freeAddrMu     sync.Mutex
	freeAddrIssued = map[string]struct{}{}
)

func freeAddr(t *testing.T) string {
	t.Helper()

	cfg := &net.ListenConfig{}

	for range 25 {
		listener, err := cfg.Listen(context.Background(), "tcp", "localhost:0")
		if err != nil {
			t.Fatalf("reserve port: %v", err)
		}

		addr := listener.Addr().String()

		_ = listener.Close()

		freeAddrMu.Lock()

		_, alreadyIssued := freeAddrIssued[addr]
		freeAddrIssued[addr] = struct{}{}

		freeAddrMu.Unlock()

		if alreadyIssued {
			continue
		}

		probe, err := cfg.Listen(context.Background(), "tcp", addr)
		if err != nil {
			continue
		}

		_ = probe.Close()

		return addr
	}

	t.Fatal("freeAddr: no verifiable free port after 25 attempts")

	return ""
}

// connect opens an SSE connection to the service, optionally carrying a
// Last-Event-ID cursor. readTimeout caps the whole connection (headers plus
// body reads) via http.Client.Timeout. The body is closed via t.Cleanup;
// ownership transfers to the caller only for reading.
func connect(
	t *testing.T,
	svc *appkit.Service,
	lastEventID string,
	readTimeout time.Duration,
) (io.ReadCloser, http.Header) {
	t.Helper()

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, "http://"+svc.Addr().String()+"/sse", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if lastEventID != "" {
		req.Header.Set("Last-Event-ID", lastEventID)
	}

	resp, err := (&http.Client{Timeout: readTimeout}).Do(req)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	return resp.Body, resp.Header
}

// TestSSEHeadersFlushThroughAppkitDefaultStack mirrors the cqrs-htmx
// ADR-001 spike's flush test (M18.3): realtime.Handler flushes headers
// before subscribing (so the 200 OK is never held hostage by the first
// event), and that flush must survive appkit's default middleware stack
// (Recovery → RequestID → Logging → SecurityHeaders) so clients and reverse
// proxies see 200 OK well before the first event.
func TestSSEHeadersFlushThroughAppkitDefaultStack(t *testing.T) {
	hub := realtime.NewHub()
	svc := newSSEService(t, hub)

	const eventDelay = 400 * time.Millisecond

	start := time.Now()

	body, header := connect(t, svc, "", 10*time.Second)
	headerElapsed := time.Since(start)

	// Broadcast only after the handler's subscription is registered: the
	// header flush precedes the subscribe step, so a timer armed from test
	// start can race it under load and drop the only event this test
	// reads. SubscriberCount is the sync point (realtime README protocol).
	subscribed := false
	for range 50 {
		if hub.SubscriberCount() == 1 {
			subscribed = true

			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if !subscribed {
		t.Fatal("handler never subscribed to the hub")
	}

	// The event fires eventDelay after the sync, keeping the flush-before-
	// event timing proof: a handler that waited for the first event before
	// flushing would show headerElapsed >= eventDelay.
	time.AfterFunc(eventDelay, func() {
		hub.Broadcast(sse.Event{Event: "ping", Data: "{}", ID: sse.NewEventID("evt-1")})
	})

	if ct := header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content-type = %q, want text/event-stream", ct)
	}

	if headerElapsed >= eventDelay {
		t.Fatalf("headers after %v, expected flushed well before the first event at %v",
			headerElapsed, eventDelay)
	}

	evt := ssetest.MustReadNEvents(t, body, 1)[0]

	if evt.Type != "ping" {
		t.Fatalf("event type = %q, want ping", evt.Type)
	}

	if evt.Data() != "{}" {
		t.Fatalf("event data = %q, want {}", evt.Data())
	}
}

// TestJournalBackedReplayThroughAppkitService exercises the wiring documented
// in the realtime module notes: cqrs-htmx's transport.NewJournalSSEStore
// bridges a go-cqrs-lite journal into realtime.NewHub(realtime.WithStore(...)),
// mounted on an appkit.Service. It proves the store contract end-to-end across
// the two repositories: first connections get NO history replay (a fresh
// subscriber never receives the whole journal), Last-Event-ID reconnects
// replay exactly the missed suffix, and live broadcasts interleave.
func TestJournalBackedReplayThroughAppkitService(t *testing.T) {
	store := memory.NewMemoryStore()
	events := seedEvents(t, 2)
	appendBatch(t, store, events)

	hub := realtime.NewHub(
		realtime.WithStore(transport.NewJournalSSEStore(store, transport.DomainEventToSSE)))
	svc := newSSEService(t, hub)

	// A first connection without a cursor must NOT receive the journal
	// history: replay is a reconnect mechanism, not a cold-start backfill.
	// The heartbeat (15s) lies far beyond the 300ms probe budget, so any
	// frame within the budget would have to be a replayed journal event.
	fresh, _ := connect(t, svc, "", 300*time.Millisecond)

	_, readErr := ssetest.ReadNEvents(fresh, 1)
	if readErr == nil {
		t.Fatal("first connection received an event, want no cold-start replay")
	}

	// Reconnect with a Last-Event-ID cursor: exactly the missed suffix
	// (event 2) is replayed from the journal through the full stack.
	reconnected, _ := connect(t, svc, events[0].ID().String(), 10*time.Second)

	after := ssetest.MustReadNEvents(t, reconnected, 1)[0]
	ssetest.RequireEventID(t, after, events[1].ID().String())
	ssetest.RequireEventType(t, after, "event")

	// Live broadcasts interleave on the reconnected stream. The replayed
	// event already proves the handler passed its subscribe step, but the
	// explicit SubscriberCount sync keeps the house protocol (and the
	// sse-protocol guard) satisfied even if handler internals reorder.
	subscribed := false
	for range 50 {
		if hub.SubscriberCount() == 1 {
			subscribed = true

			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if !subscribed {
		t.Fatal("handler never subscribed to the hub")
	}

	hub.Broadcast(sse.Event{Event: "live", Data: "{}", ID: sse.NewEventID("live-1")})

	live := ssetest.MustReadNEvents(t, reconnected, 1)[0]
	ssetest.RequireEventID(t, live, "live-1")
}

func seedEvents(t *testing.T, count int) []event.Event {
	t.Helper()

	aggID, err := id.ParseStreamID(ulid.Make().String())
	if err != nil {
		t.Fatalf("parse stream ID: %v", err)
	}

	events := make([]event.Event, 0, count)
	for i := 1; i <= count; i++ {
		evt, err := event.New(
			event.Type(fmt.Sprintf("test.event.%d", i)),
			aggID,
			"test",
			event.Version(i),
			fmt.Sprintf(`{"id":%d}`, i),
		)
		if err != nil {
			t.Fatalf("create event %d: %v", i, err)
		}

		events = append(events, evt)
	}

	return events
}

func appendBatch(t *testing.T, store *memory.MemoryStore, events []event.Event) {
	t.Helper()

	ref := id.StreamRef{ID: events[0].StreamID(), Type: "test"}
	err := store.AppendBatch(context.Background(), ref, events)
	if err != nil {
		t.Fatalf("append batch: %v", err)
	}
}
