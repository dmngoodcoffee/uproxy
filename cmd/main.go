package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"uproxy/internal/config"
	"uproxy/internal/proxy"

	log "github.com/sirupsen/logrus"
)

func main() {
	conf := config.New()
	pxy := proxy.New(conf)

	if err := proxy.SetProxy(*conf.Port); err != nil {
		log.Fatalf("setting proxy error: %s", err)
	}

	defer func() {
		if err := proxy.UnsetProxy(); err != nil {
			log.Fatal(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go pxy.Start(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	sig := <-signalChan
	log.Printf("Received signal: %s. Shutting down...", sig)

	cancel()
}
