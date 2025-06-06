package monitoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveRuntimeStats(t *testing.T) {
	m := New(5)
	m.PollSignal = make(chan any, 1)
	m.saveRuntimeStats()

	stats := m.GetRuntimeStats()

	require.NotEmpty(t, stats)
	assert.Contains(t, stats, "Alloc")
}

func TestSaveGopsStats(t *testing.T) {
	m := New(5)
	m.saveGopsStats()

	require.NotNil(t, m.gopsStats)

	_, ok := m.gopsStats["CPUutilization1"]
	assert.True(t, ok)
	_, ok = m.gopsStats["FreeMemory"]
	assert.True(t, ok)
	_, ok = m.gopsStats["TotalMemory"]
	assert.True(t, ok)
}
