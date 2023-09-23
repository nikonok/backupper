package controllers

import (
	"context"
	"strings"
	"time"

	"github.com/nikonok/backupper/helpers"
	log "github.com/nikonok/backupper/logger"
)

type operationType int

const (
	Copy operationType = iota
	Delete
	None
)

type fileOperation struct {
	currentOp operationType
	nextOp    operationType
}

type scheduledDeleteOperation struct {
	fullFilename string
	when         time.Time
}

type Controller struct {
	logger log.Logger

	workChan   chan string
	copyChan   chan string
	deleteChan chan string

	resultChan chan string

	fileOngoing         map[string]*fileOperation
	fileDeleteScheduled map[string]*scheduledDeleteOperation
}

func CreateController(workChan, copyChan, deleteChan chan string, logger log.Logger) *Controller {
	return &Controller{
		logger:     logger,
		workChan:   workChan,
		copyChan:   copyChan,
		deleteChan: deleteChan,

		resultChan: make(chan string, helpers.RESULT_CHAN_SIZE),

		fileOngoing:         make(map[string]*fileOperation),
		fileDeleteScheduled: make(map[string]*scheduledDeleteOperation),
	}
}

func (controller *Controller) GetResultChan() chan string {
	return controller.resultChan
}

func (controller *Controller) Run(ctx context.Context) {
	controller.logger.LogDebug("Starting Controller")
	for {
		select {
		case <-ctx.Done():
			controller.logger.LogDebug("Stopping Controller")
			return
		case work := <-controller.workChan:
			controller.logger.LogInfo("Controller received new work: " + work)
			controller.handleWork(work)
		case result := <-controller.resultChan:
			controller.logger.LogDebug("Controller received result: " + result)
			controller.handleResult(result)
			controller.triggerNextForFile(result)
		}
	}
}

func (controller *Controller) handleWork(work string) {
	var opType operationType
	if strings.HasPrefix(work, helpers.DELETE_PREFIX) {
		// in case of scheduled delete work, skip rest processing
		if doDeleteNow := controller.handleIfScheduledDelete(work); !doDeleteNow {
			return
		}

		work = strings.TrimPrefix(work, helpers.DELETE_PREFIX)
		opType = Delete
		controller.logger.LogDebug("Got delete work for: " + work)
	} else {
		opType = Copy
		controller.logger.LogDebug("Got copy work for: " + work)
	}

	if value, ok := controller.fileOngoing[work]; ok {
		if value.currentOp != Delete && (opType == Delete || value.nextOp == None) {
			value.nextOp = opType
		} else {
			controller.logger.LogDebug("Dropped work for: " + work)
		}
	} else {
		controller.fileOngoing[work] = &fileOperation{
			currentOp: opType,
			nextOp:    None,
		}
		controller.sendWork(work, opType)
	}
}

func (controller *Controller) handleIfScheduledDelete(work string) bool {
	dateTimeStr, filename, isParsed := helpers.ParseScheduledDelete(work)
	if !isParsed {
		return true
	}

	if len(filename) == 0 {
		controller.logger.LogDebug("After date and time parsing filename is empty. Considering as not scheduled delete")
		return true
	}

	if value, ok := controller.fileDeleteScheduled[filename]; ok {
		if value.when.After(time.Now()) {
			controller.logger.LogDebug("Already scheduled delete for " + filename)
			return false
		} else {
			controller.logger.LogDebug("Got scheduled delete for " + filename)
			return true
		}
	}

	deleteTime, err := time.Parse(time.RFC3339, dateTimeStr)
	if err != nil {
		controller.logger.LogDebug("Failed to parse '" + dateTimeStr + "'. Considering as not scheduled delete")
		return true
	}

	controller.logger.LogInfo("Parsed date and time for " + filename + " as " + deleteTime.String())

	controller.fileDeleteScheduled[filename] = &scheduledDeleteOperation{
		fullFilename: work,
		when:         deleteTime,
	}

	controller.scheduleDelete(work, deleteTime)

	return false
}

func (controller *Controller) scheduleDelete(work string, deleteTime time.Time) {
	go func() {
		duration := deleteTime.Sub(time.Now())
		if duration > 0 {
			controller.logger.LogDebug("Trying to sleep for " + duration.String() + ", work = " + work)
			time.Sleep(duration)
			controller.logger.LogDebug("Triggering delete after sleep for " + work)
		} else {
			controller.logger.LogDebug("Delete time is in the past. Triggering delete for " + work)
		}
		controller.workChan <- work
	}()
}

func (controller *Controller) sendWork(work string, opType operationType) {
	switch opType {
	case Copy:
		controller.copyChan <- work
	case Delete:
		controller.deleteChan <- work
	}
}

func (controller *Controller) triggerNextForFile(file string) {
	if value, ok := controller.fileOngoing[file]; ok {
		value.currentOp, value.nextOp = value.nextOp, None
		controller.sendWork(file, value.currentOp)
	}
}

func (controller *Controller) handleResult(result string) {
	if value, ok := controller.fileOngoing[result]; ok {
		if value.nextOp == None {
			delete(controller.fileOngoing, result)
		}
	} else {
		controller.logger.LogError("fatal in Controller: not found file operation info")
	}

	delete(controller.fileDeleteScheduled, result)
}
