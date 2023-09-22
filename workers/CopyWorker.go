package workers

import (
	"bufio"
	"context"
	"os"
	"path/filepath"

	log "github.com/nikonok/backupper/logger"
)

type CopyWorker struct {
	logger log.Logger

	hotFolderPath    string
	backupFolderPath string

	workChan   chan string
	resultChan chan string
}

func CreateCopyWorker(hotFolderPath, backupFolderPath string, workChan, resultChan chan string, logger log.Logger) Worker {
	return &CopyWorker{
		logger: logger,

		hotFolderPath:    hotFolderPath,
		backupFolderPath: backupFolderPath,

		workChan:   workChan,
		resultChan: resultChan,
	}
}

func (worker *CopyWorker) work(ctx context.Context) {
	worker.logger.LogDebug("Starting CopyWorker")
	for {
		select {
		case <-ctx.Done():
			worker.logger.LogDebug("Stopping CopyWorker")
			return
		case work := <-worker.workChan:
			worker.logger.LogInfo("CopyWorker received new work: " + work)
			worker.handleWork(work)
			worker.resultChan <- work
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

func (worker *CopyWorker) handleWork(work string) {
	srcPath := filepath.Join(worker.hotFolderPath, work)
	destPath := filepath.Join(worker.backupFolderPath, work+".bak")

	err := copyFile(srcPath, destPath)
	if err != nil {
		worker.logger.LogWarn("CopyWorker error: " + err.Error())
	}
}
