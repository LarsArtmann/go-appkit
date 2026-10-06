package cqrs

// The ServeSSE read-model stream: a dispatched command flows through the
// live pipeline (command -> journal -> projection host -> metaengine
// collection) and out to an HTTP client as a Server-Sent Event.

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/larsartmann/go-cqrs-lite/command/v4"
	"github.com/larsartmann/go-cqrs-lite/decider/v4"
	"github.com/larsartmann/go-cqrs-lite/event/v4"
	"github.com/larsartmann/go-cqrs-lite/id/v4"
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

type taskCommandState struct {
	Count int
}

var taskCommandDecider = decider.Decider[taskCommandState]{
	Initial: taskCommandState{},
	Apply: func(state taskCommandState, evt event.Event) (taskCommandState, error) {
		if evt.Type() == "task.created" {
			state.Count++
		}

		return state, nil
	},
}

// registerTaskCommand wires a dispatchable task.create command whose handler
// emits the domain test's task.created event, so live dispatches reach the
// declared QuerySet.
func registerTaskCommand(t *testing.T, eventSvc *EventService) {
	t.Helper()

	err := RegisterDecider(eventSvc, "Tasks", taskCommandDecider)
	if err != nil {
		t.Fatalf("RegisterDecider: %v", err)
	}

	err = RegisterCommand[*command.BasicCommand, taskCommandState](eventSvc, "task.create",
		func(ctx context.Context, cmd *command.BasicCommand) system.Op[taskCommandState] {
			return system.Execute(ctx, cmd.StreamID(), "Tasks",
				func(_ taskCommandState, ver event.Version) ([]event.Event, error) {
					evt, evtErr := event.New("task.created", cmd.StreamID(), "Tasks", ver+1,
						TaskCreated{
							ID:       cmd.StreamID().String(),
							Title:    "Streamed live",
							Status:   "active",
							Priority: 1,
						})
					if evtErr != nil {
						return nil, evtErr //nolint:wrapcheck // test boundary
					}

					return []event.Event{evt}, nil
				})
		})
	if err != nil {
		t.Fatalf("RegisterCommand: %v", err)
	}
}

func TestEventService_DomainQuerySet_ServeSSEStreamsLiveChanges(t *testing.T) {
	t.Parallel()

	eventSvc := newDomainService(t)
	registerTaskCommand(t, eventSvc)

	err := eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("StartProjections: %v", err)
	}

	watcher := metaengine.NewWatcher[TaskView](eventSvc.System().MetaEngine(), "tasks")
	defer watcher.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = metaengine.ServeSSE(w, r, watcher) // stream ends with the request
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatalf("connect to SSE stream: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on the SSE stream, got %d", resp.StatusCode)
	}

	cmd, err := command.New("task.create", id.NewStreamID())
	if err != nil {
		t.Fatalf("create command: %v", err)
	}

	err = eventSvc.Dispatch(context.Background(), cmd)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	scanner := bufio.NewScanner(resp.Body)

	found := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") && strings.Contains(line, "Streamed live") {
			found = true

			break
		}
	}

	scanErr := scanner.Err()
	if scanErr != nil && !found {
		t.Fatalf("reading the SSE stream: %v", scanErr)
	}

	if !found {
		t.Fatal("the dispatched task never arrived on the SSE stream")
	}
}

// readSSEStream connects (optionally with Last-Event-ID) and returns the
// data lines plus the last id: observed before the wanted count or the 5s
// deadline. Wire order is data-then-id, so it keeps scanning past the
// wanted data line until its id: follows.
func readSSEStream(t *testing.T, serverURL, lastEventID string, want int) ([]string, string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	if lastEventID != "" {
		req.Header.Set("Last-Event-ID", lastEventID)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("connect to SSE stream: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	var dataLines []string

	var lastID string

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		if data, ok := strings.CutPrefix(scanner.Text(), "data:"); ok {
			dataLines = append(dataLines, strings.TrimSpace(data))
		}

		if idLine, ok := strings.CutPrefix(scanner.Text(), "id:"); ok {
			lastID = strings.TrimSpace(idLine)
		}

		if len(dataLines) == want && lastID != "" {
			break
		}
	}

	if len(dataLines) != want {
		t.Fatalf("expected %d SSE data lines, got %d", want, len(dataLines))
	}

	if lastID == "" {
		t.Fatal("the replay stream must carry id: fields for reconnection")
	}

	return dataLines, lastID
}

// The reconnection contract of the replay journal: a client that connects
// fresh receives the full journal, and a client that reconnects with the
// last received Last-Event-ID receives exactly the missed suffix — never
// the already-delivered prefix.
func TestEventService_DomainQuerySet_ServeSSEReplaysMissedEventsOnReconnect(t *testing.T) {
	t.Parallel()

	eventSvc := newDomainService(t)
	registerTaskCommand(t, eventSvc)

	err := eventSvc.StartProjections(context.Background())
	if err != nil {
		t.Fatalf("StartProjections: %v", err)
	}

	watcher := metaengine.NewWatcher[TaskView](eventSvc.System().MetaEngine(), "tasks")
	defer watcher.Close()

	replay := watcher.WithReplay(64) // enable Last-Event-ID reconnection

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = metaengine.ServeSSE(w, r, watcher)
	}))
	defer server.Close()

	firstID := id.NewStreamID()
	secondID := id.NewStreamID()

	for _, streamID := range []id.StreamID{firstID, secondID} {
		cmd, cmdErr := command.New("task.create", streamID)
		if cmdErr != nil {
			t.Fatalf("create command: %v", cmdErr)
		}

		dispatchErr := eventSvc.Dispatch(context.Background(), cmd)
		if dispatchErr != nil {
			t.Fatalf("dispatch: %v", dispatchErr)
		}
	}

	waitFor(t, "both events recorded in the replay journal", func() bool {
		return replay.LatestSeq() == 2
	})

	full, fullLastID := readSSEStream(t, server.URL, "", 2)
	if !strings.Contains(full[0], firstID.String()) || !strings.Contains(full[1], secondID.String()) {
		t.Errorf("fresh connection must replay the journal in dispatch order, got: %v", full)
	}

	missed, _ := readSSEStream(t, server.URL, fullLastID, 1)
	if !strings.Contains(missed[0], secondID.String()) || strings.Contains(missed[0], firstID.String()) {
		t.Errorf("reconnect must replay exactly the missed suffix, got: %v", missed)
	}
}
