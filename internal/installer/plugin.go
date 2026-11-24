// Package installer manages the Antigravity plugin bundle and OS launcher configuration.
package installer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/AlirezaAbd-dev/NEXA-Rast/internal/cdp"
)

const defaultPluginName = "nexa-rast"

// Launcher configures the OS application launcher so Antigravity starts with CDP enabled.
type Launcher interface {
	Setup(ctx context.Context, port int) (LauncherResult, error)
	Restore(ctx context.Context) error
	Status(ctx context.Context, port int) LauncherStatus
}

// LauncherResult describes what Setup changed or what manual step is needed.
type LauncherResult struct {
	UpdatedPath string
	ManualHint  string
}

// LauncherStatus reports whether the OS launcher is configured for CDP.
type LauncherStatus struct {
	Configured bool
	Detail     string
}

// Config holds paths and options for installing or inspecting the plugin.
type Config struct {
	HomeDir    string
	BinaryPath string
	Port       int
}

// InstallResult summarizes the installed plugin and launcher paths.
type InstallResult struct {
	PluginDir  string
	BinaryDest string
	Launcher   LauncherResult
}

// StatusReport summarizes the current installation and runtime health.
type StatusReport struct {
	PluginInstalled bool
	PluginDir       string
	Launcher        LauncherStatus
	CDPReachable    bool
	ActiveTargets   int
}

// Manager orchestrates plugin installation, uninstallation, and status checks.
type Manager struct {
	cfg        Config
	launcher   Launcher
	discoverer cdp.TargetDiscoverer
}

// NewManager creates a Manager with the provided configuration, launcher, and CDP discoverer.
func NewManager(cfg Config, launcher Launcher, discoverer cdp.TargetDiscoverer) *Manager {
	return &Manager{
		cfg:        cfg,
		launcher:   launcher,
		discoverer: discoverer,
	}
}

func (m *Manager) pluginDir() string {
	return filepath.Join(m.cfg.HomeDir, ".gemini", "config", "plugins", defaultPluginName)
}

// Install copies the binary into the plugin bundle, writes plugin.json and sidecar.json,
// and configures the OS launcher to pass --remote-debugging-port.
func (m *Manager) Install(ctx context.Context) (InstallResult, error) {
	if m.cfg.HomeDir == "" {
		return InstallResult{}, fmt.Errorf("home directory is required")
	}
	if m.cfg.BinaryPath == "" {
		return InstallResult{}, fmt.Errorf("binary path is required")
	}

	pluginDir := m.pluginDir()
	binDir := filepath.Join(pluginDir, "bin")
	sidecarDir := filepath.Join(pluginDir, "sidecars", "rtl-injector")

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return InstallResult{}, fmt.Errorf("create plugin bin dir: %w", err)
	}
	if err := os.MkdirAll(sidecarDir, 0o755); err != nil {
		return InstallResult{}, fmt.Errorf("create sidecar dir: %w", err)
	}

	binDest := filepath.Join(binDir, filepath.Base(m.cfg.BinaryPath))
	if err := copyExecutable(m.cfg.BinaryPath, binDest); err != nil {
		return InstallResult{}, fmt.Errorf("install binary: %w", err)
	}

	if err := writeJSON(filepath.Join(pluginDir, "plugin.json"), map[string]any{
		"name":        defaultPluginName,
		"displayName": "NEXA Rast",
		"version":     "0.1.0",
		"description": "Fixes RTL (Persian, Arabic, Hebrew) rendering and code-block LTR isolation in Antigravity 2.0 chat.",
		"repository":  "https://github.com/AlirezaAbd-dev/NEXA-Rast",
	}); err != nil {
		return InstallResult{}, fmt.Errorf("write plugin.json: %w", err)
	}

	if err := writeJSON(filepath.Join(sidecarDir, "sidecar.json"), map[string]any{
		"display_name":   "NEXA Rast RTL Injector",
		"description":    "Injects CSS-first RTL support into Antigravity 2.0 chat windows over CDP.",
		"command":        binDest,
		"args":           []string{"sidecar", "--port", strconv.Itoa(m.cfg.Port)},
		"restart_policy": "on-failure",
	}); err != nil {
		return InstallResult{}, fmt.Errorf("write sidecar.json: %w", err)
	}

	launcherRes, err := m.launcher.Setup(ctx, m.cfg.Port)
	if err != nil {
		return InstallResult{}, fmt.Errorf("configure launcher: %w", err)
	}

	return InstallResult{
		PluginDir:  pluginDir,
		BinaryDest: binDest,
		Launcher:   launcherRes,
	}, nil
}

// Uninstall removes the plugin directory and restores the OS launcher.
func (m *Manager) Uninstall(ctx context.Context) error {
	if err := os.RemoveAll(m.pluginDir()); err != nil {
		return fmt.Errorf("remove plugin dir: %w", err)
	}
	if err := m.launcher.Restore(ctx); err != nil {
		return fmt.Errorf("restore launcher: %w", err)
	}
	return nil
}

// Status inspects the plugin directory, OS launcher, and live CDP endpoint.
func (m *Manager) Status(ctx context.Context) StatusReport {
	pluginDir := m.pluginDir()
	_, err := os.Stat(filepath.Join(pluginDir, "plugin.json"))

	report := StatusReport{
		PluginInstalled: err == nil,
		PluginDir:       pluginDir,
		Launcher:        m.launcher.Status(ctx, m.cfg.Port),
	}

	if m.discoverer != nil {
		if targets, err := m.discoverer.Discover(ctx); err == nil {
			report.CDPReachable = true
			report.ActiveTargets = len(targets)
		}
	}
	return report
}

func copyExecutable(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source binary: %w", err)
	}
	defer in.Close()

	tmp := dst + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("create temp binary: %w", err)
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("copy binary: %w", err)
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close binary: %w", err)
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename binary: %w", err)
	}
	return nil
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
