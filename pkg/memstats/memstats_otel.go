package memstats

import (
	"context"
	"runtime"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/unit"
)

// config contains optional settings for reporting runtime metrics.
type config struct {
	// MinimumReadMemStatsInterval sets the minimum interval
	// between calls to runtime.ReadMemStats().  Negative values
	// are ignored.
	MinimumReadMemStatsInterval time.Duration

	// MeterProvider sets the metric.MeterProvider.  If nil, the global
	// Provider will be used.
	MeterProvider metric.MeterProvider

	// Export extra metrics
	extraRuntimeMetrics bool

	// Labels to use eg. from a resource
	labels []label.KeyValue
}

// DefaultMinimumReadMemStatsInterval is the default minimum interval
// between calls to runtime.ReadMemStats().  Use the
// WithMinimumReadMemStatsInterval() option to modify this setting in
// Start().
const DefaultMinimumReadMemStatsInterval time.Duration = 15 * time.Second

// WithMinimumReadMemStatsInterval sets a minimum interval between calls to
// runtime.ReadMemStats(), which is a relatively expensive call to make
// frequently.  This setting is ignored when `d` is negative.
func WithMinimumReadMemStatsInterval(d time.Duration) Option {
	return minimumReadMemStatsIntervalOption(d)
}

// WithExtraRuntimeMetrics sets a flag that if set to true will allow extra metrics
// to be emitted eg. goroutine num.
func WithExtraRuntimeMetrics() Option {
	return extraRuntimeMetricsOption(true)
}

// WithExtraRuntimeMetrics sets a flag that if set to true will allow extra metrics
// to be emitted eg. goroutine num.
func WithLabels(labels []label.KeyValue) Option {
	return labelsOption(labels)
}

type minimumReadMemStatsIntervalOption time.Duration

type extraRuntimeMetricsOption bool

type labelsOption []label.KeyValue

// ApplyRuntime implements Option.
func (o minimumReadMemStatsIntervalOption) ApplyRuntime(c *config) {
	if o >= 0 {
		c.MinimumReadMemStatsInterval = time.Duration(o)
	}
}

func (o extraRuntimeMetricsOption) ApplyRuntime(c *config) {
	c.extraRuntimeMetrics = bool(o)
}

func (o labelsOption) ApplyRuntime(c *config) {
	c.labels = o
}

