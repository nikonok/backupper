package watchers

import (
	"context"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
)

type EventWatcher struct {
	hotFolderPath string
	copyWork      chan string
	fsWatcher     *fsnotify.Watcher
}

func CreateEventWatcher(hotFolderPath string, copyWork chan string) Watcher {
	return &EventWatcher{
		hotFolderPath: hotFolderPath,
		copyWork:      copyWork,
	}
}

func (watcher *EventWatcher) watch(ctx context.Context) {
	var err error
	watcher.fsWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	defer watcher.fsWatcher.Close()

	err = watcher.fsWatcher.Add(watcher.hotFolderPath)
	if err != nil {
		panic(err)
	}

	watcher.runLoop(ctx)
}

func (watcher *EventWatcher) runLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Exiting watcher")
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
			fmt.Println("Error:" + err.Error())
		}
	}
}

func (watcher *EventWatcher) processFile(fileName string) {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		panic(err.Error())
	}

	if fileInfo.Mode().IsRegular() {
		fmt.Printf("Working with: %s\n", fileName)
		watcher.copyWork <- fileInfo.Name()
	}
}
