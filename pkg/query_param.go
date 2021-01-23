package pkg

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"strings"
	"time"
)

const QueryParamTimeLayout = "2006-01-02 15:04:05"

type QueryParam struct {
	Hostname  string
	StartTime time.Time
	EndTime   time.Time
}

func parseQueryParam(line string) (QueryParam, error) {
	var (
		queryParam QueryParam
		err        error
	)

	parts := strings.Split(line, ",")
	if len(parts) != 3 {
		return queryParam, fmt.Errorf("wrong query_param format: %v", line)
	}

	queryParam.Hostname = parts[0]
	queryParam.StartTime, err = time.Parse(QueryParamTimeLayout, parts[1])
	if err != nil {
		return queryParam, errors.Wrapf(err, "wrong time format: %s", line)
	}

	queryParam.EndTime, err = time.Parse(QueryParamTimeLayout, parts[2])
	if err != nil {
		return queryParam, errors.Wrapf(err, "wrong time format: %s", line)
	}

	return queryParam, nil
}

func processQueryParams(inputFile *os.File, chanQueryParam chan<- QueryParam, errChan chan<- error) {
	scanner := bufio.NewScanner(inputFile)
	scanner.Split(bufio.ScanLines)

	scanner.Scan() // skip the header
	for scanner.Scan() {
		line := scanner.Text()
		queryParam, err := parseQueryParam(line)
		if err != nil {
			errChan <- err
		} else {
			chanQueryParam <- queryParam
		}
	}
}
