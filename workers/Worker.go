package workers

import (
	"context"
	"sync"
)

type Worker interface {
	work(context.Context)
}

func StartWorker(worker Worker, ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		worker.work(ctx)
	}()
}
