package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"watchcow/internal/docker"
)

const (
	// Default output directory for generated fnOS app packages
	defaultOutputDir = "/tmp/watchcow-apps"
)

func main() {
	// Parse command line flags
	outputDir := flag.String("output", defaultOutputDir, "Output directory for generated fnOS app packages")
	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	// Configure slog
	var logLevel slog.Level
	if *debug {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("WatchCow - fnOS App Generator for Docker")
	slog.Info("========================================")
	slog.Info("Configuration",
		"outputDir", *outputDir,
		"debug", *debug)

	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create and start Docker monitor
	monitor, err := docker.NewMonitor(*outputDir)
	if err != nil {
		slog.Error("Failed to create Docker monitor", "error", err)
		os.Exit(1)
	}
	defer monitor.Stop()

	// Start monitoring
	go monitor.Start(ctx)

	slog.Info("Monitoring started (Press Ctrl+C to stop)")
	slog.Info("")
	slog.Info("To enable fnOS app generation for a container, add these labels:")
	slog.Info("  watchcow.enable: \"true\"")
	slog.Info("  watchcow.display_name: \"Your App Name\"")
	slog.Info("  watchcow.service_port: \"8080\"")
	slog.Info("")
	slog.Info("Optional labels (following fnOS manifest conventions):")
	slog.Info("  watchcow.appname, watchcow.version, watchcow.desc, watchcow.maintainer")
	slog.Info("")

	// Wait for shutdown signal
	<-sigChan
	slog.Info("Shutting down...")
}
