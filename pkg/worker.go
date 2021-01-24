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

// Worker is responsible for receive the QueryParam, time the query execution and aggregate the metrics
type Worker struct {
	id            int
	conn          *pgx.Conn
	logger        *logrus.Entry
	jobChan       chan Job
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
	//var (
	//	ts       time.Time
	//	maxUsage float64
	//	minUsage float64
	//)
	//
	//for rows.Next() {
	//	if err := rows.Scan(&ts, &maxUsage, &minUsage); err != nil {
	//		return 0, err
	//	}
	//	w.logger.Debug(ts, maxUsage, minUsage)
	//}

	return duration, nil
}

func (w *Worker) Run(resultsChan chan<- QueryResult, errChan chan<- error) {
	w.logger.Infof("Running worker %v", w)
	for {
		select {
		case job := <-w.jobChan:
			w.logger.Debugf("Got: %v", job)
			duration, err := w.runQuery(job.QueryParam)
			if err != nil {
				w.logger.Warn("Error encountered: ", err)
				resultsChan <- QueryResult{
					JobID: job.JobID,
					Error: err,
				}
				break
			}
			w.logger.Debugf("Sent to results chan")
			resultsChan <- QueryResult{
				JobID: job.JobID,
				Result: duration,
			}
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
		jobChan:       make(chan Job),
		terminateChan: make(chan struct{}),
	}, nil
}
