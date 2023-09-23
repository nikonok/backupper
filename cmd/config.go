package main

import (
	"flag"
	"path/filepath"
	"strconv"

	"github.com/nikonok/backupper/helpers"
	"github.com/nikonok/backupper/logger"
)

var (
	isView           = flag.Bool("view", false, "Pass to view log file")
	hotFolderPath    = flag.String("hot-path", helpers.DEFAULT_HOT_FOLDER_PATH, "Path to watched dir")
	backupFolderPath = flag.String("backup-path", helpers.DEFAULT_BACKUP_FOLDER_PATH, "Path to backup folder")
	logPath          = flag.String("log", helpers.DEFAULT_LOG_PATH, "Path to log file")
	logLevel         = flag.Int("log-level", logger.Debug, "Pass to set log level from 0 to "+strconv.Itoa(logger.Error))
	watcherType      = flag.String("watcher", helpers.DEFAULT_WATCHER_TYPE, "Pass to set watcher. Possible: "+
		helpers.SysCallWatcherName+", "+helpers.EventWatcherName+", "+helpers.TimerWatcherName)

	dateFilter = flag.String("view-date-filter", "", "Date to filter logs by (Format: YYYY/MM/DD, like showed in logs)")
	regexFilter = flag.String("view-regex", helpers.DEFAULT_REGEX, "Regex to filter logs")
)

func CreateConfig() *helpers.AppConfig {
	flag.Parse()

	mode := helpers.Default
	if *isView {
		mode = helpers.ViewLogs
	}

	hotFolder, err := filepath.Abs(*hotFolderPath)
	if err != nil {
		panic(err)
	}

	backupFolderPath, err := filepath.Abs(*backupFolderPath)
	if err != nil {
		panic(err)
	}

	appCfg := &helpers.AppConfig{
		HotFolderPath:    hotFolder,
		BackupFolderPath: backupFolderPath,
		LoggerFilePath:   *logPath,

		Mode:     mode,
		LogLevel: logger.LogLevel(*logLevel),
		Watcher:  helpers.WatcherNamesConversion[*watcherType],
	}

	if appCfg.Mode == helpers.ViewLogs {
		appCfg.ViewCfg = &helpers.ViewConfig{
			Regex: *regexFilter,
			Date: *dateFilter,
		}
	}

	return appCfg
}
