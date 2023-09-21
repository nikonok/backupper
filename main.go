package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	workChanSize = 10
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

func runWatcher(ctx context.Context, appCfg *AppConfig, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(appCfg.TickerDuration)
		for range ticker.C {
			fmt.Println("Start reading dir")

			files, err := os.ReadDir(appCfg.HotFolderPath)
			if err != nil {
				panic(err)
			}

			for _, file := range files {
				if file.Type().IsRegular() {
					fmt.Println("Processing " + file.Name())

					fileInfo, err := file.Info()
					if err != nil {
						panic(err)
					}

					if savedFileInfo, ok := appCfg.FileCollection[fileInfo.Name()]; !ok {
						appCfg.FileCollection[fileInfo.Name()] = FileInfo{
							FileName:         fileInfo.Name(),
							ModificationTime: fileInfo.ModTime(),
							IsProcessed:      false,
						}

						appCfg.CopyWork <- fileInfo.Name()
					} else if savedFileInfo.ModificationTime.Before(fileInfo.ModTime()) {
						savedFileInfo.IsProcessed = false
						savedFileInfo.ModificationTime = fileInfo.ModTime()
						appCfg.FileCollection[fileInfo.Name()] = savedFileInfo

						appCfg.CopyWork <- fileInfo.Name()
					}
				}
			}

			fmt.Println("End reading dir")
		}
	}()
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
			err := copyFile(srcPath, destPath);
			if err != nil {
				fmt.Println("Error while coping " + err.Error())
			}
		}
	}()
}

func main() {
	appCfg := AppConfig{
		HotFolderPath:    "./hot",
		BackupFolderPath: "./backup",
		TickerDuration:   10 * time.Second,
		FileCollection:   make(map[string]FileInfo),
		CopyWork:         make(chan string, workChanSize),
	}
	ctx := context.Background()

	var wg sync.WaitGroup

	runCopier(ctx, &appCfg, &wg)
	runWatcher(ctx, &appCfg, &wg)

	wg.Wait()

	fmt.Println("hello world")
}
