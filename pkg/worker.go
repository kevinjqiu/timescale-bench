package pkg

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
	"time"
)

// withTiming is a decorator that returns the timing of the execution of f
func withTiming(f func() error) (time.Duration, error) {
	start := time.Now()
	if err := f(); err != nil {
		return 0, err
	}
	elapsed := time.Now()
	duration := elapsed.Sub(start)
	return duration, nil
}

var defaultHasherFactory = md5.New

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

func (wp *WorkerPool) Run(doneChan chan struct{}) {
	wp.logger.Info("worker pool is running")
	resultsChan := make(chan time.Duration)
	errChan := make(chan error)

	for _, worker := range wp.workers {
		go worker.Run(resultsChan, errChan)
	}

	durations := make([]time.Duration, 0, 0)

	for {
		select {
		case duration := <-resultsChan:
			wp.logger.Debugf("Got results: %v", duration)
			durations = append(durations, duration)
		case err := <-errChan:
			wp.logger.Warnf("Encountered error: %v", err)
		case <-doneChan:
			wp.logger.Info("Shutdown workers")
			for _, worker := range wp.workers {
				//w := *worker
				//w.terminateChan <- struct{}{}
				close(worker.inputChan)
				worker.conn.Close(context.Background())
			}
			wp.logger.Info("Total: ", len(durations))
			doneChan <- struct{}{}
		}
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

// Worker is responsible for receive the QueryParam, time the query execution and aggregate the metrics
type Worker struct {
	id            int
	conn          *pgx.Conn
	logger        *logrus.Entry
	inputChan     chan QueryParam
	terminateChan chan struct{}
}

func (w *Worker) String() string {
	return fmt.Sprintf("<Worker: %d>", w.id)
}

func (w *Worker) runQuery(queryParam QueryParam) (time.Duration, error) {
	sql := `SELECT
	time_bucket('1m', ts),
	max(usage),
	min(usage)
FROM cpu_usage
WHERE
	host = $1 AND (ts BETWEEN $2 AND $3)
GROUP BY 1;
`
	var (
		rows     pgx.Rows
		err      error
		duration time.Duration
	)

	duration, err = withTiming(func() error {
		rows, err = w.conn.Query(context.Background(), sql, queryParam.Hostname, queryParam.StartTime, queryParam.EndTime)
		return err
	})

	if err != nil {
		return 0, err
	}

	defer rows.Close()

	// TODO: is this needed?
	var (
		ts       time.Time
		maxUsage float64
		minUsage float64
	)

	for rows.Next() {
		if err := rows.Scan(&ts, &maxUsage, &minUsage); err != nil {
			return 0, err
		}
		w.logger.Debug(ts, maxUsage, minUsage)
	}

	return duration, nil
}

func (w *Worker) Run(resultsChan chan<- time.Duration, errChan chan<- error) {
	for {
		select {
		case queryParam := <-w.inputChan:
			w.logger.Debugf("Got: %v", queryParam)
			duration, err := w.runQuery(queryParam)
			if err != nil {
				errChan <- err
				break
			}
			resultsChan <- duration
		case <-w.terminateChan:
			w.conn.Close(context.TODO())
			w.logger.Info("Timescaledb connection closed")
			w.logger.Info("Termination signal received. Shutting down...")
			return
		}
	}
}

func newWorker(id int) (*Worker, error) {
	conn, err := pgx.Connect(context.TODO(), "postgres://postgres:password@localhost:5432/homework")
	if err != nil {
		return nil, err
	}

	return &Worker{
		id:            id,
		conn:          conn,
		logger:        logrus.WithField("component", fmt.Sprintf("worker-%d", id)),
		inputChan:     make(chan QueryParam),
		terminateChan: make(chan struct{}),
	}, nil
}
