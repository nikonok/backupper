package helpers

import "github.com/nikonok/backupper/logger"

type ModeType int

const (
	Default ModeType = iota
	ViewLogs
)

type ViewConfig struct {
	Regex string
	Date  string
}

type AppConfig struct {
	HotFolderPath    string
	BackupFolderPath string
	LoggerFilePath   string

	Mode     ModeType
	ViewCfg  *ViewConfig
	LogLevel logger.LogLevel
	Watcher  WatcherType
}
