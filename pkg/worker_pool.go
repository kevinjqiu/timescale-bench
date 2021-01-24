package pkg

import (
	"github.com/sirupsen/logrus"
	"time"
)

// WorkerPool represents a pool of worker goroutines
// It's responsible for routing a query to a specific worker
type WorkerPool struct {
	logger     *logrus.Entry
	numWorkers int
	workers    []*Worker
}

func (wp *WorkerPool) selectWorker(queryParam QueryParam) *Worker {
	hostnameHash := queryParam.GetHostnameHashInt(defaultHasherFactory())
	workerId := int(hostnameHash % uint64(len(wp.workers)))
	return wp.workers[workerId]
}

func (wp *WorkerPool) dispatch(queryParam QueryParam) {
	worker := wp.selectWorker(queryParam)
	worker.inputChan <- queryParam
	wp.logger.Infof("%s is dispatched to worker: %s", queryParam, worker)
}

func (wp *WorkerPool) startWorkers(resultsChan chan<- time.Duration) {
	wp.logger.Info("worker pool is running")
	errChan := make(chan error)

	for _, worker := range wp.workers {
		go worker.Run(resultsChan, errChan)
	}
}

func newWorkerPool(numWorkers int) (*WorkerPool, error) {
	workerPool := WorkerPool{
		logger:     logrus.WithField("component", "WorkerPool"),
		numWorkers: numWorkers,
		workers:    make([]*Worker, 0, numWorkers),
	}

	for i := 0; i < numWorkers; i++ {
		worker, err := newWorker(i)
		if err != nil {
			return nil, err
		}
		workerPool.workers = append(workerPool.workers, worker)
	}

	return &workerPool, nil
}
