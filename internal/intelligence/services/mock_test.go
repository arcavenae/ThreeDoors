package services

import "context"

// mockBackend implements llm.LLMBackend for testing.
type mockBackend struct {
	name      string
	responses []string // responses returned in order; cycles if exhausted
	callCount int
	err       error
}

func (m *mockBackend) Name() string { return m.name }

func (m *mockBackend) Complete(_ context.Context, _ string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	idx := m.callCount
	if idx >= len(m.responses) {
		idx = len(m.responses) - 1
	}
	m.callCount++
	return m.responses[idx], nil
}

func (m *mockBackend) Available(_ context.Context) bool { return true }
