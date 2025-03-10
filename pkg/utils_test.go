package pkg

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMedian(t *testing.T) {
	t.Run("single element slice", func(t *testing.T) {
		m := median([]time.Duration{1 * time.Second})
		require.Equal(t, 1*time.Second, m)
	})

	t.Run("double element slice", func(t *testing.T) {
		m := median([]time.Duration{1 * time.Second, 2 * time.Second})
		require.Equal(t, 1500*time.Millisecond, m)
	})

	t.Run("slice with odd number of elements", func(t *testing.T) {
		m := median([]time.Duration{10 * time.Second, 20 * time.Second, 30 * time.Second, 4 * time.Hour, 5 * time.Hour})
		require.Equal(t, 30*time.Second, m)
	})

	t.Run("slice with even number of elements", func(t *testing.T) {
		m := median([]time.Duration{20 * time.Second, 30 * time.Second, 4 * time.Hour, 5 * time.Hour})
		require.Equal(t, 2*time.Hour+15*time.Second, m)
	})
}
