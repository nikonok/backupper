package watchers

import (
	"context"
	"os"

	log "github.com/nikonok/backupper/logger"

	"github.com/fsnotify/fsnotify"
)

type EventWatcher struct {
	logger log.Logger

	hotFolderPath string
	workChan      chan string
	fsWatcher     *fsnotify.Watcher
}

func CreateEventWatcher(hotFolderPath string, workChan chan string, logger log.Logger) Watcher {
	return &EventWatcher{
		logger:        logger,
		hotFolderPath: hotFolderPath,
		workChan:      workChan,
	}
}

func (watcher *EventWatcher) watch(ctx context.Context) {
	watcher.logger.LogDebug("Starting EventWatcher")

	var err error
	watcher.fsWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		watcher.logger.LogError("Watcher fatal: " + err.Error())
	}

	defer watcher.fsWatcher.Close()

	err = watcher.fsWatcher.Add(watcher.hotFolderPath)
	if err != nil {
		watcher.logger.LogError("Watcher fatal: " + err.Error())
	}

	watcher.runLoop(ctx)
}

func (watcher *EventWatcher) runLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			watcher.logger.LogDebug("Stopping EventWatcher")
			return
		case event, ok := <-watcher.fsWatcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
				watcher.processFile(event.Name)
			}
		case err, ok := <-watcher.fsWatcher.Errors:
			if !ok {
				return
			}
			watcher.logger.LogWarn("Watcher warning: " + err.Error())
		}
	}
}

func (watcher *EventWatcher) processFile(fileName string) {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		watcher.logger.LogError("Watcher fatal: " + err.Error())
	}

	if fileInfo.Mode().IsRegular() {
		watcher.logger.LogInfo("Watcher is sending new work for: " + fileName)
		watcher.workChan <- fileInfo.Name()
	}
}
