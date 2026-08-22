package realtime_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/larsartmann/go-appkit/realtime"
	"github.com/larsartmann/go-sse"
	"github.com/larsartmann/go-sse/ssetest"
)

// memStore is a minimal in-memory sse.EventStore for testing replay.
type memStore struct {
	events []sse.Event
}

func (s *memStore) EventsAfter(lastID sse.EventID) ([]sse.Event, error) {
	var result []sse.Event

	found := lastID.IsZero()

	for _, evt := range s.events {
		if found {
			result = append(result, evt)

			continue
		}

		if evt.ID.Get() == lastID.Get() {
			found = true
		}
	}

	return result, nil
}

func mustParseID(s string) sse.EventID {
	id, err := sse.ParseEventID(s)
	if err != nil {
		panic(err)
	}

	return id
}

// httpGetURL issues a context-aware GET (noctx-compliant test helper).
func httpGetURL(t *testing.T, url string) (*http.Response, error) {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("build GET %s: %v", url, err)
	}

	return http.DefaultClient.Do(req) //nolint:wrapcheck // test helper
}

// waitForSubscriber polls SubscriberCount until it reaches want or timeout.
func waitForSubscriber(t *testing.T, hub *realtime.Hub, want int) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		if hub.SubscriberCount() >= want {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatalf("timed out waiting for %d subscribers (got %d)", want, hub.SubscriberCount())
}

// readSSEFrame reads one SSE frame (ending in \n\n) from a streaming body.
// Skips heartbeat comment frames (starting with ":").
func readSSEFrame(t *testing.T, r io.Reader) string {
	t.Helper()

	buf := make([]byte, 0, 256)
	var tmp [1]byte

	for {
		n, err := r.Read(tmp[:])
		if n > 0 {
			buf = append(buf, tmp[:n]...)

			if len(buf) >= 2 && buf[len(buf)-1] == '\n' && buf[len(buf)-2] == '\n' {
				frame := string(buf)
				// Skip heartbeat comment frames
				if strings.HasPrefix(frame, ":") {
					buf = buf[:0]

					continue
				}

				return frame
			}
		}

		if err != nil {
			t.Fatalf("read error: %v (got so far: %q)", err, string(buf))
		}
	}
}

// --- Hub tests ---

func TestHub_NewHub_CreatesBroadcaster(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()

	if hub.Broadcaster == nil {
		t.Fatal("Broadcaster is nil")
	}

	if hub.Store != nil {
		t.Fatal("Store should be nil by default")
	}
}

func TestHub_WithStore(t *testing.T) {
	t.Parallel()

	store := &memStore{}
	hub := realtime.NewHub(realtime.WithStore(store))

	if hub.Store != store {
		t.Fatal("Store not set")
	}
}

