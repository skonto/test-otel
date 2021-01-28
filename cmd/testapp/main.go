package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/skonto/test-otel/pkg/memstats"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/label"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/semconv"
	"google.golang.org/grpc"
)

func initMetrics() {
	ctx := context.Background()
	driver := otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint("0.0.0.0:55680"),
		otlpgrpc.WithDialOption(grpc.WithBlock()), // useful for testing
	)
	exp, err := otlp.NewExporter(ctx, driver)
	handleErr(err, "failed to create exporter")
	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("testapp"),
			label.Key("name").String("stavros"),
		),
	)
	handleErr(err, "failed to create resource")
	cont := controller.New(
		processor.New(
			simple.NewWithExactDistribution(),
			exp,
			processor.WithMemory(true),
		),
		controller.WithPusher(exp),
		controller.WithCollectPeriod(2*time.Second),
		controller.WithResource(res),
	)
	if err := cont.Start(context.Background()); err != nil {
		log.Fatal("could not start controller:", err)
	}
	promExporter, err := prometheus.NewExporter(prometheus.Config{}, cont)
	if err != nil {
		log.Fatal("could not initialize prometheus:", err)
	}
	http.HandleFunc("/", promExporter.ServeHTTP)
	go func() {
		log.Fatal(http.ListenAndServe(":17000", nil))
	}()

	otel.SetMeterProvider(cont.MeterProvider())
	fmt.Println("Prometheus server running on :17000")
	fmt.Println("Exporting OTLP to :55680")
}

func main() {
	fmt.Println("Starting local example")
	initMetrics()
	if err := memstats.Start(
		memstats.WithMinimumReadMemStatsInterval(time.Second),
		memstats.WithLabels([]label.KeyValue{label.Key("app_name").String("testapp")}),
	); err != nil {
		panic(err)
	}
	// TODO add proper shutdown
	select {}
}
func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}
