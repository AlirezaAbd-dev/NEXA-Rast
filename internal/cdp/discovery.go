// Package cdp discovers Antigravity 2.0 renderer targets over Chrome DevTools Protocol
// and injects the RTL payload into active pages.
package cdp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Target represents a single debuggable Chromium target returned by /json.
type Target struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"`
	Title                string `json:"title"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// TargetDiscoverer lists active CDP targets eligible for script injection.
type TargetDiscoverer interface {
	Discover(ctx context.Context) ([]Target, error)
}

// HTTPDiscoverer queries Electron's local CDP HTTP endpoint (/json).
type HTTPDiscoverer struct {
	baseURL string
	client  *http.Client
}

// NewHTTPDiscoverer creates an HTTPDiscoverer targeting http://127.0.0.1:<port>.
func NewHTTPDiscoverer(port int) *HTTPDiscoverer {
	return NewHTTPDiscovererWithURL(fmt.Sprintf("http://127.0.0.1:%d", port))
}

// NewHTTPDiscovererWithURL creates an HTTPDiscoverer for a custom base URL (useful in tests).
func NewHTTPDiscovererWithURL(baseURL string) *HTTPDiscoverer {
	return &HTTPDiscoverer{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

// Discover fetches /json and returns eligible page/webview targets.
func (d *HTTPDiscoverer) Discover(ctx context.Context) ([]Target, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.baseURL+"/json", nil)
	if err != nil {
		return nil, fmt.Errorf("build cdp discovery request: %w", err)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query cdp targets: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cdp /json returned status %d", resp.StatusCode)
	}

	var raw []Target
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode cdp targets: %w", err)
	}

	targets := make([]Target, 0, len(raw))
	for _, t := range raw {
		if isEligibleTarget(t) {
			targets = append(targets, t)
		}
	}
	return targets, nil
}

func isEligibleTarget(t Target) bool {
	if t.ID == "" || t.WebSocketDebuggerURL == "" {
		return false
	}
	if strings.HasPrefix(t.URL, "devtools://") || strings.HasPrefix(t.URL, "chrome-extension://") {
		return false
	}
	return t.Type == "page" || t.Type == "webview"
}
