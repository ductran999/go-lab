# Metric types — what each one is for, when, and what it tells you

> TL;DR: **Counter** counts events (Rate/Errors). **Gauge** reads the
> current level (Utilization/Saturation). **Histogram** maps latency
> into buckets (Duration/SLO, aggregates). **Summary** pre-computes
> quantiles on one instance (precise, never aggregates).

| Type      | Reads as                      | Use for                                                                                              | What it tells                                                                                                                                                          | How to use                                                                                        |
| --------- | ----------------------------- | ---------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------- |
| Counter   | what happened                 | counting events: requests, errors, orders, retries (RED Rate/Errors)                                 | throughput and error ratio over time; rate drop = traffic lost, error-ratio climb = budget burning                                                                     | `rate(http_requests_total[5m])`; ratio of rates for %; never read the raw value                   |
| Gauge     | the state right now           | current level, up or down: queue depth, pool in-use, goroutines, memory (USE Utilization/Saturation) | how close a resource is to full; saturation climbing on flat rate = each request costs more (leak, slow downstream)                                                    | `avg_over_time` / `max_over_time`; never `rate()` a gauge                                         |
| Histogram | how long it took, everywhere  | latency/size distributions in buckets (RED Duration, SLOs)                                           | p50/p99 across all replicas and SLO compliance; p99 climbing on flat p50 = sick tail (GC, noisy neighbor); one bucket sample carries the trace that explains the spike | `histogram_quantile(0.99, …)`; pick bucket boundaries from the SLO; `sum()` across replicas first |
| Summary   | how long it took, on this box | tail latency where only one instance exists, or buckets can't be picked upfront                      | precise local p99 — and nothing global (3 replicas' p99s never combine)                                                                                                | read the quantile directly, no function needed; don't use past 1 replica                          |

## Choosing — ask in order

| #   | Question                                           | Type      | Query                                                        |
| --- | -------------------------------------------------- | --------- | ------------------------------------------------------------ |
| 1   | How many / how fast? (events)                      | Counter   | `rate()` / `increase()`; ratio of rates for %                |
| 2   | At what level right now? (something always exists) | Gauge     | raw value, `avg_over_time`, `max_over_time`                  |
| 3   | What share meets the bar? (SLO, many replicas)     | Histogram | `histogram_quantile()`; `bucket{le=SLO} / bucket{le="+Inf"}` |
| 4   | Exact quantile here, buckets unknown, one box?     | Summary   | read `quantile="…"` directly                                 |

Stop at the first "yes": an event count is never a gauge, a
per-request value is never a gauge, a multi-replica quantile is
never a summary.

## In this lab

| Piece                    | Mapping                                                                                                                                                                                   |
| ------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Observe()` per request  | counter `http_requests_total{route,status}` + histogram `http_request_duration_seconds{route,status}` + summary `http_response_size_bytes{route}` (bytes counted in the logging recorder) |
| `Track()` around request | gauge `http_inflight_requests{route}` — Inc on entry, `defer done()` Dec on exit; stuck above 0 after load = leak                                                                         |
| Exemplars                | `trace_id` on both; OpenMetrics only (`EnableOpenMetrics: true`) — plain text hides them                                                                                                  |
| `./bench.sh N C`         | counter delta = N, `le="+Inf"` = N, one `# {trace_id="…"}` per series → graph spike jumps to Jaeger                                                                                       |
