package pkg

import (
	"time"
)

func median(sortedDurations []time.Duration) time.Duration {
	size := len(sortedDurations)
	mid := size / 2

	if size % 2 == 0 {
		return (sortedDurations[mid-1] + sortedDurations[mid]) / 2
	} else {
		return sortedDurations[mid]
	}
}
