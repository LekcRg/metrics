package home

import (
	"strconv"
	"testing"

	"github.com/LekcRg/metrics/internal/server/storage"
)

// Before
// BenchmarkGenerateHTML-8   1648   723813 ns/op   1927271 B/op   10022 allocs/op

// After
// BenchmarkGenerateHTML-8   3722   326983 ns/op   1098840 B/op   4006 allocs/op

func BenchmarkGenerateHTML(b *testing.B) {
	const lenList = 1000
	var (
		gaugeVal   = storage.Gauge(1234560.789)
		counterVal = storage.Counter(1234560789)
	)

	gaugeList := make(storage.GaugeCollection, lenList)
	counterList := make(storage.CounterCollection, lenList)
	for i := range lenList {
		gaugeList["gauge-"+strconv.Itoa(i)] = gaugeVal
		counterList["counter-"+strconv.Itoa(i)] = counterVal
	}

	b.ResetTimer()
	for range b.N {
		generateHTML(storage.Database{
			Gauge:   gaugeList,
			Counter: counterList,
		})
	}

	b.ReportAllocs()
}
