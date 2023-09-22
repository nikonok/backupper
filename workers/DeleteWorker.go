package workers

import (
	"context"
	"os"
	"path/filepath"

	log "github.com/nikonok/backupper/logger"
)

const (
	DELETE_PREFIX = "delete_"
)

type DeleteWorker struct {
	logger log.Logger

	hotFolderPath    string
	backupFolderPath string

	workChan   chan string
	resultChan chan string
}

func CreateDeleteWorker(hotFolderPath, backupFolderPath string, workChan, resultChan chan string, logger log.Logger) Worker {
	return &DeleteWorker{
		logger: logger,

		hotFolderPath:    hotFolderPath,
		backupFolderPath: backupFolderPath,

		workChan:   workChan,
		resultChan: resultChan,
	}
}

func (worker *DeleteWorker) work(ctx context.Context) {
	worker.logger.LogDebug("Starting DeleteWorker")
	for {
		select {
		case <-ctx.Done():
			worker.logger.LogDebug("Stopping DeleteWorker")
			return
		case work := <-worker.workChan:
			worker.logger.LogInfo("DeleteWorker received new work: " + work)
			worker.handleWork(work)
			worker.resultChan <- work
		}
	}
}

func removeFile(path string) error {
	return os.Remove(path)
}

func (worker *DeleteWorker) handleWork(work string) {
	originalFile := filepath.Join(worker.hotFolderPath, work)
	backupFile := filepath.Join(worker.backupFolderPath, work+".bak")
	deleteFile := filepath.Join(worker.hotFolderPath, DELETE_PREFIX+work)

	if err := removeFile(originalFile); err != nil {
		worker.logger.LogWarn("DeleteWorker error on original file: " + err.Error())
	}

	if err := removeFile(backupFile); err != nil {
		worker.logger.LogWarn("DeleteWorker error on backup file: " + err.Error())
	}

	if err := removeFile(deleteFile); err != nil {
		worker.logger.LogWarn("DeleteWorker error on delete_ file: " + err.Error())
	}
}