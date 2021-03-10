package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/unit"
)

var (
	requests        metric.Int64Counter
	requestsB       metric.BoundInt64Counter
	requestLatency  metric.Float64ValueRecorder
	requestLatencyB metric.BoundFloat64ValueRecorder
)

func initMeter() {
	exporter, err := prometheus.InstallNewPipeline(prometheus.Config{
		// View API is not ready yet, so we have to set up global boundaries here
		// https://github.com/open-telemetry/opentelemetry-go/issues/689
		// https://github.com/open-telemetry/opentelemetry-go/issues/689#issuecomment-622137029
		DefaultHistogramBoundaries: []float64{1, 5, 10, 50, 100},
	}, basic.WithCollectPeriod(1*time.Second)) // How fast metrics exported from Prometheus
	if err != nil {
		log.Panicf("failed to initialize prometheus exporter %v", err)
	}
	http.HandleFunc("/metrics", exporter.ServeHTTP)
	go func() {
		_ = http.ListenAndServe(":9090", nil)
	}()

	fmt.Println("Prometheus server running on :9090")
}

func initSyncIntruments() {
	meter := otel.GetMeterProvider().Meter(
		"go.opentelemetry.io/contrib/instrumentation/runtime",
		metric.WithInstrumentationVersion("semver:"+""))

	var err error
	requests, err = meter.NewInt64Counter("request.count", metric.WithDescription("number of requests received"))

	if err != nil {
		fmt.Printf("failed to setup requestBytes instrument %s\n", err.Error())
		os.Exit(1)
	}

	otherLabels := []label.KeyValue{label.String("path", "/api/list/other"), label.String("host", "remote")}
	// Unbinding needs to be done manually to free mem
	requestsB = requests.Bind(otherLabels...)

	requestLatency, err = meter.NewFloat64ValueRecorder("request.latency", metric.WithDescription("request latencies"), metric.WithUnit(unit.Milliseconds))
	if err != nil {
		fmt.Printf("failed to setup requestLatency instrument %s\n", err.Error())
		os.Exit(1)
	}

	requestLatencyB = requestLatency.Bind(otherLabels...)
}

func main() {
	initMeter()
	initSyncIntruments()

	go func() {
		// TODO make it more realistic using a server
		for {
			count := int64(rand.Float64() * 100)
			latency := rand.Float64() * 10
			recordMetrics(count, latency)
			recordMetricsB(count, latency)
			time.Sleep(time.Millisecond * 100)
		}
	}()

	fmt.Printf("Example finished updating, please visit :9090\n")
	select {}
}

func recordMetrics(count int64, latency float64) {
	// Create labels dynamically to emulate real use
	labels := []label.KeyValue{label.String("path", "/api/list/foo"), label.String("host", "localhost")}
	requests.Add(context.TODO(), count, labels...)
	requestLatency.Record(context.TODO(), latency, labels...)
}

func recordMetricsB(count int64, latency float64) {
	requestsB.Add(context.TODO(), count)
	requestLatencyB.Record(context.TODO(), latency)
}
