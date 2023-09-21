package watchers

import (
	"context"
	"sync"
)

type Watcher interface {
	watch(context.Context)
}

func StartWatcher(watcher Watcher, ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		watcher.watch(ctx)
	}()
}
