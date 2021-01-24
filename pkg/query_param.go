package pkg

import (
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"hash"
	"strings"
	"time"
)

const QueryParamTimeLayout = "2006-01-02 15:04:05"

// QueryParam represents a row in the input file containing the query parameters
type QueryParam struct {
	Hostname  string
	StartTime time.Time
	EndTime   time.Time
}

func (qp QueryParam) GetHostnameHashInt(hasher hash.Hash) uint64 {
	hasher.Write([]byte(qp.Hostname))
	hashBytes := hasher.Sum(nil)
	hashInt := binary.BigEndian.Uint64(hashBytes)
	return hashInt
}

func (qp QueryParam) String() string {
	return fmt.Sprintf("<QueryParam: host=%s, start=%s, end=%s>", qp.Hostname, qp.StartTime, qp.EndTime)
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
