package cdp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// Injector delivers a JavaScript payload to a CDP target.
type Injector interface {
	Inject(ctx context.Context, target Target, script string) error
}

// WSInjector connects to a target's WebSocket debugger URL and evaluates the payload
// immediately as well as on future document loads.
type WSInjector struct {
	dialTimeout time.Duration
}

// NewWSInjector returns a WSInjector with a default operation timeout.
func NewWSInjector() *WSInjector {
	return &WSInjector{
		dialTimeout: 5 * time.Second,
	}
}

type cdpRequest struct {
	ID     int            `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

type cdpResponse struct {
	ID    int `json:"id"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Inject opens a short-lived CDP WebSocket session, registers the script for future
// navigations, and evaluates it immediately in the current document.
func (w *WSInjector) Inject(ctx context.Context, target Target, script string) error {
	opCtx, cancel := context.WithTimeout(ctx, w.dialTimeout)
	defer cancel()

	conn, _, err := websocket.Dial(opCtx, target.WebSocketDebuggerURL, nil)
	if err != nil {
		return fmt.Errorf("dial cdp target %s: %w", target.ID, err)
	}
	defer conn.CloseNow()

	commands := []cdpRequest{
		{ID: 1, Method: "Page.enable"},
		{
			ID:     2,
			Method: "Page.addScriptToEvaluateOnNewDocument",
			Params: map[string]any{"source": script},
		},
		{
			ID:     3,
			Method: "Runtime.evaluate",
			Params: map[string]any{
				"expression": script,
				"silent":     true,
			},
		},
	}

	for _, cmd := range commands {
		if err := sendAndAwait(opCtx, conn, cmd); err != nil {
			return fmt.Errorf("cdp %s on target %s: %w", cmd.Method, target.ID, err)
		}
	}

	_ = conn.Close(websocket.StatusNormalClosure, "done")
	return nil
}

func sendAndAwait(ctx context.Context, conn *websocket.Conn, cmd cdpRequest) error {
	if err := wsjson.Write(ctx, conn, cmd); err != nil {
		return fmt.Errorf("write command: %w", err)
	}

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return fmt.Errorf("read response: %w", err)
		}
		var resp cdpResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			continue
		}
		if resp.ID != cmd.ID {
			// Ignore asynchronous CDP events (e.g., Page.frameNavigated).
			continue
		}
		if resp.Error != nil {
			return fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return nil
	}
}
