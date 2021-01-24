package pkg

import (
	"github.com/pkg/errors"
	"time"
)

func median(sortedDurations []time.Duration) (time.Duration, error) {
	if len(sortedDurations) == 0 {
		return 0, errors.New("sortedDurations cannot be empty")
	}

	size := len(sortedDurations)
	mid := size / 2

	if size % 2 == 0 {
		return (sortedDurations[mid-1] + sortedDurations[mid]) / 2, nil
	} else {
		return sortedDurations[mid], nil
	}
}
