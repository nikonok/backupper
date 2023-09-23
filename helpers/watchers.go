package helpers

type WatcherType int

const (
	SysCallWatcherType = iota
	EventWatcherType
	TimerWatcherType
)

const (
	SysCallWatcherName = "syscall"
	EventWatcherName   = "event"
	TimerWatcherName   = "timer"
)

var WatcherNamesConversion = map[string]WatcherType{
	SysCallWatcherName: SysCallWatcherType,
	EventWatcherName:   EventWatcherType,
	TimerWatcherName:   TimerWatcherType,
}
