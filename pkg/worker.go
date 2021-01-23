package pkg

import "github.com/sirupsen/logrus"

// WorkerPool represents a pool of worker goroutines
// It's responsible for routing a query to a specific worker
type WorkerPool struct {
	numWorkers int
	workers    []*Worker
}

func (wp *WorkerPool) Dispatch(queryParam QueryParam) {

}

func newWorkerPool(numWorkers int) *WorkerPool {
	workerPool := WorkerPool{
		numWorkers: numWorkers,
		workers:    make([]*Worker, 0, numWorkers),
	}

	for i := 0; i < numWorkers; i++ {
		workerPool.workers = append(workerPool.workers, newWorker(i))
	}

	return &workerPool
}

// Worker is responsible for receive the QueryParam, time the query execution and aggregate the metrics
type Worker struct {
	id            int
	inputChan     chan QueryParam
	errChan       chan error
	terminateChan chan struct{}
}

func (w *Worker) Run() {
	logger := logrus.WithField("worker", w.id)
	for {
		select {
		case queryParam := <-w.inputChan:
			logger.Info("Got: %v", queryParam)
		case err := <-w.errChan:
			logger.Warn("Encountered error: %v", err)
		case <-w.terminateChan:
			// TODO: display results
			return
		}
	}
}

func newWorker(id int) *Worker {
	return &Worker{
		id:            id,
		inputChan:     make(chan QueryParam),
		errChan:       make(chan error),
		terminateChan: make(chan struct{}),
	}
}