type memstatsOtel struct {
	config config

	meter metric.Meter
	// Alloc is bytes of allocated heap objects.
	//
	// This is the same as heapAlloc (see below).
	alloc metric.Int64UpDownSumObserver

	// totalAlloc is cumulative bytes allocated for heap objects.
	//
	// totalAlloc increases as heap objects are allocated, but
	// unlike Alloc and HeapAlloc, it does not decrease when
	// objects are freed.
	totalAlloc metric.Int64UpDownSumObserver

	// sys is the total bytes of memory obtained from the OS.
	//
	// sys is the sum of the XSys fields below. Sys measures the
	// virtual address space reserved by the Go runtime for the
	// heap, stacks, and other internal data structures. It's
	// likely that not all of the virtual address space is backed
	// by physical memory at any given moment, though in general
	// it all was at some point.
	sys metric.Int64UpDownSumObserver

	// lookups is the number of pointer lookups performed by the
	// runtime.
	//
	// This is primarily useful for debugging runtime internals.
	lookups metric.Int64UpDownSumObserver

	// mallocs is the cumulative count of heap objects allocated.
	// The number of live objects is Mallocs - Frees.
	mallocs metric.Int64UpDownSumObserver

	// frees is the cumulative count of heap objects freed.
	frees metric.Int64UpDownSumObserver

	// heapAlloc is bytes of allocated heap objects.
	//
	// "Allocated" heap objects include all reachable objects, as
	// well as unreachable objects that the garbage collector has
	// not yet freed. Specifically, heapAlloc increases as heap
	// objects are allocated and decreases as the heap is swept
	// and unreachable objects are freed. Sweeping occurs
	// incrementally between GC cycles, so these two processes
	// occur simultaneously, and as a result HeapAlloc tends to
	// change smoothly (in contrast with the sawtooth that is
	// typical of stop-the-world garbage collectors).
	heapAlloc metric.Int64UpDownSumObserver

	// heapSys is bytes of heap memory obtained from the OS.
	//
	// heapSys measures the amount of virtual address space
	// reserved for the heap. This includes virtual address space
	// that has been reserved but not yet used, which consumes no
	// physical memory, but tends to be small, as well as virtual
	// address space for which the physical memory has been
	// returned to the OS after it became unused (see HeapReleased
	// for a measure of the latter).
	//
	// heapSys estimates the largest size the heap has had.
	heapSys metric.Int64UpDownSumObserver

	// heapIdle is bytes in idle (unused) spans.
	//
	// Idle spans have no objects in them. These spans could be
	// (and may already have been) returned to the OS, or they can
	// be reused for heap allocations, or they can be reused as
	// stack memory.
	//
	// heapIdle minus heapReleased estimates the amount of memory
	// that could be returned to the OS, but is being retained by
	// the runtime so it can grow the heap without requesting more
	// memory from the OS. If this difference is significantly
	// larger than the heap size, it indicates there was a recent
	// transient spike in live heap size.
	heapIdle metric.Int64UpDownSumObserver

	// heapInuse is bytes in in-use spans.
	//
	// In-use spans have at least one object in them. These spans
	// can only be used for other objects of roughly the same
	// size.
	//
	// heapInuse minus heapAlloc estimates the amount of memory
	// that has been dedicated to particular size classes, but is
	// not currently being used. This is an upper bound on
	// fragmentation, but in general this memory can be reused
	// efficiently.
	heapInuse metric.Int64UpDownSumObserver

	// HeapReleased is bytes of physical memory returned to the OS.
	//
	// This counts heap memory from idle spans that was returned
	// to the OS and has not yet been reacquired for the heap.
	heapReleased metric.Int64UpDownSumObserver

	// HeapObjects is the number of allocated heap objects.
	//
	// Like HeapAlloc, this increases as objects are allocated and
	// decreases as the heap is swept and unreachable objects are
	// freed.
	heapObjects metric.Int64UpDownSumObserver

	// StackInuse is bytes in stack spans.
	//
	// In-use stack spans have at least one stack in them. These
	// spans can only be used for other stacks of the same size.
	//
	// There is no StackIdle because unused stack spans are
	// returned to the heap (and hence counted toward HeapIdle).
	stackInuse metric.Int64UpDownSumObserver

	// stackSys is bytes of stack memory obtained from the OS.
	//
	// stackSys is StackInuse, plus any memory obtained directly
	// from the OS for OS thread stacks (which should be minimal).
	stackSys metric.Int64UpDownSumObserver

	// mSpanInuse is bytes of allocated mspan structures.
	mSpanInuse metric.Int64UpDownSumObserver

	// mSpanSys is bytes of memory obtained from the OS for mspan
	// structures.
	mSpanSys metric.Int64UpDownSumObserver

	// mCacheInuse is bytes of allocated mcache structures.
	mCacheInuse metric.Int64UpDownSumObserver

	// mCacheSys is bytes of memory obtained from the OS for
	// mcache structures.
	mCacheSys metric.Int64UpDownSumObserver

	// buckHashSys is bytes of memory in profiling bucket hash tables.
	buckHashSys metric.Int64UpDownSumObserver

	// gCSys is bytes of memory in garbage collection metadata.
	gCSys metric.Int64UpDownSumObserver

	// otherSys is bytes of memory in miscellaneous off-heap
	// runtime allocations.
	otherSys metric.Int64UpDownSumObserver

	// nextGC is the target heap size of the next GC cycle.
	//
	// The garbage collector's goal is to keep HeapAlloc ≤ NextGC.
	// At the end of each GC cycle, the target for the next cycle
	// is computed based on the amount of reachable data and the
	// value of GOGC.
	nextGC metric.Int64UpDownSumObserver

	// lastGC is the time the last garbage collection finished, as
	// nanoseconds since 1970 (the UNIX epoch).
	lastGC metric.Int64UpDownSumObserver

	// pauseTotalNs is the cumulative nanoseconds in GC
	// stop-the-world pauses since the program started.
	//
	// During a stop-the-world pause, all goroutines are paused
	// and only the garbage collector can run.
	pauseTotalNs metric.Int64UpDownSumObserver

	// numGC is the number of completed GC cycles.
	numGC metric.Int64UpDownSumObserver

	// numForcedGC is the number of GC cycles that were forced by
	// the application calling the GC function.
	numForcedGC metric.Int64UpDownSumObserver

	// gCCPUFraction is the fraction of this program's available
	// CPU time used by the GC since the program started.
	//
	// gCCPUFraction is expressed as a number between 0 and 1,
	// where 0 means GC has consumed none of this program's CPU. A
	// program's available CPU time is defined as the integral of
	// GOMAXPROCS since the program started. That is, if
	// GOMAXPROCS is 2 and a program has been running for 10
	// seconds, its "available CPU" is 20 seconds. GCCPUFraction
	// does not include CPU time used for write barrier activity.
	//
	// This is the same as the fraction of CPU reported by
	// GODEBUG=gctrace=1.
	gCCPUFraction metric.Float64UpDownSumObserver
}

