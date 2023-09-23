package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/nikonok/backupper/controllers"
	"github.com/nikonok/backupper/helpers"
	"github.com/nikonok/backupper/initial"
	log "github.com/nikonok/backupper/logger"
	"github.com/nikonok/backupper/watchers"
	"github.com/nikonok/backupper/workers"
)

func main() {
	appCfg := CreateConfig()

	// Create logger
	logger, err := log.CreateDualLogger(appCfg.LoggerFilePath, log.Debug)
	if err != nil {
		fmt.Println("Cannot init logger")
		panic(err)
	}

	defer func() {
		if err := logger.Close(); err != nil {
			fmt.Println("Cannot close logger")
		}
	}()

	// create context and handle signals
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		logger.LogDebug("Caught signal: " + sig.String())
		logger.LogInfo("Graceful shutdown started")
		cancel()
	}()

	switch appCfg.Mode {
	case helpers.Default:
		runBackup(ctx, appCfg, logger)
	case helpers.ViewLogs:
		runViewLog(appCfg)
	default:
		logger.LogError("Unknown mode selected")
	}
}

func runBackup(ctx context.Context, appCfg *helpers.AppConfig, logger log.Logger) {
	logger.LogDebug("Running backup mode")

	// create WaitGroup and channels
	var wg sync.WaitGroup

	watcherChan := make(chan string, helpers.WORKER_CHAN_SIZE)
	copyChan := make(chan string, helpers.WORKER_CHAN_SIZE)
	deleteChan := make(chan string, helpers.WORKER_CHAN_SIZE)

	logger.LogInfo("Init finished, starting components")

	// create/start controller and get result chan
	controller := controllers.CreateController(watcherChan, copyChan, deleteChan, logger)
	resultChan := controller.GetResultChan()

	wg.Add(1)
	go func() {
		defer wg.Done()
		controller.Run(ctx)
	}()

	// create/start workers
	for i := 0; i < helpers.COPY_WORKERS_AMOUNT; i++ {
		workers.StartWorker(workers.CreateCopyWorker(appCfg, copyChan, resultChan, logger), ctx, &wg)
	}
	for i := 0; i < helpers.DELETE_WORKERS_AMOUNT; i++ {
		workers.StartWorker(workers.CreateDeleteWorker(appCfg, deleteChan, resultChan, logger), ctx, &wg)
	}

	// make initial check of the dir
	initial.CreateInitialChecker(appCfg, watcherChan, logger).Check(ctx)

	// create/start watcher
	var watcher watchers.Watcher
	switch appCfg.Watcher {
	case helpers.SysCallWatcherType:
		watcher = watchers.CreateSysCallWatcher(appCfg.HotFolderPath, watcherChan, logger)
	case helpers.EventWatcherType:
		watcher = watchers.CreateEventWatcher(appCfg.HotFolderPath, watcherChan, logger)
	case helpers.TimerWatcherType:
		watcher = watchers.CreateTimerWatcher(appCfg.HotFolderPath, watcherChan, helpers.TICKER_DURATION, logger)
	}

	watchers.StartWatcher(watcher, ctx, &wg)

	wg.Wait()
}
