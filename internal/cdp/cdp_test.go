package cdp

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPDiscoverer_Discover(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantCount  int
		wantID     string
		wantErr    bool
	}{
		{
			name:       "filters eligible page and webview targets",
			statusCode: http.StatusOK,
			body: `[
				{"id":"1","type":"page","title":"Antigravity","url":"file:///app/index.html","webSocketDebuggerUrl":"ws://127.0.0.1:9222/devtools/page/1"},
				{"id":"2","type":"page","title":"DevTools","url":"devtools://devtools/bundled/inspector.html","webSocketDebuggerUrl":"ws://127.0.0.1:9222/devtools/page/2"},
				{"id":"3","type":"service_worker","title":"SW","url":"https://example.com/sw.js","webSocketDebuggerUrl":"ws://127.0.0.1:9222/devtools/page/3"},
				{"id":"4","type":"webview","title":"Chat","url":"https://localhost/chat","webSocketDebuggerUrl":"ws://127.0.0.1:9222/devtools/page/4"}
			]`,
			wantCount: 2,
			wantID:    "1",
		},
		{
			name:       "returns error on non-200 status",
			statusCode: http.StatusInternalServerError,
			body:       `internal error`,
			wantErr:    true,
		},
		{
			name:       "returns error on invalid json",
			statusCode: http.StatusOK,
			body:       `{not-json`,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/json" {
					http.NotFound(w, r)
					return
				}
				w.WriteHeader(tc.statusCode)
				_, _ = io.WriteString(w, tc.body)
			}))
			defer srv.Close()

			d := NewHTTPDiscovererWithURL(srv.URL)
			got, err := d.Discover(context.Background())
			if tc.wantErr {
				if err == nil {
					t.Fatal("Discover() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Discover() unexpected error: %v", err)
			}
			if len(got) != tc.wantCount {
				t.Fatalf("Discover() count = %d, want %d", len(got), tc.wantCount)
			}
			if tc.wantCount > 0 && got[0].ID != tc.wantID {
				t.Errorf("Discover()[0].ID = %q, want %q", got[0].ID, tc.wantID)
			}
		})
	}
}

type fakeDiscoverer struct {
	targets []Target
	err     error
}

func (f *fakeDiscoverer) Discover(_ context.Context) ([]Target, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.targets, nil
}

type fakeInjector struct {
	calls []string
	err   error
}

func (f *fakeInjector) Inject(_ context.Context, target Target, _ string) error {
	if f.err != nil {
		return f.err
	}
	f.calls = append(f.calls, target.ID)
	return nil
}

func TestSidecar_SyncOnce(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("injects once per target ws url and prunes closed targets", func(t *testing.T) {
		disc := &fakeDiscoverer{
			targets: []Target{{ID: "t1", Type: "page", WebSocketDebuggerURL: "ws://1/1"}},
		}
		inj := &fakeInjector{}
		s := NewSidecar(disc, inj, "payload", time.Second, logger)

		s.SyncOnce(context.Background())
		s.SyncOnce(context.Background())
		if len(inj.calls) != 1 {
			t.Fatalf("expected 1 injection for unchanged target, got %d", len(inj.calls))
		}

		// Target disappears and comes back with a new WS URL.
		disc.targets = nil
		s.SyncOnce(context.Background())
		if len(s.injected) != 0 {
			t.Fatalf("expected closed target to be pruned, got %d", len(s.injected))
		}

		disc.targets = []Target{{ID: "t1", Type: "page", WebSocketDebuggerURL: "ws://1/2"}}
		s.SyncOnce(context.Background())
		if len(inj.calls) != 2 {
			t.Fatalf("expected re-injection on new ws url, got %d", len(inj.calls))
		}
	})

	t.Run("retries on next sync after injection failure", func(t *testing.T) {
		disc := &fakeDiscoverer{
			targets: []Target{{ID: "t1", Type: "page", WebSocketDebuggerURL: "ws://1/1"}},
		}
		inj := &fakeInjector{err: errors.New("dial failed")}
		s := NewSidecar(disc, inj, "payload", time.Second, logger)

		s.SyncOnce(context.Background())
		if len(s.injected) != 0 {
			t.Fatal("failed injection should not mark target as injected")
		}

		inj.err = nil
		s.SyncOnce(context.Background())
		if len(inj.calls) != 1 {
			t.Fatalf("expected retry to succeed, got %d calls", len(inj.calls))
		}
	})
}

func TestSidecar_RunStopsOnCancel(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	disc := &fakeDiscoverer{}
	inj := &fakeInjector{}
	s := NewSidecar(disc, inj, "payload", 10*time.Millisecond, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Millisecond)
	defer cancel()

	err := s.Run(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run() err = %v, want context.DeadlineExceeded", err)
	}
}
