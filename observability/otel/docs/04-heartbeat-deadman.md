# Heartbeat dead-man — detect dead nodes, page once

> TL;DR: heartbeat UPSERTs `nodes.last_seen`; expose
> `node_last_heartbeat_unix`; alert `max by (node)` age over
> 3×TTL. One consumer → Kafka event direct; many consumers →
> metrics (+ webhook to Kafka).

## 1. Store

| Rule                                   | Why                                                                     |
| -------------------------------------- | ----------------------------------------------------------------------- |
| UPSERT `nodes.last_seen` per heartbeat | 1 row/node — `GROUP BY` raw heartbeats scans history every poll         |
| writer = heartbeat receiver            | no extra hop; receipt timestamp from monitor clock, never payload clock |

## 2. Expose

| Path                           | When                                                |
| ------------------------------ | --------------------------------------------------- |
| postgres_exporter custom query | Postgres already exported — 5 lines YAML, zero code |
| 30-line poller (15-30s loop)   | other DBs, or no exporter yet                       |

## 3. Three replicas, one truth

| Approach                 | Verdict                                                        |
| ------------------------ | -------------------------------------------------------------- |
| `max by (node)` in query | ✅ default — alive if any replica saw it; dead replica ignored |
| shared Redis             | consistent but every heartbeat pays a remote write             |
| sticky/leader ingest     | single writer, adds election for the monitor itself            |

Same DB behind all replicas → identical numbers + free redundancy.

## 4. Metric or Kafka event

| Consumers          | Pick                                                                   |
| ------------------ | ---------------------------------------------------------------------- |
| pager only         | DB poll → classify → Kafka direct (exact, rich payload, no scrape lag) |
| page + graph + SLO | metrics, webhook to Kafka if the notifier lives on events              |
| undecided          | one poll loop, two sinks (publish event + `Set()` gauge)               |

## 5. Alert without noise

| Guard                                 | Why                                                          |
| ------------------------------------- | ------------------------------------------------------------ |
| `for: 5m` on the rule                 | one dropped packet is not death                              |
| join inventory (`and inventory == 1`) | decommissioned nodes leave ghost series                      |
| separate `up{monitor} == 0` alert     | all-nodes-dead at once means the watcher died, not the fleet |
