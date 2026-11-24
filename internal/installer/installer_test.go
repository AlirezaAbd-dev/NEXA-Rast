package installer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPatchDesktopFile(t *testing.T) {
	tests := []struct {
		name  string
		input string
		port  int
		want  string
	}{
		{
			name:  "inserts flag before %U",
			input: "[Desktop Entry]\nExec=env BAMF_DESKTOP_FILE_HINT=/var/lib/snapd/desktop/applications/antigravity_antigravity.desktop /snap/bin/antigravity %U\n",
			port:  9222,
			want:  "Exec=env BAMF_DESKTOP_FILE_HINT=/var/lib/snapd/desktop/applications/antigravity_antigravity.desktop /snap/bin/antigravity --remote-debugging-port=9222 %U",
		},
		{
			name:  "updates existing debugging port",
			input: "[Desktop Entry]\nExec=/usr/bin/antigravity --remote-debugging-port=9000 %F\n",
			port:  9333,
			want:  "Exec=/usr/bin/antigravity --remote-debugging-port=9333 %F",
		},
		{
			name:  "appends flag when no placeholder token exists",
			input: "[Desktop Entry]\nExec=/usr/bin/antigravity\n",
			port:  9222,
			want:  "Exec=/usr/bin/antigravity --remote-debugging-port=9222",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := patchDesktopFile(tc.input, tc.port)
			if !strings.Contains(got, tc.want) {
				t.Errorf("patchDesktopFile() =\n%s\nwant substring:\n%s", got, tc.want)
			}
		})
	}
}

func TestManager_InstallUninstallStatus(t *testing.T) {
	tmp := t.TempDir()
	homeDir := filepath.Join(tmp, "home")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a fake source binary and a fake snap .desktop file.
	binSrc := filepath.Join(tmp, "nexa-rast")
	if err := os.WriteFile(binSrc, []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	desktopSrc := filepath.Join(tmp, "antigravity_antigravity.desktop")
	if err := os.WriteFile(desktopSrc, []byte("[Desktop Entry]\nName=Antigravity\nExec=/snap/bin/antigravity %U\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	launcher := &LinuxLauncher{
		homeDir:    homeDir,
		candidates: []string{desktopSrc},
	}
	mgr := NewManager(Config{
		HomeDir:    homeDir,
		BinaryPath: binSrc,
		Port:       9222,
	}, launcher, nil)

	ctx := context.Background()
	res, err := mgr.Install(ctx)
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(res.PluginDir, "plugin.json")); err != nil {
		t.Fatalf("missing plugin.json: %v", err)
	}
	if _, err := os.Stat(filepath.Join(res.PluginDir, "sidecars", "rtl-injector", "sidecar.json")); err != nil {
		t.Fatalf("missing sidecar.json: %v", err)
	}
	if _, err := os.Stat(res.BinaryDest); err != nil {
		t.Fatalf("missing installed binary: %v", err)
	}

	status := mgr.Status(ctx)
	if !status.PluginInstalled {
		t.Error("expected PluginInstalled = true")
	}
	if !status.Launcher.Configured {
		t.Errorf("expected Launcher.Configured = true, got %+v", status.Launcher)
	}

	if err := mgr.Uninstall(ctx); err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}
	if _, err := os.Stat(res.PluginDir); !os.IsNotExist(err) {
		t.Fatalf("expected plugin dir removed after uninstall")
	}
	if _, err := os.Stat(res.Launcher.UpdatedPath); !os.IsNotExist(err) {
		t.Fatalf("expected desktop override removed after uninstall")
	}
}
