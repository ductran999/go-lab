// Package harness is a minimal agent loop: observe → think → act.
// A Brain decides the next action from history, the Registry runs
// tools, the Gate permits each call (allow/ask/deny — the same gate
// that made you reject my bash commands). Production swaps MockBrain
// for a model API call; the loop, registry, and gate stay identical.
package harness

import (
	"context"
	"fmt"
	"strings"
)

// Action is one step: call tool with input, or finish with an answer.
type Action struct {
	Tool   string
	Input  string
	Done   bool
	Answer string
}

// Brain picks the next action. MockBrain replays a script (offline
// demo); a real brain calls a model with history + tool schemas.
type Brain interface {
	Next(ctx context.Context, history []string) (Action, error)
}

// Tool runs one action by name.
type Tool func(ctx context.Context, input string) (string, error)

// Gate decides allow/ask/deny per tool. Asks go to AskFunc —
// in opencode that's the permission prompt; here the demo answers
// from a script so the run stays deterministic.
type Gate struct {
	Deny    []string
	Ask     []string
	AskFunc func(tool string) bool
}

func match(patterns []string, tool string) bool {
	for _, p := range patterns {
		if p == "*" || p == tool || strings.HasPrefix(tool, strings.TrimSuffix(p, "*")) {
			return true
		}
	}

	return false
}

// Permit reports allow (true) or deny (false). Ask patterns consult
// AskFunc; no AskFunc means ask defaults to deny (safe side).
func (g Gate) Permit(tool string) bool {
	if match(g.Deny, tool) {
		return false
	}

	if match(g.Ask, tool) {
		return g.AskFunc != nil && g.AskFunc(tool)
	}

	return true
}

// Agent ties brain, tools, gate, and a step budget together.
type Agent struct {
	Brain    Brain
	Tools    map[string]Tool
	Gate     Gate
	MaxSteps int
}

// Run loops until Done, budget out, or a tool/gate failure.
// Every step appends to history — the model's entire memory.
func (a Agent) Run(ctx context.Context, goal string) (string, error) {
	history := []string{"goal: " + goal}

	for step := 1; step <= a.MaxSteps; step++ {
		action, err := a.Brain.Next(ctx, history)
		if err != nil {
			return "", fmt.Errorf("step %d think: %w", step, err)
		}

		if action.Done {
			return action.Answer, nil
		}

		history = append(history, fmt.Sprintf("act: %s(%s)", action.Tool, action.Input))

		if !a.Gate.Permit(action.Tool) {
			history = append(history, "denied: "+action.Tool)

			continue
		}

		tool, ok := a.Tools[action.Tool]
		if !ok {
			history = append(history, "no such tool: "+action.Tool)

			continue
		}

		out, err := tool(ctx, action.Input)
		if err != nil {
			history = append(history, fmt.Sprintf("error: %s: %v", action.Tool, err))

			continue
		}

		history = append(history, fmt.Sprintf("obs: %s", out))
	}

	return "", fmt.Errorf("budget out after %d steps", a.MaxSteps)
}

// MockBrain replays scripted actions (each Next pops one). Offline
// stand-in for a model: same interface, zero network.
type MockBrain struct {
	Script []Action
}

// Next implements Brain.
func (m *MockBrain) Next(_ context.Context, _ []string) (Action, error) {
	if len(m.Script) == 0 {
		return Action{}, fmt.Errorf("script out of steps")
	}

	next := m.Script[0]
	m.Script = m.Script[1:]

	return next, nil
}
