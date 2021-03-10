package main

import (
	"testing"
)

func BenchmarkMetricsRecording(b *testing.B) {
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

		b.Run(c.name+"-parallel", func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if c.name == "binding" {
						recordMetricsB(int64(100), float64(100))
					} else {
						recordMetrics(int64(100), float64(100))
					}
				}
			})
		})
	}
}
