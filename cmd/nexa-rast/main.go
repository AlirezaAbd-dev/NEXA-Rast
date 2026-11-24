package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AlirezaAbd-dev/NEXA-Rast/internal/assets"
	"github.com/AlirezaAbd-dev/NEXA-Rast/internal/cdp"
	"github.com/AlirezaAbd-dev/NEXA-Rast/internal/installer"
)

const version = "0.1.0"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "nexa-rast: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch args[0] {
	case "install":
		return runInstall(ctx, args[1:])
	case "uninstall":
		return runUninstall(ctx, args[1:])
	case "status":
		return runStatus(ctx, args[1:])
	case "sidecar":
		return runSidecar(ctx, args[1:])
	case "version", "--version", "-v":
		fmt.Printf("nexa-rast %s\n", version)
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func newManager(port int) (*installer.Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home dir: %w", err)
	}
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("resolve executable path: %w", err)
	}
	launcher := installer.NewOSLauncher(homeDir)
	discoverer := cdp.NewHTTPDiscoverer(port)
	return installer.NewManager(installer.Config{
		HomeDir:    homeDir,
		BinaryPath: exePath,
		Port:       port,
	}, launcher, discoverer), nil
}

func parsePortFlag(name string, args []string) (int, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	port := fs.Int("port", 9222, "Chrome DevTools Protocol port")
	if err := fs.Parse(args); err != nil {
		return 0, err
	}
	if *port <= 0 || *port > 65535 {
		return 0, fmt.Errorf("invalid port %d", *port)
	}
	return *port, nil
}

func runInstall(ctx context.Context, args []string) error {
	port, err := parsePortFlag("install", args)
	if err != nil {
		return err
	}
	mgr, err := newManager(port)
	if err != nil {
		return err
	}
	res, err := mgr.Install(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Installed nexa-rast plugin to %s\n", res.PluginDir)
	if res.Launcher.UpdatedPath != "" {
		fmt.Printf("Configured launcher override at %s (--remote-debugging-port=%d)\n", res.Launcher.UpdatedPath, port)
	}
	if res.Launcher.ManualHint != "" {
		fmt.Println(res.Launcher.ManualHint)
	}
	fmt.Println("Restart Antigravity 2.0 from your app launcher to activate RTL injection.")
	return nil
}

func runUninstall(ctx context.Context, args []string) error {
	port, err := parsePortFlag("uninstall", args)
	if err != nil {
		return err
	}
	mgr, err := newManager(port)
	if err != nil {
		return err
	}
	if err := mgr.Uninstall(ctx); err != nil {
		return err
	}
	fmt.Println("Uninstalled nexa-rast plugin and restored launcher configuration.")
	return nil
}

func runStatus(ctx context.Context, args []string) error {
	port, err := parsePortFlag("status", args)
	if err != nil {
		return err
	}
	mgr, err := newManager(port)
	if err != nil {
		return err
	}
	st := mgr.Status(ctx)
	fmt.Printf("Plugin installed: %v (%s)\n", st.PluginInstalled, st.PluginDir)
	fmt.Printf("Launcher configured: %v (%s)\n", st.Launcher.Configured, st.Launcher.Detail)
	fmt.Printf("CDP reachable (port %d): %v (targets: %d)\n", port, st.CDPReachable, st.ActiveTargets)
	return nil
}

func runSidecar(ctx context.Context, args []string) error {
	port, err := parsePortFlag("sidecar", args)
	if err != nil {
		return err
	}
	payload, err := assets.PayloadJS()
	if err != nil {
		return fmt.Errorf("load rtl payload: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	sidecar := cdp.NewSidecar(
		cdp.NewHTTPDiscoverer(port),
		cdp.NewWSInjector(),
		payload,
		2*time.Second,
		logger,
	)

	err = sidecar.Run(ctx)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func printUsage() {
	fmt.Print(`nexa-rast - RTL (Persian/Arabic/Hebrew) chat fix for Antigravity 2.0

Usage:
  nexa-rast <command> [--port 9222]

Commands:
  install     Install the Antigravity plugin & configure launcher debugging port
  uninstall   Remove the plugin & restore the launcher configuration
  status      Show plugin, launcher, and CDP connection status
  sidecar     Run the background CDP RTL injector (started by Antigravity)
  version     Print version information
`)
}
