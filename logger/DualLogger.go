package logger

import (
	"fmt"
	"log"
	"os"
)

const (
	LOG_CHAN_SIZE = 10
)

type DualLogger struct {
	logLevel LogLevel

	logChan  chan string

	stdLog  *log.Logger
	fileLog *log.Logger
}

func CreateDualLogger(filePath string, logLevel LogLevel) (Logger, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	logger := &DualLogger{
		logLevel: logLevel,
		logChan: make(chan string, LOG_CHAN_SIZE),
		stdLog:  log.New(os.Stdout, "", log.LstdFlags),
		fileLog: log.New(file, "", log.LstdFlags),
	}

	go logger.background()

	return logger, nil
}

func (logger *DualLogger) background() {
	for msg := range logger.logChan {
		logger.stdLog.Print(msg)
		logger.fileLog.Print(msg)
	}
}

// Warning: not thread safe!
func (logger *DualLogger) Close() error {
	close(logger.logChan)

	if file, ok := logger.fileLog.Writer().(*os.File); ok {
		return file.Close()
	}

	return nil
}

func makePrefix(msg string, logLevel LogLevel) string {
	return fmt.Sprintf("[%s] %s", levelNames[logLevel], msg)
}

func (logger *DualLogger) log(msg string, logLevel LogLevel) {
	if(logLevel < logger.logLevel) {
		return
	}

	logger.logChan <- makePrefix(msg, logLevel)
	
	if(logLevel == Error) {
		logger.Close()
		os.Exit(1)
	}
}

func (logger *DualLogger) LogDebug(msg string) {
	logger.log(msg, Debug)
}

func (logger *DualLogger) LogInfo(msg string) {
	logger.log(msg, Info)
}

func (logger *DualLogger) LogWarn(msg string) {
	logger.log(msg, Warn)
}

func (logger *DualLogger) LogError(msg string) {
	logger.log(msg, Error)
}
