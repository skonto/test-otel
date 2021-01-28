The idea is to briefly demonstrate the following architecture covering two use cases

a) export metrics from the app side for Prometheus to scrape them directly

b) push metrics from the app to the collector and let the latter to export them

![architecture](arch.png)

Start with the collector running as a central local service:
```bash
docker run --rm -p 13133:13133 -p 14250:14250 -p 14268:14268 \
-p 55678-55680:55678-55680 -p 6060:6060 -p 7276:7276 -p 8888:8888 -p 8889:8889 \
-p 9411:9411 -p 9943:9943 \
-v "$(pwd)/local.yaml":/otel-local-config.yaml \
--name otelcol otel/opentelemetry-collector \
--config /otel-local-config.yaml otel/opentelemetry-collector-dev:latest```
```

Start the app:

```bash
go run ./cmd/testapp/main.go 
Starting local example
Prometheus server running on :17000
Exporting OTLP to :55680```

You should get output as follows at the collector stdout:

```

```bash
Descriptor:
     -> Name: stavros.runtime.go.mem.alloc
     -> Description: The number of bytes of allocated heap objects.
     -> Unit: 
     -> DataType: IntSum
     -> IsMonotonic: false
     -> AggregationTemporality: AGGREGATION_TEMPORALITY_CUMULATIVE
IntDataPoints #0
StartTime: 1611783783333527191
Timestamp: 1611783841333723881
Value: 1838400

```

Metrics are also exported at the 0.0.0.0:8889 port (from the collector container). 
Same metrics are exposed locally from the app at port http://localhost:17000/
Here is a dump of the metrics on 8889:
```bash
# HELP runtime_go_mem_alloc The number of bytes of allocated heap objects.
# TYPE runtime_go_mem_alloc counter
runtime_go_mem_alloc{app_name="testapp"} 1.419088e+06
# HELP runtime_go_mem_bucket_hash_sys The number of bytes of memory in profiling bucket hash tables.
# TYPE runtime_go_mem_bucket_hash_sys counter
runtime_go_mem_bucket_hash_sys{app_name="testapp"} 3.436808e+06
# HELP runtime_go_mem_frees The cumulative count of heap objects freed.
# TYPE runtime_go_mem_frees counter
runtime_go_mem_frees{app_name="testapp"} 238
# HELP runtime_go_mem_heap_alloc The number of bytes of allocated heap objects.
# TYPE runtime_go_mem_heap_alloc counter
runtime_go_mem_heap_alloc{app_name="testapp"} 1.419088e+06
# HELP runtime_go_mem_heap_idle The number of bytes in idle (unused) spans.
# TYPE runtime_go_mem_heap_idle counter
runtime_go_mem_heap_idle{app_name="testapp"} 6.356992e+07
# HELP runtime_go_mem_heap_inuse The number of bytes in in-use spans.
# TYPE runtime_go_mem_heap_inuse counter
runtime_go_mem_heap_inuse{app_name="testapp"} 2.62144e+06
# HELP runtime_go_mem_heap_objects Number of allocated heap objects
# TYPE runtime_go_mem_heap_objects counter
runtime_go_mem_heap_objects{app_name="testapp"} 6316
# HELP runtime_go_mem_heap_released The number of bytes of physical memory returned to the OS.
# TYPE runtime_go_mem_heap_released counter
runtime_go_mem_heap_released{app_name="testapp"} 6.3504384e+07
# HELP runtime_go_mem_heap_sys The number of bytes of heap memory obtained from the OS.
# TYPE runtime_go_mem_heap_sys counter
runtime_go_mem_heap_sys{app_name="testapp"} 6.619136e+07
# HELP runtime_go_mem_last_gc The time the last garbage collection finished, as nanoseconds since 1970 (the UNIX epoch).
# TYPE runtime_go_mem_last_gc counter
runtime_go_mem_last_gc{app_name="testapp"} 0
# HELP runtime_go_mem_loookups The number of pointer lookups performed by the runtime.
# TYPE runtime_go_mem_loookups counter
runtime_go_mem_loookups{app_name="testapp"} 0
# HELP runtime_go_mem_mallocs The cumulative count of heap objects allocated.
# TYPE runtime_go_mem_mallocs counter
runtime_go_mem_mallocs{app_name="testapp"} 6554
# HELP runtime_go_mem_mcache_in_use The number of bytes of allocated mcache structures.
# TYPE runtime_go_mem_mcache_in_use counter
runtime_go_mem_mcache_in_use{app_name="testapp"} 20832
# HELP runtime_go_mem_mcache_sys The number of bytes of memory obtained from the OS for mcache structures.
# TYPE runtime_go_mem_mcache_sys counter
runtime_go_mem_mcache_sys{app_name="testapp"} 32768
# HELP runtime_go_mem_mspan_in_use The number of bytes of allocated mspan structures.
# TYPE runtime_go_mem_mspan_in_use counter
runtime_go_mem_mspan_in_use{app_name="testapp"} 46512
# HELP runtime_go_mem_mspan_sys The number of bytes of memory obtained from the OS for mspan structures.
# TYPE runtime_go_mem_mspan_sys counter
runtime_go_mem_mspan_sys{app_name="testapp"} 49152
# HELP runtime_go_mem_next_gc The target heap size of the next GC cycle.
# TYPE runtime_go_mem_next_gc counter
runtime_go_mem_next_gc{app_name="testapp"} 4.473924e+06
# HELP runtime_go_mem_num_forced_gc The number of GC cycles that were forced by the application calling the GC function.
# TYPE runtime_go_mem_num_forced_gc counter
runtime_go_mem_num_forced_gc{app_name="testapp"} 0
# HELP runtime_go_mem_num_gc The number of completed GC cycles.
# TYPE runtime_go_mem_num_gc counter
runtime_go_mem_num_gc{app_name="testapp"} 0
# HELP runtime_go_mem_num_gc_cpu_fraction The fraction of this program's available CPU time used by the GC since the program started.
# TYPE runtime_go_mem_num_gc_cpu_fraction counter
runtime_go_mem_num_gc_cpu_fraction{app_name="testapp"} 0
# HELP runtime_go_mem_other_sys The number of bytes of memory in miscellaneous off-heap runtime allocations.
# TYPE runtime_go_mem_other_sys counter
runtime_go_mem_other_sys{app_name="testapp"} 755301
# HELP runtime_go_mem_stack_in_use The number of bytes in stack spans.
# TYPE runtime_go_mem_stack_in_use counter
runtime_go_mem_stack_in_use{app_name="testapp"} 917504
# HELP runtime_go_mem_stack_sys The number of bytes of stack memory obtained from the OS.
# TYPE runtime_go_mem_stack_sys counter
runtime_go_mem_stack_sys{app_name="testapp"} 917504
# HELP runtime_go_mem_sys The total bytes of memory obtained from the OS.
# TYPE runtime_go_mem_sys counter
runtime_go_mem_sys{app_name="testapp"} 7.1387144e+07
# HELP runtime_go_mem_total_alloc The cumulative bytes allocated for heap objects.
# TYPE runtime_go_mem_total_alloc counter
runtime_go_mem_total_alloc{app_name="testapp"} 1.419088e+06
# HELP runtime_go_mem_total_gc_pause_ns The cumulative nanoseconds in GC stop-the-world pauses since the program started.
# TYPE runtime_go_mem_total_gc_pause_ns counter
runtime_go_mem_total_gc_pause_ns{app_name="testapp"} 0
```