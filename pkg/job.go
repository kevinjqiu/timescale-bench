package pkg

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"sync"
	"time"
)

type Result struct {
	JobID  string
	Result time.Duration
}

type ResultMap struct {
	*sync.Mutex
	m map[string]*Result
}

func (rm *ResultMap) Set(jobID string, result *Result) {
	rm.Lock()
	defer rm.Unlock()

	rm.m[jobID] = result
}

func (rm *ResultMap) IsDone() bool {
	rm.Lock()
	defer rm.Unlock()

	for _, result := range rm.m {
		if result == nil {
			return false
		}
	}
	return true
}

func newResultMap() ResultMap {
	return ResultMap{
		Mutex: new(sync.Mutex),
		m:     make(map[string]*Result),
	}
}

type Job struct {
	JobID      string
	QueryParam QueryParam
}

func (j Job) String() string {
	return fmt.Sprintf("<Job: id=%v, queryParam=%v>", j.JobID, j.QueryParam)
}

func newJob(queryParam QueryParam) Job {
	jobID := uuid.NewV4().String()
	return Job{
		JobID: jobID,
		QueryParam: queryParam,
	}
}

