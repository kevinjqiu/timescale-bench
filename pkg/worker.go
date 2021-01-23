package pkg

import (
	"crypto/md5"
	"fmt"
	"github.com/sirupsen/logrus"
)

var defaultHasherFactory = md5.New

// WorkerPool represents a pool of worker goroutines
// It's responsible for routing a query to a specific worker
type WorkerPool struct {
	numWorkers int
	workers    []*Worker
}

func (wp *WorkerPool) selectWorker(queryParam QueryParam) *Worker {
	hostnameHash := queryParam.GetHostnameHashInt(defaultHasherFactory())
	workerId := int(hostnameHash % uint64(len(wp.workers)))

	return wp.workers[workerId]
}

func (wp *WorkerPool) Dispatch(queryParam QueryParam) {
	worker := wp.selectWorker(queryParam)
	worker.inputChan <- queryParam
	logrus.Infof("%s is dispatched to worker: %s", queryParam, worker)
}

func (wp *WorkerPool) Run() {
	for _, worker := range wp.workers {
		go worker.Run()
	}

	select {}
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

func (w *Worker) String() string {
	return fmt.Sprintf("<Worker: %d>", w.id)
}

func (w *Worker) Run() {
	logger := logrus.WithField("worker", w.id)
	for {
		select {
		case queryParam := <-w.inputChan:
			logger.Infof("Got: %v", queryParam)
		case err := <-w.errChan:
			logger.Warnf("Encountered error: %v", err)
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
