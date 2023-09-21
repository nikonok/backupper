package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/fsnotify/fsnotify"
	"github.com/nikonok/backupper/watchers"
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

	TickerDuration time.Duration

	FileCollection map[string]FileInfo

	CopyWork chan string
}

func runInitialCheck(ctx context.Context, appCfg *AppConfig) {
	files, err := os.ReadDir(appCfg.HotFolderPath)
	if err != nil {
		panic(err)
	}

	needsBackup := func(srcPath, dstPath string) bool {
		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			return false
		}

		dstInfo, err := os.Stat(dstPath)
		if err != nil || os.IsNotExist(err) {
			// If the destination backup file doesn't exist, then we need a backup.
			return true
		}

		return srcInfo.ModTime().After(dstInfo.ModTime())
	}

	for _, file := range files {
		if file.Type().IsRegular() {
			fmt.Println("Initial processing " + file.Name())

			fileInfo, err := file.Info()
			if err != nil {
				panic(err)
			}

			srcPath := filepath.Join(appCfg.HotFolderPath, file.Name())
			dstPath := filepath.Join(appCfg.BackupFolderPath, file.Name()+".bak")

			if needsBackup(srcPath, dstPath) {
				appCfg.CopyWork <- fileInfo.Name()
			}
		}
	}
}

func copyFile(srcPath string, destPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	bufReader := bufio.NewReader(srcFile)
	bufWriter := bufio.NewWriter(destFile)

	_, err = bufReader.WriteTo(bufWriter)
	if err != nil {
		return err
	}

	err = bufWriter.Flush()
	if err != nil {
		return err
	}

	return nil
}

func runCopier(ctx context.Context, appCfg *AppConfig, wg *sync.WaitGroup) {
	if _, err := os.Stat(appCfg.BackupFolderPath); os.IsNotExist(err) {
		err := os.MkdirAll(appCfg.BackupFolderPath, 0755) // 0755 is the file permission
		if err != nil {
			panic(err)
		}
		fmt.Println("Back up dir created")
	} else {
		fmt.Println("Back up dir exists")
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for fileName := range appCfg.CopyWork {
			fmt.Println("Copier new work " + fileName)
			srcPath := appCfg.HotFolderPath + "/" + fileName
			destPath := appCfg.BackupFolderPath + "/" + fileName + ".bak"
			err := copyFile(srcPath, destPath)
			if err != nil {
				fmt.Println("Error while coping " + err.Error())
			}
		}
	}()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println("Caught signal: " + sig.String())
		cancel()
	}()

	appCfg := AppConfig{
		HotFolderPath:    "./hot",
		BackupFolderPath: "./backup",
		TickerDuration:   1 * time.Second,
		FileCollection:   make(map[string]FileInfo),
		CopyWork:         make(chan string, workChanSize),
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

	var wg sync.WaitGroup

	// runCopier(ctx, &appCfg, &wg)
	// runInitialCheck(ctx, &appCfg)

	// watcher := watchers.CreateSysCallWatcher(appCfg.HotFolderPath, appCfg.CopyWork)
	// watcher := watchers.CreateEventWatcher(appCfg.HotFolderPath, appCfg.CopyWork)
	watcher := watchers.CreateTimerWatcher(appCfg.HotFolderPath, appCfg.CopyWork, appCfg.TickerDuration)
	watchers.StartWatcher(watcher, ctx, &wg)

	wg.Wait()

	fmt.Println("End of program")
}
