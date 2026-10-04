package harness

import (
	"context"
	"strings"
	"testing"
)

func TestLoopRunsThinkActObserve(t *testing.T) {
	t.Parallel()

	var calls []string

	agent := Agent{
		Brain: &MockBrain{Script: []Action{
			{Tool: "probe", Input: "x"},
			{Done: true, Answer: "healthy"},
		}},
		Tools: map[string]Tool{
			"probe": func(_ context.Context, input string) (string, error) {
				calls = append(calls, input)

				return "ok", nil
			},
		},
		MaxSteps: 5,
	}

	answer, err := agent.Run(context.Background(), "check it")
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if answer != "healthy" || len(calls) != 1 || calls[0] != "x" {
		t.Fatalf("answer=%q calls=%v", answer, calls)
	}
}

func TestGateDeniesAndLoopContinues(t *testing.T) {
	t.Parallel()

	var calls []string

	agent := Agent{
		Brain: &MockBrain{Script: []Action{
			{Tool: "rm_rf", Input: "/"},
			{Tool: "safe", Input: "y"},
			{Done: true, Answer: "done"},
		}},
		Tools: map[string]Tool{
			"rm_rf": func(_ context.Context, input string) (string, error) {
				calls = append(calls, input)

				return "deleted everything", nil
			},
			"safe": func(_ context.Context, input string) (string, error) {
				calls = append(calls, input)

				return "ok", nil
			},
		},
		Gate:     Gate{Deny: []string{"rm_*"}},
		MaxSteps: 5,
	}

	answer, err := agent.Run(context.Background(), "be careful")
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if answer != "done" || len(calls) != 1 || calls[0] != "y" {
		t.Fatalf("denied tool must never run: answer=%q calls=%v", answer, calls)
	}
}

func TestBudgetStopsRunaway(t *testing.T) {
	t.Parallel()

	agent := Agent{
		Brain: &MockBrain{Script: []Action{
			{Tool: "loop"}, {Tool: "loop"}, {Tool: "loop"}, {Tool: "loop"}, {Tool: "loop"},
		}},
		Tools: map[string]Tool{
			"loop": func(_ context.Context, _ string) (string, error) { return "again", nil },
		},
		MaxSteps: 3,
	}

	_, err := agent.Run(context.Background(), "never finish")
	if err == nil || !strings.Contains(err.Error(), "budget out") {
		t.Fatalf("err = %v, want budget-out", err)
	}
}
