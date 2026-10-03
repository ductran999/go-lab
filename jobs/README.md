# Jobs — background work you can trust

Scheduled and durable execution: cron fires it, workers survive it.

| Dir         | What                                                |
| ----------- | --------------------------------------------------- |
| `cron/`     | robfig/cron schedules, ctx timeouts, graceful drain |
| `temporal/` | reserved: durable workflows (retry/state/history)   |

Ties: `../observability/` (trace cron runs), `../messaging/`
(queues feeding workers).
