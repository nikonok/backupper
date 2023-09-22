package controllers

import (
	"context"
	"strings"

	log "github.com/nikonok/backupper/logger"
)

const (
	DELETE_PREFIX = "delete_"
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

type Controller struct {
	logger log.Logger

	workChan   chan string
	copyChan   chan string
	deleteChan chan string

	resultChan chan string

	fileOngoing map[string]*fileOperation
}

func CreateController(workChan, copyChan, deleteChan chan string, logger log.Logger) *Controller {
	return &Controller{
		logger:     logger,
		workChan:   workChan,
		copyChan:   copyChan,
		deleteChan: deleteChan,

		resultChan: make(chan string),

		fileOngoing: make(map[string]*fileOperation),
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
	if strings.HasPrefix(work, DELETE_PREFIX) {
		work = strings.TrimPrefix(work, DELETE_PREFIX)
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
}
