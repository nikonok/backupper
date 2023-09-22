package initial

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	log "github.com/nikonok/backupper/logger"
)

const (
	DELETE_PREFIX = "delete_"
)

type InitialChecker struct {
	logger log.Logger

	hotFolderPath    string
	backupFolderPath string
	workChan         chan string
}

func CreateInitialChecker(hotFolderPath, backupFolderPath string, workChan chan string, logger log.Logger) *InitialChecker {
	return &InitialChecker{
		logger:     logger,
		hotFolderPath:    hotFolderPath,
		backupFolderPath: backupFolderPath,
		workChan:         workChan,
	}
}

func isBackupNeeded(srcPath, dstPath string) (bool, error) {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return false, err
	}

	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	return srcInfo.ModTime().After(dstInfo.ModTime()), nil
}

func isDeleteNeeded(filename string) bool {
	return strings.HasPrefix(filename, DELETE_PREFIX)
}

func (checker *InitialChecker) Check(ctx context.Context) {
	checker.logger.LogDebug("Starting Checker")

	files, err := os.ReadDir(checker.hotFolderPath)
	if err != nil {
		checker.logger.LogError("Checker fatal: " + err.Error())
	}

	for _, file := range files {
		if file.Type().IsRegular() {
			checker.processFile(file)
		}
	}

	checker.logger.LogDebug("Finished Checker")
}

func (checker *InitialChecker) processFile(file fs.DirEntry) {
	if isDeleteNeeded(file.Name()) {
		checker.logger.LogDebug("Checker found delete work")
		checker.workChan <- file.Name()
		return
	}

	srcPath := filepath.Join(checker.hotFolderPath, file.Name())
	dstPath := filepath.Join(checker.backupFolderPath, file.Name()+".bak")

	isNeeded, err := isBackupNeeded(srcPath, dstPath)
	if err != nil {
		checker.logger.LogError("Checker fatal: " + err.Error())
	}
	if isNeeded {
		checker.logger.LogDebug("Checker found copy work")
		checker.workChan <- file.Name()
	}
}
