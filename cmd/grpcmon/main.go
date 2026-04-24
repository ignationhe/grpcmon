// Command grpcmon is a lightweight terminal dashboard for monitoring
// live gRPC service health and latency metrics.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/yourorg/grpcmon/internal/config"
	"github.com/yourorg/grpcmon/internal/probe"
	"github.com/yourorg/grpcmon/internal/ui"
)

const (
	// defaultConfigPath is the default location for the configuration file.
	defaultConfigPath = "grpcmon.yaml"
	// defaultHistorySize is the number of historical probe results to retain per target.
	defaultHistorySize = 60
)

func main() {
	cfgPath := flag.String("config", defaultConfigPath, "path to YAML config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grpcmon: failed to load config: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.Targets) == 0 {
		fmt.Fprintln(os.Stderr, "grpcmon: no targets defined in config")
		os.Exit(1)
	}

	// Build per-target history stores.
	histories := make(map[string]*probe.History, len(cfg.Targets))
	for _, t := range cfg.Targets {
		histories[t.Address] = probe.NewHistory(defaultHistorySize)
	}

	// Create the tview application and dashboard.
	app := tview.NewApplication()
	dashboard := ui.New(cfg.Targets, histories)

	// Wire up the poller: on each tick, update histories and refresh the UI.
	poller := probe.NewPoller(cfg.Targets, time.Duration(cfg.Interval))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
		app.Stop()
	}()

	// Start polling in the background.
	go func() {
		poller.Run(ctx, func(results []probe.Result, pollTime time.Time) {
			for _, r := range results {
				if h, ok := histories[r.Address]; ok {
					h.Add(r)
				}
			}
			// Queue a UI update on the tview event loop.
			app.QueueUpdateDraw(func() {
				dashboard.Update(results, pollTime)
			})
		})
	}()

	// Configure the root primitive and run the terminal UI.
	root := dashboard.Root()
	app.SetRoot(root, true).SetFocus(root)

	// Allow 'q' or Escape to quit.
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			cancel()
			app.Stop()
			return nil
		}
		if event.Rune() == 'q' {
			cancel()
			app.Stop()
			return nil
		}
		return event
	})

	if err := app.Run(); err != nil {
		log.Fatalf("grpcmon: terminal UI error: %v", err)
	}
}
