package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric/controller/basic"
)

// Direct copy of https://github.com/open-telemetry/opentelemetry-go/blob/master/example/prometheus/main.go
// Just to play around
var (
	lemonsKey = label.Key("ex.com/lemons")
)

func initMeter() {
	exporter, err := prometheus.InstallNewPipeline(prometheus.Config{}, basic.WithCollectPeriod(1*time.Minute))
	if err != nil {
		log.Panicf("failed to initialize prometheus exporter %v", err)
	}
	http.HandleFunc("/metrics", exporter.ServeHTTP)
	go func() {
		_ = http.ListenAndServe(":9090", nil)
	}()

	fmt.Println("Prometheus server running on :9090")
}

func main() {
	initMeter()
	if err := runtime.Start(
		runtime.WithMinimumReadMemStatsInterval(time.Second),
	); err != nil {
		panic(err)
	}

	meter := otel.Meter("ex.com/basic")

	observerLock := new(sync.RWMutex)
	observerValueToReport := new(float64)
	observerLabelsToReport := new([]label.KeyValue)
	cb := func(_ context.Context, result metric.Float64ObserverResult) {
		(*observerLock).RLock()
		value := *observerValueToReport
		labels := *observerLabelsToReport
		(*observerLock).RUnlock()
		result.Observe(value, labels...)
	}
	_ = metric.Must(meter).NewFloat64ValueObserver("ex.com.one", cb,
		metric.WithDescription("A ValueObserver set to a value"),
	)

	valuerecorder := metric.Must(meter).NewFloat64ValueRecorder("ex.com.two")
	counter := metric.Must(meter).NewFloat64Counter("ex.com.three")

	commonLabels := []label.KeyValue{lemonsKey.Int(10), label.String("A", "1"), label.String("B", "2"), label.String("C", "3")}
	notSoCommonLabels := []label.KeyValue{lemonsKey.Int(13)}

	ctx := context.Background()

	(*observerLock).Lock()
	*observerValueToReport = 1.0
	*observerLabelsToReport = commonLabels
	(*observerLock).Unlock()
	meter.RecordBatch(
		ctx,
		commonLabels,
		valuerecorder.Measurement(2.0),
		counter.Measurement(12.0),
	)

	time.Sleep(5 * time.Second)

	(*observerLock).Lock()
	*observerValueToReport = 1.0
	*observerLabelsToReport = notSoCommonLabels
	(*observerLock).Unlock()
	meter.RecordBatch(
		ctx,
		notSoCommonLabels,
		valuerecorder.Measurement(2.0),
		counter.Measurement(22.0),
	)

	time.Sleep(5 * time.Second)

	(*observerLock).Lock()
	*observerValueToReport = 13.0
	*observerLabelsToReport = commonLabels
	(*observerLock).Unlock()
	meter.RecordBatch(
		ctx,
		commonLabels,
		valuerecorder.Measurement(12.0),
		counter.Measurement(13.0),
	)

	fmt.Printf("Example finished updating, please visit :9090\n")

	select {}
}
