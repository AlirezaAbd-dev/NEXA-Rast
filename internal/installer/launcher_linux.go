//go:build linux

package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var debugPortFlagRe = regexp.MustCompile(`--remote-debugging-port=\d+`)

// LinuxLauncher manages XDG .desktop overrides so Antigravity launches with CDP enabled.
type LinuxLauncher struct {
	homeDir    string
	candidates []string
}

// NewOSLauncher returns the Linux XDG .desktop launcher manager.
func NewOSLauncher(homeDir string) Launcher {
	return &LinuxLauncher{
		homeDir: homeDir,
		candidates: []string{
			"/var/lib/snapd/desktop/applications/antigravity_antigravity.desktop",
			"/usr/share/applications/antigravity.desktop",
			filepath.Join(homeDir, ".local", "share", "applications", "antigravity.desktop"),
		},
	}
}

// Setup creates or updates a user-level .desktop entry with --remote-debugging-port=<port>.
func (l *LinuxLauncher) Setup(_ context.Context, port int) (LauncherResult, error) {
	srcPath, raw, err := l.findDesktopSource()
	if err != nil {
		return LauncherResult{
			ManualHint: fmt.Sprintf("Could not locate Antigravity .desktop file automatically. Launch Antigravity with --remote-debugging-port=%d", port),
		}, nil
	}

	destDir := filepath.Join(l.homeDir, ".local", "share", "applications")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return LauncherResult{}, fmt.Errorf("create applications dir: %w", err)
	}

	destPath := filepath.Join(destDir, filepath.Base(srcPath))
	if srcPath == destPath {
		bakPath := destPath + ".nexa-rast.bak"
		if _, err := os.Stat(bakPath); os.IsNotExist(err) {
			if err := os.WriteFile(bakPath, raw, 0o644); err != nil {
				return LauncherResult{}, fmt.Errorf("backup desktop file: %w", err)
			}
		}
	}

	patched := patchDesktopFile(string(raw), port)
	if err := os.WriteFile(destPath, []byte(patched), 0o644); err != nil {
		return LauncherResult{}, fmt.Errorf("write desktop override: %w", err)
	}

	return LauncherResult{UpdatedPath: destPath}, nil
}

// Restore removes the user-level .desktop override or restores its backup.
func (l *LinuxLauncher) Restore(_ context.Context) error {
	destDir := filepath.Join(l.homeDir, ".local", "share", "applications")
	for _, cand := range l.candidates {
		destPath := filepath.Join(destDir, filepath.Base(cand))
		bakPath := destPath + ".nexa-rast.bak"

		if _, err := os.Stat(bakPath); err == nil {
			if err := os.Rename(bakPath, destPath); err != nil {
				return fmt.Errorf("restore desktop backup: %w", err)
			}
			continue
		}

		// Only remove overrides that do not point back to themselves as original source,
		// or strip the flag if it's a user-only file.
		if cand != destPath {
			if err := os.Remove(destPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove desktop override: %w", err)
			}
		}
	}
	return nil
}

// Status checks whether any user .desktop override contains the debugging port flag.
func (l *LinuxLauncher) Status(_ context.Context, port int) LauncherStatus {
	wantFlag := fmt.Sprintf("--remote-debugging-port=%d", port)
	destDir := filepath.Join(l.homeDir, ".local", "share", "applications")
	for _, cand := range l.candidates {
		destPath := filepath.Join(destDir, filepath.Base(cand))
		data, err := os.ReadFile(destPath)
		if err != nil {
			continue
		}
		if strings.Contains(string(data), wantFlag) {
			return LauncherStatus{
				Configured: true,
				Detail:     destPath,
			}
		}
	}
	return LauncherStatus{
		Configured: false,
		Detail:     fmt.Sprintf("no user .desktop override with %s", wantFlag),
	}
}

func (l *LinuxLauncher) findDesktopSource() (string, []byte, error) {
	// Prefer an existing user override if present, otherwise fall back to system/snap entries.
	destDir := filepath.Join(l.homeDir, ".local", "share", "applications")
	for _, cand := range l.candidates {
		userPath := filepath.Join(destDir, filepath.Base(cand))
		if data, err := os.ReadFile(userPath); err == nil {
			return userPath, data, nil
		}
	}
	for _, cand := range l.candidates {
		if data, err := os.ReadFile(cand); err == nil {
			return cand, data, nil
		}
	}
	return "", nil, fmt.Errorf("no antigravity .desktop entry found")
}

func patchDesktopFile(content string, port int) string {
	flag := fmt.Sprintf("--remote-debugging-port=%d", port)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, "Exec=") {
			continue
		}
		if debugPortFlagRe.MatchString(line) {
			lines[i] = debugPortFlagRe.ReplaceAllString(line, flag)
			continue
		}
		lines[i] = insertFlagInExec(line, flag)
	}
	return strings.Join(lines, "\n")
}

func insertFlagInExec(execLine, flag string) string {
	for _, token := range []string{" %U", " %u", " %F", " %f"} {
		if idx := strings.Index(execLine, token); idx != -1 {
			return execLine[:idx] + " " + flag + execLine[idx:]
		}
	}
	return execLine + " " + flag
}
