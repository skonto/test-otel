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
$ go test -run=^$ -test.v  -test.benchmem=true -test.bench=^BenchmarkMetricsRecording$ ./cmd/basicapi/
goos: linux
goarch: amd64
pkg: github.com/skonto/test-otel/cmd/basicapi
BenchmarkMetricsRecording
BenchmarkMetricsRecording/binding
BenchmarkMetricsRecording/binding-12         	100000000	        12.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/binding-parallel
BenchmarkMetricsRecording/binding-parallel-12         	439574839	         3.16 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/nobinding
BenchmarkMetricsRecording/nobinding-12                	14812455	        80.8 ns/op	     128 B/op	       1 allocs/op
BenchmarkMetricsRecording/nobinding-parallel
BenchmarkMetricsRecording/nobinding-parallel-12       	36191892	        40.7 ns/op	     128 B/op	       1 allocs/op
PASS
ok  	github.com/skonto/test-otel/cmd/basicapi	6.672s
$ go test -run=^$ -test.v  -test.benchmem=true  -cpu=3000 -test.bench=^BenchmarkMetricsRecording$ ./cmd/basicapi/
goos: linux
goarch: amd64
pkg: github.com/skonto/test-otel/cmd/basicapi
BenchmarkMetricsRecording
BenchmarkMetricsRecording/binding
BenchmarkMetricsRecording/binding-3000         	100000000	        14.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/binding-parallel
BenchmarkMetricsRecording/binding-parallel-3000         	406842705	         3.27 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/nobinding
BenchmarkMetricsRecording/nobinding-3000                	 9660532	       111 ns/op	     128 B/op	       1 allocs/op
BenchmarkMetricsRecording/nobinding-parallel
BenchmarkMetricsRecording/nobinding-parallel-3000       	18009536	        72.4 ns/op	     128 B/op	       1 allocs/op
PASS
ok  	github.com/skonto/test-otel/cmd/basicapi	5.982s
$ go test -run=^$ -test.v  -test.benchmem=true  -cpu=15000 -test.bench=^BenchmarkMetricsRecording$ ./cmd/basicapi/
goos: linux
goarch: amd64
pkg: github.com/skonto/test-otel/cmd/basicapi
BenchmarkMetricsRecording
BenchmarkMetricsRecording/binding
BenchmarkMetricsRecording/binding-15000         	90593160	        13.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/binding-parallel
BenchmarkMetricsRecording/binding-parallel-15000         	355757488	         3.28 ns/op	       0 B/op	       0 allocs/op
BenchmarkMetricsRecording/nobinding
BenchmarkMetricsRecording/nobinding-15000                	 8751476	       126 ns/op	     128 B/op	       1 allocs/op
BenchmarkMetricsRecording/nobinding-parallel
BenchmarkMetricsRecording/nobinding-parallel-15000       	21148315	       102 ns/op	     128 B/op	       1 allocs/op
PASS
ok  	github.com/skonto/test-otel/cmd/basicapi	9.570s

```