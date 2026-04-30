package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/procwatch/internal/alert"
	"github.com/user/procwatch/internal/config"
	"github.com/user/procwatch/internal/monitor"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("procwatch: failed to load config: %v", err)
	}

	sender := alert.NewSender(cfg.WebhookURL, 10*time.Second)
	watcher := monitor.NewWatcher(cfg, sender)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	watcher.Run(ctx)
}
