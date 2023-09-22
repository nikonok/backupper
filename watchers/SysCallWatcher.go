package watchers

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	log "github.com/nikonok/backupper/logger"
)

const (
	BUFFER_SIZE = 4096
)

type SysCallWatcher struct {
	logger log.Logger

	hotFolderPath string
	workChan      chan string
	buf           []byte
	fd            int
}

func CreateSysCallWatcher(hotFolderPath string, workChan chan string, logger log.Logger) Watcher {
	return &SysCallWatcher{
		logger:        logger,
		hotFolderPath: hotFolderPath,
		workChan:      workChan,
		buf:           make([]byte, BUFFER_SIZE),
	}
}

func (watcher *SysCallWatcher) watch(ctx context.Context) {
	watcher.logger.LogDebug("Starting Watcher")

	var err error
	watcher.fd, err = syscall.InotifyInit1(syscall.IN_NONBLOCK)
	if err != nil {
		watcher.logger.LogError("Watcher fatal: " + err.Error())
	}
	defer syscall.Close(watcher.fd)

	wd, err := syscall.InotifyAddWatch(watcher.fd, watcher.hotFolderPath, syscall.IN_CREATE|syscall.IN_MODIFY)
	if err != nil {
		watcher.logger.LogError("Watcher fatal: " + err.Error())
	}

	defer syscall.InotifyRmWatch(watcher.fd, uint32(wd))

	watcher.runLoop(ctx)
}

func (watcher *SysCallWatcher) runLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			watcher.logger.LogDebug("Stopping Watcher")
			return
		default:
			n, err := syscall.Read(watcher.fd, watcher.buf[:])
			if err == syscall.EAGAIN {
				time.Sleep(100 * time.Millisecond)
				continue
			} else if err != nil {
				watcher.logger.LogError("Watcher fatal: " + err.Error())
			}

			watcher.precessEvent(uint32(n))
		}
	}
}

func (watcher *SysCallWatcher) precessEvent(n uint32) {
	var offset uint32
	for offset < n {
		raw := (*syscall.InotifyEvent)(unsafe.Pointer(&watcher.buf[offset]))
		mask := raw.Mask
		name := string(watcher.buf[offset+syscall.SizeofInotifyEvent : offset+syscall.SizeofInotifyEvent+raw.Len])

		if mask&syscall.IN_CREATE == syscall.IN_CREATE || mask&syscall.IN_MODIFY == syscall.IN_MODIFY {
			name = strings.TrimRight(name, "\x00")
			watcher.processFile(name)
		}
		offset += syscall.SizeofInotifyEvent + raw.Len
	}
}

func (watcher *SysCallWatcher) processFile(fileName string) {
	fullPath := filepath.Join(watcher.hotFolderPath, fileName)

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		watcher.logger.LogError("Watcher fatal: " + err.Error())
	}

	if fileInfo.Mode().IsRegular() {
		watcher.logger.LogInfo("Watcher is sending new work for: " + fileName)
		watcher.workChan <- fileInfo.Name()
	}
}
