//go:build !linux

package installer

import (
	"context"
	"fmt"
	"runtime"
)

type otherLauncher struct{}

// NewOSLauncher returns a platform-specific launcher helper for non-Linux systems.
func NewOSLauncher(_ string) Launcher {
	return &otherLauncher{}
}

func (o *otherLauncher) Setup(_ context.Context, port int) (LauncherResult, error) {
	switch runtime.GOOS {
	case "darwin":
		return LauncherResult{
			ManualHint: fmt.Sprintf("Launch Antigravity with: open -a Antigravity --args --remote-debugging-port=%d", port),
		}, nil
	case "windows":
		return LauncherResult{
			ManualHint: fmt.Sprintf("Add --remote-debugging-port=%d to your Antigravity shortcut Target.", port),
		}, nil
	default:
		return LauncherResult{
			ManualHint: fmt.Sprintf("Launch Antigravity with --remote-debugging-port=%d", port),
		}, nil
	}
}

func (o *otherLauncher) Restore(_ context.Context) error {
	return nil
}

func (o *otherLauncher) Status(_ context.Context, port int) LauncherStatus {
	return LauncherStatus{
		Configured: false,
		Detail:     fmt.Sprintf("manual launch flag required on %s (--remote-debugging-port=%d)", runtime.GOOS, port),
	}
}
