# Tools — what each layer gives the agent

> Bare model chats. Layers make it work.

| Combo        | Gets       | Means                                                               |
| ------------ | ---------- | ------------------------------------------------------------------- |
| AI + Skill   | experience | workflows, patterns, formats — knows HOW (doc-table, go-strict)     |
| AI + MCP     | hands      | tools + live data beyond reach — can ACT (bench_run, metrics_query) |
| AI + Harness | guardrails | permission gates, step budgets, checkpoints — acts SAFELY           |

Skill answers "do it which way" (right way). Harness answers
"how far may it go" (allowed tools, ask-before-danger, stop
conditions). MCP answers "with what" (the tools themselves).
All three compose: skill shapes the plan, MCP tools execute it,
harness keeps the run inside bounds.

Lying underneath: `mcp/` (go-lab as agent tools),
`harness/` (the loop around any brain).
