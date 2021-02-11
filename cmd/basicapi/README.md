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