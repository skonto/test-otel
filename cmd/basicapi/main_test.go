package main

import (
	"testing"
)

func BenchmarkWithoutBinding(b *testing.B) {
	initSyncIntruments()
	cases := []struct {
		name string
	}{
		{"binding"},
		{"nobinding"},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				if c.name == "binding" {
					recordMetricsB(int64(100), float64(100))
				} else {
					recordMetrics(int64(100), float64(100))
				}
			}
		})
	}
}
