package cdp

import (
	"context"
	"log/slog"
	"time"
)

// Sidecar watches CDP targets and injects the RTL payload into any new or restarted windows.
type Sidecar struct {
	discoverer   TargetDiscoverer
	injector     Injector
	payload      string
	pollInterval time.Duration
	logger       *slog.Logger
	injected     map[string]string
}

// NewSidecar constructs a Sidecar with the given dependencies and polling interval.
func NewSidecar(discoverer TargetDiscoverer, injector Injector, payload string, pollInterval time.Duration, logger *slog.Logger) *Sidecar {
	if logger == nil {
		logger = slog.Default()
	}
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	return &Sidecar{
		discoverer:   discoverer,
		injector:     injector,
		payload:      payload,
		pollInterval: pollInterval,
		logger:       logger,
		injected:     make(map[string]string),
	}
}

// Run polls for CDP targets until ctx is canceled.
func (s *Sidecar) Run(ctx context.Context) error {
	s.SyncOnce(ctx)

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.SyncOnce(ctx)
		}
	}
}

// SyncOnce discovers current targets, injects into any unseen or changed targets,
// and prunes targets that are no longer present.
func (s *Sidecar) SyncOnce(ctx context.Context) {
	targets, err := s.discoverer.Discover(ctx)
	if err != nil {
		// Electron may still be starting up or launched without debugging port; stay quiet at debug level.
		s.logger.Debug("cdp discovery unavailable", "err", err)
		return
	}

	active := make(map[string]struct{}, len(targets))
	for _, t := range targets {
		active[t.ID] = struct{}{}
		if prevWS, ok := s.injected[t.ID]; ok && prevWS == t.WebSocketDebuggerURL {
			continue
		}
		if err := s.injector.Inject(ctx, t, s.payload); err != nil {
			s.logger.Warn("failed to inject rtl payload", "target", t.ID, "err", err)
			continue
		}
		s.injected[t.ID] = t.WebSocketDebuggerURL
		s.logger.Info("injected rtl payload", "target", t.ID, "title", t.Title)
	}

	for id := range s.injected {
		if _, ok := active[id]; !ok {
			delete(s.injected, id)
		}
	}
}
