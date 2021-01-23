package pkg

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestParseQueryParams(t *testing.T) {
	queryParam, err := parseQueryParam("host_000008,2017-01-01 08:59:22,2017-01-01 09:59:22")
	require.NoError(t, err)

	require.Equal(t, "host_000008", queryParam.Hostname)
	require.Equal(t, "2017-01-01T08:59:22Z", queryParam.StartTime.Format(time.RFC3339))
	require.Equal(t, "2017-01-01T09:59:22Z", queryParam.EndTime.Format(time.RFC3339))
}