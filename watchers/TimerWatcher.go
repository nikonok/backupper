package watchers

import (
	"context"
	"io/fs"
	"os"
	"time"

	log "github.com/nikonok/backupper/logger"
)

type fileInfo struct {
	ModificationTime time.Time
}

type TimerWatcher struct {
	logger log.Logger

	hotFolderPath  string
	workChan       chan string
	tickerDuration time.Duration
	fileCollection map[string]fileInfo

	ticker *time.Ticker
}

func CreateTimerWatcher(hotFolderPath string, workChan chan string, tickerDuration time.Duration, logger log.Logger) Watcher {
	watcher := &TimerWatcher{
		logger:         logger,
		hotFolderPath:  hotFolderPath,
		workChan:       workChan,
		tickerDuration: tickerDuration,
		fileCollection: make(map[string]fileInfo),
	}
	watcher.initialScan()
	return watcher
}

func (watcher *TimerWatcher) initialScan() {
	files, err := os.ReadDir(watcher.hotFolderPath)
	if err != nil {
		watcher.logger.LogError("Watcher fatal: " + err.Error())
	}

	for _, file := range files {
		if file.Type().IsRegular() {
			fInfo, err := file.Info()
			if err != nil {
				watcher.logger.LogError("Watcher fatal: " + err.Error())
			}
			watcher.fileCollection[fInfo.Name()] = fileInfo{
				ModificationTime: fInfo.ModTime(),
			}
		}
	}
}

func (watcher *TimerWatcher) watch(ctx context.Context) {
	watcher.logger.LogDebug("Starting TimerWatcher")

	watcher.ticker = time.NewTicker(watcher.tickerDuration)
	watcher.runLoop(ctx)
}

func (watcher *TimerWatcher) runLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			watcher.logger.LogDebug("Stopping SysCallWatcher")
			return
		case <-watcher.ticker.C:
			watcher.processDir()
		}
	}
}

func (watcher *TimerWatcher) processDir() {
	files, err := os.ReadDir(watcher.hotFolderPath)
	if err != nil {
		watcher.logger.LogError("Watcher fatal: " + err.Error())
	}

	for _, file := range files {
		if file.Type().IsRegular() {
			watcher.processFile(file)
		}
	}
}

func (watcher *TimerWatcher) processFile(file fs.DirEntry) {
	fInfo, err := file.Info()
	if err != nil {
		watcher.logger.LogError("Watcher fatal: " + err.Error())
	}

	if savedFileInfo, ok := watcher.fileCollection[fInfo.Name()]; !ok {
		watcher.fileCollection[fInfo.Name()] = fileInfo{
			ModificationTime: fInfo.ModTime(),
		}

		watcher.logger.LogInfo("Watcher is sending new work for: " + fInfo.Name())
		watcher.workChan <- fInfo.Name()
	} else if savedFileInfo.ModificationTime.Before(fInfo.ModTime()) {
		savedFileInfo.ModificationTime = fInfo.ModTime()
		watcher.fileCollection[fInfo.Name()] = savedFileInfo

		watcher.logger.LogInfo("Watcher is sending new work for: " + fInfo.Name())
		watcher.workChan <- fInfo.Name()
	}
}
