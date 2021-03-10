Sample output:

```
# HELP request_count number of requests received
# TYPE request_count counter
request_count{host="localhost",path="/api/list/foo"} 12782
request_count{host="remote",path="/api/list/other"} 12782
# HELP request_latency request latencies
# TYPE request_latency histogram
request_latency_bucket{host="localhost",path="/api/list/foo",le="1"} 27
request_latency_bucket{host="localhost",path="/api/list/foo",le="5"} 134
request_latency_bucket{host="localhost",path="/api/list/foo",le="10"} 271
request_latency_bucket{host="localhost",path="/api/list/foo",le="50"} 271
request_latency_bucket{host="localhost",path="/api/list/foo",le="100"} 271
request_latency_bucket{host="localhost",path="/api/list/foo",le="+Inf"} 271
request_latency_sum{host="localhost",path="/api/list/foo"} 1370.2792491321175
request_latency_count{host="localhost",path="/api/list/foo"} 271
request_latency_bucket{host="remote",path="/api/list/other",le="1"} 27
request_latency_bucket{host="remote",path="/api/list/other",le="5"} 134
request_latency_bucket{host="remote",path="/api/list/other",le="10"} 271
request_latency_bucket{host="remote",path="/api/list/other",le="50"} 271
request_latency_bucket{host="remote",path="/api/list/other",le="100"} 271
request_latency_bucket{host="remote",path="/api/list/other",le="+Inf"} 271
request_latency_sum{host="remote",path="/api/list/other"} 1370.2792491321175
request_latency_count{host="remote",path="/api/list/other"} 271
```


To run the benchmarks:
```
go test -run=^$ -test.v  -test.benchmem=true  -cpu=3000 -test.bench=^BenchmarkMetricsRecording$ ./cmd/basicapi/
```

Sample output for different parallelism values:

```
 $go test -run=^$ -test.v  -test.benchmem=true  -test.bench=^BenchmarkMetricsRecording$ ./cmd/basicapi/
goos: linux
goarch: amd64
pkg: github.com/skonto/test-otel/cmd/basicapi
BenchmarkMetricsRecording
BenchmarkMetricsRecording/binding
BenchmarkMetricsRecording/binding-12         	98302672	        14.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/binding-parallel
BenchmarkMetricsRecording/binding-parallel-12         	427250520	         3.13 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/nobinding
BenchmarkMetricsRecording/nobinding-12                	64745592	        15.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/nobinding-parallel
BenchmarkMetricsRecording/nobinding-parallel-12       	357612045	         3.71 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/skonto/test-otel/cmd/basicapi	5.774s

$go test -run=^$ -test.v  -test.benchmem=true  -cpu=3000 -test.bench=^BenchmarkMetricsRecording$ ./cmd/basicapi/
goos: linux
goarch: amd64
pkg: github.com/skonto/test-otel/cmd/basicapi
BenchmarkMetricsRecording
BenchmarkMetricsRecording/binding
BenchmarkMetricsRecording/binding-3000         	87173678	        12.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/binding-parallel
BenchmarkMetricsRecording/binding-parallel-3000         	400623286	         3.20 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/nobinding
BenchmarkMetricsRecording/nobinding-3000                	91726245	        13.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/nobinding-parallel
BenchmarkMetricsRecording/nobinding-parallel-3000       	335823018	         3.77 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/skonto/test-otel/cmd/basicapi	5.936s

$go test -run=^$ -test.v  -test.benchmem=true  -cpu=15000 -test.bench=^BenchmarkMetricsRecording$ ./cmd/basicapi/
goos: linux
goarch: amd64
pkg: github.com/skonto/test-otel/cmd/basicapi
BenchmarkMetricsRecording
BenchmarkMetricsRecording/binding
BenchmarkMetricsRecording/binding-15000         	104157100	        12.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/binding-parallel
BenchmarkMetricsRecording/binding-parallel-15000         	353368759	         3.40 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/nobinding
BenchmarkMetricsRecording/nobinding-15000                	52698126	        19.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/nobinding-parallel
BenchmarkMetricsRecording/nobinding-parallel-15000       	294592383	         4.05 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/skonto/test-otel/cmd/basicapi	10.598s

```