# Caching Performance Benchmark: Redis, Memcached, Dragonfly, and KeyDB

This research evaluates the performance of distributed caching systems when integrated with a Go-based backend. The focus is on latency, memory allocation, and the impact of payload size on system throughput.

## 💻 Test Environment

- **OS**: Windows 11 (amd64)
- **CPU**: Intel(R) Core(TM) i5-7400 CPU @ 3.00GHz (4 Cores, 4 Threads)
- **Go Version**: 1.24+ (using `b.Loop()`)
- **Drivers**: `go-redis/v9`, `gomemcache`, `redigo`

---

## 📊 Benchmark Results

### 1. Small Payload (~100 Bytes)

_Typical use case: API Key Metadata, User Sessions, Rate Limit Counters._

| Engine        | Latency (1 Core) | Latency (4 Cores) | Memory (B/op) | Allocs (op) |
| :------------ | :--------------- | :---------------- | :------------ | :---------- |
| **Memcached** | **0.92 ms**      | **0.90 ms**       | 272 B         | 7           |
| **Dragonfly** | 0.97 ms          | 0.92 ms           | 271 B         | 7           |
| **Redis**     | 1.00 ms          | 0.93 ms           | 271 B         | 7           |
| **KeyDB**     | 0.95 ms          | 1.11 ms           | 271 B         | 7           |

### 2. Large Payload (~100 KB)

_Typical use case: Large JSON Blobs, Compressed Metadata, Small Media Files._

| Engine        | Latency (1 Core) | Latency (4 Cores) | Memory (B/op) | Allocs (op) |
| :------------ | :--------------- | :---------------- | :------------ | :---------- |
| **Memcached** | **2.44 ms**      | **3.00 ms**       | **106,645 B** | 7           |
| **Redis**     | 13.89 ms         | 11.94 ms          | **244 B**     | 6           |
| **KeyDB**     | 15.39 ms         | 11.61 ms          | 244 B         | 6           |
| **Dragonfly** | 15.60 ms         | 15.12 ms          | 244 B         | 6           |

---

## 🔍 Key Engineering Insights

### 1. The "1ms Latency Floor"

Regardless of the cache engine, every distributed call on a local network incurs a **~1ms latency**.

- **The Bottleneck**: It is not the CPU or the Hash speed, but the **Network Overhead (TCP/IP stack)**.
- **Architectural Impact**: Distributed caching is **1000x slower** than in-memory `sync.Map` but **10x-50x faster** than raw Database queries.

### 2. Memcached: The Blob Specialist

For large 100KB payloads, Memcached is **5x-6x faster** than Redis.

- **Why?**: Memcached uses a simplified binary protocol and a slab allocation memory management system optimized for raw blobs. It bypasses the complex data-type logic found in Redis.

### 3. Redis: Zero-copy Memory Efficiency

A surprising result was the `B/op` (Bytes per operation):

- **Redis (244 B)**: Even when sending 100KB, the `go-redis` driver utilizes **Zero-copy** and **Buffer Pooling**. This results in near-zero pressure on the Go Garbage Collector (GC).
- **Memcached (106 KB)**: The driver allocates a fresh buffer for the entire payload, which can increase GC pauses under high-traffic scenarios.

### 4. Concurrency & Context Switching

- **Redis/KeyDB**: Showed improved performance as CPU cores increased (13ms -> 11ms), indicating efficient parallel socket I/O handling by the Go drivers.
- **Memcached**: Latency slightly _increased_ under higher concurrency (2.4ms -> 3.0ms). This often points to **Lock Contention** within the Windows network stack or the driver’s internal lock management when many goroutines compete for the same resource.

---

## 🛠️ How to Reproduce

To run these benchmarks on your own infrastructure:

```bash
# Run benchmarks with memory stats and multiple CPU counts
go test -bench=. -benchmem -cpu=1,2,4 -benchtime=10s
```
