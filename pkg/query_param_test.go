package pkg

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestParseQueryParams(t *testing.T) {
	testCases := []struct {
		name                string
		line                string
		expectError         bool
		expectedErrorString string
	}{
		{
			name:        "normal case",
			line:        "host_000008,2017-01-01 08:59:22,2017-01-01 09:59:22",
			expectError: false,
		},
		{
			name:                "not enough columns",
			line:                "host_000008,2017-01-01 08:59:22",
			expectError:         true,
			expectedErrorString: "wrong query_param format",
		},
		{
			name:                "too many columns",
			line:                "host_000008,2017-01-01 08:59:22,1,2,3",
			expectError:         true,
			expectedErrorString: "wrong query_param format",
		},
		{
			name:                "empty date",
			line:                "host_000008,2017-01-01 08:59:22,",
			expectError:         true,
			expectedErrorString: "wrong time format",
		},
		{
			name:                "wrong date format",
			line:                "host_000008,2017-01-01 08:59:22,2017-01-01T08:59:22Z",
			expectError:         true,
			expectedErrorString: "wrong time format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			queryParam, err := parseQueryParam(tc.line)
			if !tc.expectError {
				require.NoError(t, err)

				require.Equal(t, "host_000008", queryParam.Hostname)
				require.Equal(t, "2017-01-01T08:59:22Z", queryParam.StartTime.Format(time.RFC3339))
				require.Equal(t, "2017-01-01T09:59:22Z", queryParam.EndTime.Format(time.RFC3339))
			} else {
				require.True(t, -1 != strings.Index(err.Error(), tc.expectedErrorString))
			}
		})
	}
}