// Option supports configuring optional settings for runtime metrics.
type Option interface {
	// ApplyRuntime updates *config.
	ApplyRuntime(*config)
}

// newConfig computes a config from the supplied Options.
func newConfig(opts ...Option) config {
	c := config{
		MeterProvider:               otel.GetMeterProvider(),
		MinimumReadMemStatsInterval: DefaultMinimumReadMemStatsInterval,
	}
	for _, opt := range opts {
		opt.ApplyRuntime(&c)
	}
	return c
}

// Start initializes reporting of runtime metrics using the supplied config.
func Start(opts ...Option) error {
	c := newConfig(opts...)
	if c.MinimumReadMemStatsInterval < 0 {
		c.MinimumReadMemStatsInterval = DefaultMinimumReadMemStatsInterval
	}
	if c.MeterProvider == nil {
		c.MeterProvider = otel.GetMeterProvider()
	}
	r := memstatsOtel{
		meter: c.MeterProvider.Meter(
			"go.opentelemetry.io/contrib/instrumentation/runtime",
			metric.WithInstrumentationVersion("semver:"+""),
		),
		config: c,
	}
	return r.register()
}

func (r *memstatsOtel) register() error {
	if r.config.extraRuntimeMetrics {
		startTime := time.Now()
		if _, err := r.meter.NewInt64SumObserver(
			"runtime.uptime",
			func(_ context.Context, result metric.Int64ObserverResult) {
				result.Observe(time.Since(startTime).Milliseconds())
			},
			metric.WithUnit(unit.Milliseconds),
			metric.WithDescription("Milliseconds since application was initialized"),
		); err != nil {
			return err
		}

		if _, err := r.meter.NewInt64UpDownSumObserver(
			"runtime.go.goroutines",
			func(_ context.Context, result metric.Int64ObserverResult) {
				result.Observe(int64(runtime.NumGoroutine()))
			},
			metric.WithDescription("Number of goroutines that currently exist"),
		); err != nil {
			return err
		}

		if _, err := r.meter.NewInt64SumObserver(
			"runtime.go.cgo.calls",
			func(_ context.Context, result metric.Int64ObserverResult) {
				result.Observe(runtime.NumCgoCall())
			},
			metric.WithDescription("Number of cgo calls made by the current process"),
		); err != nil {
			return err
		}
	}
	if err := r.registerMemStats(); err != nil {
		return err
	}

	return nil
}

