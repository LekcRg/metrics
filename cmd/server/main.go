package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/LekcRg/metrics/internal/buildinfo"
	"github.com/LekcRg/metrics/internal/logger"
	"github.com/LekcRg/metrics/internal/server/serverapp"
	"go.uber.org/zap"
)

func exit(cancel context.CancelFunc, app *serverapp.App, exited chan any) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-sigChan
	cancel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	app.Stop(ctx)

	cancel()
	exited <- true
}

func main() {
	buildinfo.Print()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	app, err := serverapp.New(ctx, &wg)
	if err != nil {
		logger.Log.Fatal("Error while init app", zap.Error(err))
	}

	exited := make(chan any, 1)
	go exit(cancel, app, exited)

	app.Start(&wg)

	wg.Wait()
	<-exited
	logger.Log.Info("Buy, 👋!")
}
