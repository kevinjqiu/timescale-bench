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

// QueryResult represents the result of a query
type QueryResult struct {
	JobID  string
	Result time.Duration
	Error  error
}

// QueryResultMap is a synchronized map of job ids and their results
type QueryResultMap struct {
	*sync.Mutex
	m map[string]*QueryResult
}

func (rm *QueryResultMap) Set(jobID string, result *QueryResult) {
	rm.Lock()
	defer rm.Unlock()

	rm.m[jobID] = result
}

func (rm *QueryResultMap) Aggregate() BenchmarkResult {
	var ar BenchmarkResult

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

	if len(durations) == 0 {
		return ar
	}

	sort.Sort(durations)

	ar.Min = durations[0]
	ar.Max = durations[len(durations)-1]
	ar.Average = time.Duration(int(ar.TotalProcessingTime) / ar.NumQueries)
	ar.Median = median(durations)

	return ar
}

func (ar BenchmarkResult) Human() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("Num Queries: %d\n", ar.NumQueries))
	buffer.WriteString(fmt.Sprintf("Num Errors: %d\n", ar.NumErrors))
	buffer.WriteString(fmt.Sprintf("Total Processing time: %v\n", ar.TotalProcessingTime))
	buffer.WriteString(fmt.Sprintf("Min time: %v\n", ar.Min))
	buffer.WriteString(fmt.Sprintf("Max time: %v\n", ar.Max))
	buffer.WriteString(fmt.Sprintf("Average time: %v\n", ar.Average))
	buffer.WriteString(fmt.Sprintf("Median time: %v\n", ar.Median))

	return buffer.String()
}

func newResultMap() QueryResultMap {
	return QueryResultMap{
		Mutex: new(sync.Mutex),
		m:     make(map[string]*QueryResult),
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

type BenchmarkResult struct {
	NumQueries          int
	NumErrors           int
	TotalProcessingTime time.Duration
	Min                 time.Duration
	Max                 time.Duration
	Average             time.Duration
	Median              time.Duration
}
