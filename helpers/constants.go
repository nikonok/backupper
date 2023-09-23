package helpers

import "time"

const (
	// App settings
	// size of channel for new work
	WORKER_CHAN_SIZE = 100
	// size of channel for results
	RESULT_CHAN_SIZE = 10
	// prefix of "delete" files
	DELETE_PREFIX = "delete_"
	// buffer size for syscall
	BUFFER_SIZE = 4096
	// default ticker duration for Timer watcher
	TICKER_DURATION = time.Second
	// amount of Delete workers
	DELETE_WORKERS_AMOUNT = 5
	// amount of Copy workers
	COPY_WORKERS_AMOUNT = 10

	// Default args
	// default path to log file
	DEFAULT_LOG_PATH = "./log.txt"
	// default path to hot folder
	DEFAULT_HOT_FOLDER_PATH = "./hot"
	// default path to backup folder
	DEFAULT_BACKUP_FOLDER_PATH = "./backup"
	// default watcher type
	DEFAULT_WATCHER_TYPE = SysCallWatcherName
	// default regex for log viewing
	DEFAULT_REGEX = ".*"
)
