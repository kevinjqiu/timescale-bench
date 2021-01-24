package pkg

import (
	"github.com/sirupsen/logrus"
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

func (wp *WorkerPool) Dispatch(job Job) {
	worker := wp.selectWorker(job.QueryParam)
	worker.jobChan <- job
	wp.logger.Debugf("%s is dispatched to worker: %s", job, worker)
}

func (wp *WorkerPool) StartWorkers(resultsChan chan<- QueryResult) {
	wp.logger.Info("Start the worker pool")
	for _, worker := range wp.workers {
		go worker.Run(resultsChan)
	}
}

func (wp *WorkerPool) shutdown() {
	for _, worker := range wp.workers {
		worker.terminateChan <- struct{}{}
	}
}

func newWorkerPool(numWorkers int, dbURL string) (*WorkerPool, error) {
	workerPool := WorkerPool{
		logger:     logrus.WithField("component", "WorkerPool"),
		numWorkers: numWorkers,
		workers:    make([]*Worker, 0, numWorkers),
	}

	for i := 0; i < numWorkers; i++ {
		worker, err := newWorker(i, dbURL)
		if err != nil {
			return nil, err
		}
		workerPool.workers = append(workerPool.workers, worker)
	}

	return &workerPool, nil
}
