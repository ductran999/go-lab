# SLOs & metrics — promise, then prove

> TL;DR: **SLI** measures, **SLO** promises in-house, **SLA** signs
> with customers (miss = pay). Instrument **RED** (services), **USE**
> (resources), **business** (money) — in that order of survival.

## 1. SLI → SLO → SLA

- **SLI** (Indicator): the number — uptime 99.9%, p99 < 200ms,
  time-to-ready p95 < 5min.
- **SLO** (Objective): internal target — keep SLI above X. No penalty,
  but pages fire and error budgets burn.
- **SLA** (Agreement): contract with teeth — miss SLO, pay credits.
  Needs SLOs underneath, or it's a wish.
- SLOs only where failure hurts (checkout, login, provisioning);
  measure everything, promise selectively.

## 2. Instrument: RED, USE, business

- **RED** (services — have it): Rate (`http_requests_total`), Errors
  (% 5xx), Duration (histogram p50/p99). One middleware `Observe`.
- **USE** (resources): Utilization (CPU/mem), Saturation (queue,
  pool waits), Errors (OOM). Exporter/node level.
- **Business** (most valuable): orders/min, login success %,
  time-to-ready. Same middleware counting at `POST /orders` —
  same Observe call, business labels.
- RED says the app lives; business says it deserves to.