func (r *memstatsOtel) registerMemStats() error {
	var (
		err         error
		liveObjects metric.Int64UpDownSumObserver
		//gcPauseNs    metric.Int64ValueRecorder
		//lastNumGC    uint32
		lastMemStats time.Time
		memStats     runtime.MemStats

		// lock prevents a race between batch observer and instrument registration.
		lock sync.Mutex
	)

	lock.Lock()
	defer lock.Unlock()

	batchObserver := r.meter.NewBatchObserver(func(ctx context.Context, result metric.BatchObserverResult) {
		lock.Lock()
		defer lock.Unlock()
		now := time.Now()
		if now.Sub(lastMemStats) >= r.config.MinimumReadMemStatsInterval {
			runtime.ReadMemStats(&memStats)
			lastMemStats = now
		}
		var observations []metric.Observation
		observations = append(observations,
			r.alloc.Observation(int64(memStats.Alloc)),
			r.totalAlloc.Observation(int64(memStats.TotalAlloc)),
			r.sys.Observation(int64(memStats.Sys)),
			r.lookups.Observation(int64(memStats.Lookups)),
			r.mallocs.Observation(int64(memStats.Mallocs)),
			r.frees.Observation(int64(memStats.Frees)),
			r.heapAlloc.Observation(int64(memStats.HeapAlloc)),
			r.heapSys.Observation(int64(memStats.HeapSys)),
			r.heapIdle.Observation(int64(memStats.HeapIdle)),
			r.heapInuse.Observation(int64(memStats.HeapInuse)),
			r.heapReleased.Observation(int64(memStats.HeapReleased)),
			r.heapObjects.Observation(int64(memStats.HeapObjects)),
			r.stackInuse.Observation(int64(memStats.StackInuse)),
			r.stackSys.Observation(int64(memStats.StackSys)),
			r.mSpanInuse.Observation(int64(memStats.MSpanInuse)),
			r.mSpanSys.Observation(int64(memStats.MSpanSys)),
			r.mCacheInuse.Observation(int64(memStats.MCacheInuse)),
			r.mCacheSys.Observation(int64(memStats.MCacheSys)),
			r.buckHashSys.Observation(int64(memStats.BuckHashSys)),
			r.gCSys.Observation(int64(memStats.GCSys)),
			r.otherSys.Observation(int64(memStats.OtherSys)),
			r.nextGC.Observation(int64(memStats.NextGC)),
			r.lastGC.Observation(int64(memStats.LastGC)),
			r.pauseTotalNs.Observation(int64(memStats.PauseTotalNs)),
			r.numGC.Observation(int64(memStats.NumGC)),
			r.numForcedGC.Observation(int64(memStats.NumForcedGC)),
			r.gCCPUFraction.Observation(float64(memStats.GCCPUFraction)),
		)

		if r.config.extraRuntimeMetrics {
			observations = append(observations,
				liveObjects.Observation(int64(memStats.Mallocs-memStats.Frees)))
		}
		result.Observe(r.config.labels, observations...)

		// This causes a nil pointer exception after some time, not sure why although it is just a copy of
		// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/master/instrumentation/runtime/runtime.go
		// computeGCPauses(ctx, &gcPauseNs, memStats.PauseNs[:], lastNumGC, memStats.NumGC)
		// lastNumGC = memStats.NumGC
	})

	if r.alloc, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.alloc",
		metric.WithDescription("The number of bytes of allocated heap objects."),
	); err != nil {
		return err
	}

	if r.totalAlloc, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.total_alloc",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The cumulative bytes allocated for heap objects."),
	); err != nil {
		return err
	}

	if r.sys, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.sys",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The total bytes of memory obtained from the OS."),
	); err != nil {
		return err
	}

	if r.lookups, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.loookups",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of pointer lookups performed by the runtime."),
	); err != nil {
		return err
	}

	if r.mallocs, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.mallocs",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The cumulative count of heap objects allocated."),
	); err != nil {
		return err
	}

	if r.frees, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.frees",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The cumulative count of heap objects freed."),
	); err != nil {
		return err
	}

	if r.heapAlloc, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.heap_alloc",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of bytes of allocated heap objects."),
	); err != nil {
		return err
	}

	if r.heapSys, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.heap_sys",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of bytes of heap memory obtained from the OS."),
	); err != nil {
		return err
	}

	if r.heapIdle, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.heap_idle",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of bytes in idle (unused) spans."),
	); err != nil {
		return err
	}

	if r.heapInuse, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.heap_inuse",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of bytes in in-use spans."),
	); err != nil {
		return err
	}

	if r.heapReleased, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.heap_released",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of bytes of physical memory returned to the OS."),
	); err != nil {
		return err
	}

	if r.heapObjects, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.heap_objects",
		metric.WithDescription("Number of allocated heap objects"),
	); err != nil {
		return err
	}

	if r.stackInuse, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.stack_in_use",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of bytes in stack spans."),
	); err != nil {
		return err
	}

	if r.stackSys, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.stack_sys",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of bytes of stack memory obtained from the OS."),
	); err != nil {
		return err
	}

	if r.mSpanInuse, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.mspan_in_use",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of bytes of allocated mspan structures."),
	); err != nil {
		return err
	}

	if r.mSpanSys, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.mspan_sys",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of bytes of memory obtained from the OS for mspan structures."),
	); err != nil {
		return err
	}

	if r.mCacheInuse, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.mcache_in_use",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of bytes of allocated mcache structures."),
	); err != nil {
		return err
	}

	if r.mCacheSys, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.mcache_sys",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of bytes of memory obtained from the OS for mcache structures."),
	); err != nil {
		return err
	}

	if r.buckHashSys, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.bucket_hash_sys",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of bytes of memory in profiling bucket hash tables."),
	); err != nil {
		return err
	}

	if r.gCSys, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.bucket_hash_sys",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of bytes of memory in garbage collection metadata."),
	); err != nil {
		return err
	}

	if r.otherSys, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.other_sys",
		metric.WithUnit(unit.Bytes),
		metric.WithDescription("The number of bytes of memory in miscellaneous off-heap runtime allocations."),
	); err != nil {
		return err
	}

	if r.nextGC, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.next_gc",
		metric.WithDescription("The target heap size of the next GC cycle."),
	); err != nil {
		return err
	}

	if r.lastGC, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.last_gc",
		metric.WithDescription("The time the last garbage collection finished, as nanoseconds since 1970 (the UNIX epoch)."),
	); err != nil {
		return err
	}

	if r.pauseTotalNs, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.total_gc_pause_ns",
		metric.WithDescription("The cumulative nanoseconds in GC stop-the-world pauses since the program started."),
	); err != nil {
		return err
	}

	if r.numGC, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.num_gc",
		metric.WithDescription("The number of completed GC cycles."),
	); err != nil {
		return err
	}

	if r.numForcedGC, err = batchObserver.NewInt64UpDownSumObserver(
		"runtime.go.mem.num_forced_gc",
		metric.WithDescription("The number of GC cycles that were forced by the application calling the GC function."),
	); err != nil {
		return err
	}

	if r.gCCPUFraction, err = batchObserver.NewFloat64UpDownSumObserver(
		"runtime.go.mem.num_gc_cpu_fraction",
		metric.WithDescription("The fraction of this program's available CPU time used by the GC since the program started."),
	); err != nil {
		return err
	}

	if r.config.extraRuntimeMetrics {
		if liveObjects, err = batchObserver.NewInt64UpDownSumObserver(
			"runtime.go.mem.live_objects",
			metric.WithDescription("Number of live objects is the number of cumulative Mallocs - Frees"),
		); err != nil {
			return err
		}

		//if gcPauseNs, err = r.meter.NewInt64ValueRecorder(
		//	"runtime.go.gc.pause_ns",
		//	// TODO: nanoseconds units
		//	metric.WithDescription("Amount of nanoseconds in GC stop-the-world pauses"),
		//); err != nil {
		//	return err
		//}
	}
	return nil
}

//func computeGCPauses(
//	ctx context.Context,
//	recorder *metric.Int64ValueRecorder,
//	circular []uint64,
//	lastNumGC, currentNumGC uint32,
//) {
//	delta := int(int64(currentNumGC) - int64(lastNumGC))
//
//	if delta == 0 {
//		return
//	}
//
//	if delta >= len(circular) {
//		// There were > 256 collections, some may have been lost.
//		recordGCPauses(ctx, recorder, circular)
//		return
//	}
//
//	length := uint32(len(circular))
//
//	i := lastNumGC % length
//	j := currentNumGC % length
//
//	if j < i { // wrap around the circular buffer
//		recordGCPauses(ctx, recorder, circular[i:])
//		recordGCPauses(ctx, recorder, circular[:j])
//		return
//	}
//
//	recordGCPauses(ctx, recorder, circular[i:j])
//}
//
//func recordGCPauses(
//	ctx context.Context,
//	recorder *metric.Int64ValueRecorder,
//	pauses []uint64,
//) {
//	for _, pause := range pauses {
//		fmt.Println("PAUSE:", pause)
//		recorder.Record(ctx, int64(pause))
//	}
//}
