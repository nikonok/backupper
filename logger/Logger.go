package logger

type LogLevel int

const (
	Debug = iota
	Info
	Warn
	Error
)

var levelNames = map[LogLevel]string{
	Debug: "DEBUG",
	Info:  "INFO",
	Warn:  "WARN",
	Error: "ERROR",
}

type Logger interface {
	LogDebug(msg string)
	LogInfo(msg string)
	LogWarn(msg string)
	LogError(msg string)

	Close() error
}

