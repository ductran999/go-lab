// Command demo runs the harness on a scripted incident: check the
// retry buckets, then answer. One tool is gated to ask (granted),
// one is denied (skipped by the loop). Goal → acts → obs → answer,
// all printed so the loop is visible.
package main

import (
	"context"
	"fmt"
	"os"

	"go-lab/tools/harness/internal/harness"
)

func main() {
	brain := &harness.MockBrain{Script: []harness.Action{
		{Tool: "retry_stats"},
		{Tool: "restart_service", Input: "retry"},
		{Tool: "metrics_query", Input: "http_requests_total"},
		{Done: true, Answer: "buckets spread, no restart needed, traffic nominal"},
	}}

	agent := harness.Agent{
		Brain: brain,
		Tools: map[string]harness.Tool{
			"retry_stats": func(_ context.Context, _ string) (string, error) {
				return `{"hits":20,"buckets":{"0":17,"1":3}}`, nil
			},
			"restart_service": func(_ context.Context, input string) (string, error) {
				return "restarted " + input, nil
			},
			"metrics_query": func(_ context.Context, input string) (string, error) {
				return input + ": 797", nil
			},
		},
		Gate: harness.Gate{
			Deny: []string{"restart_*"},
			Ask:  []string{"metrics_*"},
			AskFunc: func(tool string) bool {
				fmt.Printf("[ask] permit %s? → yes (scripted)\n", tool)

				return true
			},
		},
		MaxSteps: 10,
	}

	answer, err := agent.Run(context.Background(), "are the retries healthy?")
	if err != nil {
		fmt.Println("agent failed:", err)
		os.Exit(1)
	}

	fmt.Println("answer:", answer)
}
