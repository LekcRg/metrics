package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/LekcRg/metrics/internal/agent/agentapp"
	"github.com/LekcRg/metrics/internal/buildinfo"
	"github.com/LekcRg/metrics/internal/logger"
)

func exit(exited chan any, app *agentapp.App) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-sigChan
	logger.Log.Info("Stopping, wait end of requests")
	app.Stop()

	exited <- true
}

func main() {
	buildinfo.Print()

	var wg sync.WaitGroup
	ctx := context.Background()

	app := agentapp.New()
	app.Start(ctx, &wg)

	exited := make(chan any, 1)
	go exit(exited, app)
	wg.Wait()

	<-exited
	logger.Log.Info("Buy, 👋!")
}