func TestHub_Broadcast_DeliversToSubscriber(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()

	ch := hub.Broadcaster.Subscribe()
	defer hub.Broadcaster.Unsubscribe(ch)

	evt := sse.Event{Event: "test", Data: "hello"}
	hub.Broadcast(evt)

	select {
	case got := <-ch:
		if got.Event != "test" || got.Data != "hello" {
			t.Fatalf("unexpected event: %+v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestHub_BroadcastMany_DeliversBatch(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()

	ch := hub.Broadcaster.Subscribe()
	defer hub.Broadcaster.Unsubscribe(ch)

	hub.BroadcastMany(
		sse.Event{Event: "a", Data: "1"},
		sse.Event{Event: "b", Data: "2"},
	)

	for i := range 2 {
		select {
		case got := <-ch:
			want := []string{"1", "2"}[i]
			if got.Data != want {
				t.Fatalf("event %d: want data %q, got %q", i, want, got.Data)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for event %d", i)
		}
	}
}

func TestHub_BroadcastPatch(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()

	ch := hub.Broadcaster.Subscribe()
	defer hub.Broadcaster.Unsubscribe(ch)

	hub.BroadcastPatch(stubPatch{sse.Event{Event: "patch", Data: "<div/>"}})

	select {
	case got := <-ch:
		if got.Event != "patch" || got.Data != "<div/>" {
			t.Fatalf("unexpected event: %+v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for patch event")
	}
}

type stubPatch struct{ evt sse.Event }

func (s stubPatch) Event() sse.Event { return s.evt }

func TestHub_SubscriberCount(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()

	if got := hub.SubscriberCount(); got != 0 {
		t.Fatalf("initial count: want 0, got %d", got)
	}

	ch := hub.Broadcaster.Subscribe()
	defer hub.Broadcaster.Unsubscribe(ch)

	if got := hub.SubscriberCount(); got != 1 {
		t.Fatalf("after subscribe: want 1, got %d", got)
	}
}

func TestHub_Health_ReturnsSnapshot(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()

	h := hub.Health()
	if h.Closed {
		t.Fatal("new hub should not be closed")
	}

	if h.SubscriberCount != 0 {
		t.Fatalf("want 0 subscribers, got %d", h.SubscriberCount)
	}

	if h.BufferSize != 64 {
		t.Fatalf("want default buffer 64, got %d", h.BufferSize)
	}
}

func TestHub_WithBufferSize(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub(realtime.WithBufferSize(128))

	h := hub.Health()
	if h.BufferSize != 128 {
		t.Fatalf("want buffer 128, got %d", h.BufferSize)
	}
}

func TestHub_Shutdown_DrainsGracefully(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := hub.Shutdown(ctx)
	if err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	h := hub.Health()
	if !h.Closed {
		t.Fatal("hub should be closed after shutdown")
	}
}

func TestHub_Close_InstantShutdown(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()
	hub.Close()

	h := hub.Health()
	if !h.Closed {
		t.Fatal("hub should be closed after Close()")
	}
}

func TestHub_OnSubscribe_OnUnsubscribe(t *testing.T) {
	t.Parallel()

	var subCount, unsubCount atomic.Int32

	hub := realtime.NewHub(
		realtime.WithOnSubscribe(func() { subCount.Add(1) }),
		realtime.WithOnUnsubscribe(func() { unsubCount.Add(1) }),
	)

	ch := hub.Broadcaster.Subscribe()
	hub.Broadcaster.Unsubscribe(ch)

	if got := subCount.Load(); got != 1 {
		t.Fatalf("OnSubscribe: want 1, got %d", got)
	}

	if got := unsubCount.Load(); got != 1 {
		t.Fatalf("OnUnsubscribe: want 1, got %d", got)
	}
}

// --- Handler / Mount tests ---

func TestHandler_ServesEvents(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()
	mux := http.NewServeMux()
	realtime.Mount(mux, "GET /events", hub, realtime.WithHeartbeat(0))

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := httpGetURL(t, server.URL+"/events")
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	waitForSubscriber(t, hub, 1)

	hub.Broadcast(sse.Event{Event: "greeting", Data: "hello world"})

	evt := ssetest.MustReadNEvents(t, resp.Body, 1)[0]

	if evt.Type != "greeting" {
		t.Fatalf("missing event type, got %q", evt.Type)
	}

	if evt.Data() != "hello world" {
		t.Fatalf("missing data, got %q", evt.Data())
	}
}

func TestHandler_CORSHeader(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()
	mux := http.NewServeMux()
	realtime.Mount(mux, "GET /events", hub, realtime.WithHeartbeat(0))

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := httpGetURL(t, server.URL+"/events")
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("CORS: want *, got %q", got)
	}
}

func TestHandler_CustomCORS(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()
	mux := http.NewServeMux()
	realtime.Mount(mux, "GET /events", hub,
		realtime.WithCORSOrigin("https://app.example.com"),
		realtime.WithHeartbeat(0),
	)

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := httpGetURL(t, server.URL+"/events")
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Fatalf("CORS: want https://app.example.com, got %q", got)
	}
}

func TestHandler_FanOut_MultipleClients(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()
	mux := http.NewServeMux()
	realtime.Mount(mux, "GET /events", hub, realtime.WithHeartbeat(0))

	server := httptest.NewServer(mux)
	defer server.Close()

	resp1, err := httpGetURL(t, server.URL+"/events")
	if err != nil {
		t.Fatalf("client 1 connect: %v", err)
	}

	defer func() { _ = resp1.Body.Close() }()

	resp2, err := httpGetURL(t, server.URL+"/events")
	if err != nil {
		t.Fatalf("client 2 connect: %v", err)
	}

	defer func() { _ = resp2.Body.Close() }()

	waitForSubscriber(t, hub, 2)

	hub.Broadcast(sse.Event{Event: "broadcast", Data: "both"})

	evt1 := ssetest.MustReadNEvents(t, resp1.Body, 1)[0]
	evt2 := ssetest.MustReadNEvents(t, resp2.Body, 1)[0]

	if evt1.Data() != "both" {
		t.Fatalf("client 1 missing data:\n%s", evt1.Data())
	}

	if evt2.Data() != "both" {
		t.Fatalf("client 2 missing data:\n%s", evt2.Data())
	}
}

func TestHandler_Filter_OnlyMatchingEvents(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()
	mux := http.NewServeMux()
	realtime.Mount(mux, "GET /events", hub,
		realtime.WithFilter(func(evt sse.Event) bool {
			return evt.Event == "important"
		}),
		realtime.WithHeartbeat(0),
	)

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := httpGetURL(t, server.URL+"/events")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	waitForSubscriber(t, hub, 1)

	hub.Broadcast(sse.Event{Event: "noise", Data: "ignored"})
	hub.Broadcast(sse.Event{Event: "important", Data: "delivered"})

	evt := ssetest.MustReadNEvents(t, resp.Body, 1)[0]

	if evt.Data() == "ignored" {
		t.Fatalf("filtered event leaked through:\n%s", evt.Data())
	}

	if evt.Data() != "delivered" {
		t.Fatalf("matching event missing, got:\n%s", evt.Data())
	}
}

func TestHandler_Replay_OnReconnect(t *testing.T) {
	t.Parallel()

	store := &memStore{
		events: []sse.Event{
			{Event: "old", Data: "1", ID: mustParseID("1")},
			{Event: "old", Data: "2", ID: mustParseID("2")},
			{Event: "old", Data: "3", ID: mustParseID("3")},
		},
	}

	hub := realtime.NewHub(realtime.WithStore(store))
	mux := http.NewServeMux()
	realtime.Mount(mux, "GET /events", hub, realtime.WithHeartbeat(0))

	server := httptest.NewServer(mux)
	defer server.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/events", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	req.Header.Set("Last-Event-ID", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	events := ssetest.MustReadNEvents(t, resp.Body, 2)
	if events[0].Data() != "2" {
		t.Fatalf("first replayed event should be data: 2, got:\n%s", events[0].Data())
	}

	if events[1].Data() != "3" {
		t.Fatalf("second replayed event should be data: 3, got:\n%s", events[1].Data())
	}
}

func TestHandler_NoReplay_WithoutStore(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()
	mux := http.NewServeMux()
	realtime.Mount(mux, "GET /events", hub, realtime.WithHeartbeat(0))

	server := httptest.NewServer(mux)
	defer server.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/events", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	req.Header.Set("Last-Event-ID", "42")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	waitForSubscriber(t, hub, 1)

	hub.Broadcast(sse.Event{Event: "live", Data: "now"})

	evt := ssetest.MustReadNEvents(t, resp.Body, 1)[0]
	if evt.Data() != "now" {
		t.Fatalf("live event should arrive, got:\n%s", evt.Data())
	}
}

func TestHandler_Heartbeat_Sent(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()
	mux := http.NewServeMux()
	realtime.Mount(mux, "GET /events", hub,
		realtime.WithHeartbeat(50*time.Millisecond),
	)

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := httpGetURL(t, server.URL+"/events")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	// Read raw bytes — heartbeat is a comment frame that readSSEFrame skips.
	// We read directly to detect it.
	buf := make([]byte, 0, 64)
	var tmp [1]byte

	for {
		n, err := resp.Body.Read(tmp[:])
		if n > 0 {
			buf = append(buf, tmp[:n]...)
			if len(buf) >= 2 && buf[len(buf)-1] == '\n' && buf[len(buf)-2] == '\n' {
				break
			}
		}

		if err != nil {
			t.Fatalf("read error before heartbeat: %v", err)
		}
	}

	if got := string(buf); !strings.Contains(got, "heartbeat") {
		t.Fatalf("expected heartbeat frame, got: %q", got)
	}
}

func TestHandler_ReturnsHTTPHandler(t *testing.T) {
	t.Parallel()

	hub := realtime.NewHub()
	h := realtime.Handler(hub, realtime.WithHeartbeat(0))

	if h == nil {
		t.Fatal("Handler returned nil")
	}

	server := httptest.NewServer(h)
	defer server.Close()

	resp, err := httpGetURL(t, server.URL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != sse.ContentType {
		t.Fatalf("Content-Type: want %q, got %q", sse.ContentType, ct)
	}
}

// blockingStore delays EventsAfter until released, simulating a slow store
// read during which live events race the replay snapshot.
type blockingStore struct {
	inner   sse.EventStore
	release chan struct{}
}

func (s *blockingStore) EventsAfter(lastID sse.EventID) ([]sse.Event, error) {
	<-s.release

	return s.inner.EventsAfter(lastID) //nolint:wrapcheck // store pass-through
}

// TestHandler_SubscribeBeforeReplay_ClosesGap proves the handler subscribes
// before reading the replay store: an event broadcast while the store read is
// blocked (and therefore NOT in the replay snapshot) still reaches the client
// after the replayed events. The old replay-then-subscribe order lost it.
func TestHandler_SubscribeBeforeReplay_ClosesGap(t *testing.T) {
	t.Parallel()

	store := &blockingStore{
		inner: &memStore{events: []sse.Event{
			{Event: "old", Data: "1", ID: mustParseID("1")},
			{Event: "old", Data: "2", ID: mustParseID("2")},
			{Event: "old", Data: "3", ID: mustParseID("3")},
		}},
		release: make(chan struct{}),
	}

	hub := realtime.NewHub(realtime.WithStore(store))
	mux := http.NewServeMux()
	realtime.Mount(mux, "GET /events", hub, realtime.WithHeartbeat(0))

	server := httptest.NewServer(mux)
	defer server.Close()

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, server.URL+"/events", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	req.Header.Set("Last-Event-ID", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	// The handler subscribes before reading the store: once the subscriber
	// count is 1 while the store read is still blocked, any broadcast is
	// buffered by the channel, not lost.
	waitForSubscriber(t, hub, 1)

	hub.Broadcast(sse.Event{Event: "live", Data: "gap", ID: mustParseID("5")})

	close(store.release)

	events := ssetest.MustReadNEvents(t, resp.Body, 3)
	if events[0].Data() != "2" {
		t.Fatalf("first replayed event should be data: 2, got:\n%s", events[0].Data())
	}

	if events[1].Data() != "3" {
		t.Fatalf("second replayed event should be data: 3, got:\n%s", events[1].Data())
	}

	if events[2].Data() != "gap" {
		t.Fatalf("event broadcast during store read should be forwarded, got:\n%s", events[2].Data())
	}
}

// TestHandler_ReplayLiveDedup proves events that are both replayed and
// buffered (appended after subscribe, snapshot-read before) are delivered
// exactly once: the ID sent during replay is skipped in the live loop.
func TestHandler_ReplayLiveDedup(t *testing.T) {
	t.Parallel()

	store := &blockingStore{
		inner: &memStore{events: []sse.Event{
			{Event: "old", Data: "1", ID: mustParseID("1")},
			{Event: "old", Data: "2", ID: mustParseID("2")},
			{Event: "old", Data: "3", ID: mustParseID("3")},
		}},
		release: make(chan struct{}),
	}

	hub := realtime.NewHub(realtime.WithStore(store))
	mux := http.NewServeMux()
	realtime.Mount(mux, "GET /events", hub, realtime.WithHeartbeat(0))

	server := httptest.NewServer(mux)
	defer server.Close()

	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodGet, server.URL+"/events", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	req.Header.Set("Last-Event-ID", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	waitForSubscriber(t, hub, 1)

	// ID 2 is in the store (will be replayed) AND broadcast now (buffers in
	// the live channel) — the overlap case. ID 7 is live-only.
	hub.Broadcast(sse.Event{Event: "old", Data: "2", ID: mustParseID("2")})
	hub.Broadcast(sse.Event{Event: "live", Data: "7", ID: mustParseID("7")})

	close(store.release)

	frames := make([]string, 0, 3)
	for range 3 {
		frames = append(frames, readSSEFrame(t, resp.Body))
	}

	joined := strings.Join(frames, "")
	if got := strings.Count(joined, "id: 2"); got != 1 {
		t.Fatalf("replayed+buffered event should be delivered exactly once, got %d:\n%s", got, joined)
	}

	for _, want := range []string{"data: 2", "data: 3", "data: 7"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected %s among delivered frames:\n%s", want, joined)
		}
	}
}
