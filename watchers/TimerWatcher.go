package watchers

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"time"
)

type fileInfo struct {
	ModificationTime time.Time
}

type TimerWatcher struct {
	hotFolderPath  string
	copyWork       chan string
	tickerDuration time.Duration
	fileCollection map[string]fileInfo

	ticker *time.Ticker
}

func CreateTimerWatcher(hotFolderPath string, copyWork chan string, tickerDuration time.Duration) Watcher {
	return &TimerWatcher{
		hotFolderPath:  hotFolderPath,
		copyWork:       copyWork,
		tickerDuration: tickerDuration,
		fileCollection: make(map[string]fileInfo),
	}
}

func (watcher *TimerWatcher) watch(ctx context.Context) {
	watcher.ticker = time.NewTicker(watcher.tickerDuration)
	watcher.runLoop(ctx)
}

func (watcher *TimerWatcher) runLoop(ctx context.Context) {
	for range watcher.ticker.C {
		fmt.Println("Start reading dir")

		files, err := os.ReadDir(watcher.hotFolderPath)
		if err != nil {
			panic(err)
		}

		for _, file := range files {
			if file.Type().IsRegular() {
				watcher.processFile(file)
			}
		}

		fmt.Println("End reading dir")
	}
}

func (watcher *TimerWatcher) processFile(file fs.DirEntry) {
	fmt.Println("Processing " + file.Name())

	fInfo, err := file.Info()
	if err != nil {
		panic(err)
	}

	if savedFileInfo, ok := watcher.fileCollection[fInfo.Name()]; !ok {
		watcher.fileCollection[fInfo.Name()] = fileInfo{
			ModificationTime: fInfo.ModTime(),
		}

		watcher.copyWork <- fInfo.Name()
	} else if savedFileInfo.ModificationTime.Before(fInfo.ModTime()) {
		savedFileInfo.ModificationTime = fInfo.ModTime()
		watcher.fileCollection[fInfo.Name()] = savedFileInfo

		watcher.copyWork <- fInfo.Name()
	}
}
