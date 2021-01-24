package pkg

import (
	"bytes"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"sort"
	"sync"
	"time"
)

// DurationsList makes []time.Duration sortable
// because go has no generics...
type DurationsList []time.Duration

func (d DurationsList) Len() int {
	return len(d)
}

func (d DurationsList) Less(i, j int) bool {
	return d[i] < d[j]
}

func (d DurationsList) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Result represents the result of a query
type Result struct {
	JobID  string
	Result time.Duration
	Error  error
}

// ResultMap is a synchronized map of job ids and their results
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

func (rm *ResultMap) Aggregate() AggregatedResult {
	var ar AggregatedResult

	durations := make(DurationsList, 0, 0)

	for _, result := range rm.m {
		if result.Error != nil {
			ar.NumErrors += 1
			continue
		}

		ar.NumQueries += 1
		ar.TotalProcessingTime += result.Result

		durations = append(durations, result.Result)
	}

	sort.Sort(durations)

	ar.Min = durations[0]
	ar.Max = durations[len(durations)-1]
	ar.Average = time.Duration(int(ar.TotalProcessingTime) / ar.NumQueries)

	if len(durations) % 2 == 0 {
		mid := len(durations) / 2
		ar.Median = (durations[mid] + durations[mid+1]) / 2
	} else {
		mid := len(durations) / 2
		ar.Median = durations[mid]
	}

	return ar
}

func (ar AggregatedResult) Human() string {
	var buffer bytes.Buffer

	buffer.WriteString("Results:\n")
	buffer.WriteString(fmt.Sprintf("Num Queries: %d\n", ar.NumQueries))
	buffer.WriteString(fmt.Sprintf("Num Errors: %d\n", ar.NumErrors))
	buffer.WriteString(fmt.Sprintf("Total Processing time: %v\n", ar.TotalProcessingTime))
	buffer.WriteString(fmt.Sprintf("Min time: %v\n", ar.Min))
	buffer.WriteString(fmt.Sprintf("Max time: %v\n", ar.Max))
	buffer.WriteString(fmt.Sprintf("Average time: %v\n", ar.Average))
	buffer.WriteString(fmt.Sprintf("Median time: %v\n", ar.Median))

	return buffer.String()
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
		JobID:      jobID,
		QueryParam: queryParam,
	}
}

type AggregatedResult struct {
	NumQueries          int
	NumErrors           int
	TotalProcessingTime time.Duration
	Min                 time.Duration
	Max                 time.Duration
	Average             time.Duration
	Median              time.Duration
}
