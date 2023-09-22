package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/nikonok/backupper/controllers"
	"github.com/nikonok/backupper/initial"
	log "github.com/nikonok/backupper/logger"
	"github.com/nikonok/backupper/watchers"
	"github.com/nikonok/backupper/workers"
)

const (
	workChanSize = 100
)

type FileInfo struct {
	FileName         string
	ModificationTime time.Time
	IsProcessed      bool
}

type AppConfig struct {
	HotFolderPath    string
	BackupFolderPath string
	LoggerFilePath   string

	TickerDuration time.Duration

	FileCollection map[string]FileInfo

	CopyWork chan string
}

func main() {
	// TODO: remove
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// TODO: normal config
	appCfg := AppConfig{
		HotFolderPath:    "./hot",
		BackupFolderPath: "./backup",
		TickerDuration:   1 * time.Second,
		LoggerFilePath:   "./log.txt",
	}

	var err error

	appCfg.HotFolderPath, err = filepath.Abs(appCfg.HotFolderPath)
	if err != nil {
		panic(err)
	}

	appCfg.BackupFolderPath, err = filepath.Abs(appCfg.BackupFolderPath)
	if err != nil {
		panic(err)
	}
	//

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

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		logger.LogInfo("Caught signal: " + sig.String())
		cancel()
	}()

	var wg sync.WaitGroup

	watcherChan := make(chan string, workChanSize)
	copyChan := make(chan string, workChanSize)
	deleteChan := make(chan string, workChanSize)

	controller := controllers.CreateController(watcherChan, copyChan, deleteChan, logger)
	resultChan := controller.GetResultChan()

	wg.Add(1)
	go func() {
		defer wg.Done()
		controller.Run(ctx)
	}()

	copyWorker := workers.CreateCopyWorker(appCfg.HotFolderPath, appCfg.BackupFolderPath, copyChan, resultChan, logger)
	deleteWorker := workers.CreateDeleteWorker(appCfg.HotFolderPath, appCfg.BackupFolderPath, deleteChan, resultChan, logger)

	workers.StartWorker(copyWorker, ctx, &wg)
	workers.StartWorker(deleteWorker, ctx, &wg)

	initial.CreateInitialChecker(appCfg.HotFolderPath, appCfg.BackupFolderPath, watcherChan, logger).Check(ctx)

	watcher := watchers.CreateSysCallWatcher(appCfg.HotFolderPath, watcherChan, logger)
	watchers.StartWatcher(watcher, ctx, &wg)

	wg.Wait()
}
