package adapters_test

// Contract Test Scaffold for Epic 9 — Testing Strategy
//
// This file provides test scaffolding for future adapter contract tests.
// Each section below maps to a planned Epic 9 story and should be filled
// in as those stories are implemented.
//
// Usage: Implement each TODO section by creating a ProviderFactory for the
// adapter under test and calling adapters.RunContractTests(t, factory).
// See contract_test.go (TextFileProvider) for the reference pattern.
//
// Adapter contract tests validate that every TaskProvider implementation
// behaves consistently across the shared contract defined in contract.go.
//
// Scaffolded adapters:
//   - Calendar adapter (read-only provider for .ics task sources)
//   - Remote/API adapter (HTTP-based task synchronization)
//   - Composite adapter (multi-provider aggregation)

import "testing"

// TestCalendarAdapterContract validates a future calendar-based TaskProvider.
//
// TODO(epic-9): Implement when calendar adapter satisfies TaskProvider interface.
// The calendar adapter is expected to be read-only; contract tests should
// handle ErrReadOnly from SaveTask/DeleteTask/MarkComplete gracefully.
func TestCalendarAdapterContract(t *testing.T) {
	t.Skip("scaffold: calendar adapter not yet implemented (Epic 9)")
}

// TestRemoteAdapterContract validates a future HTTP/API-based TaskProvider.
//
// TODO(epic-9): Implement when remote adapter satisfies TaskProvider interface.
// Tests should use httptest.NewServer for isolation. Consider testing:
//   - Network timeout handling
//   - Retry behavior on transient failures
//   - Offline fallback via WAL integration
func TestRemoteAdapterContract(t *testing.T) {
	t.Skip("scaffold: remote adapter not yet implemented (Epic 9)")
}

// TestCompositeAdapterContract validates a future multi-provider TaskProvider.
//
// TODO(epic-9): Implement when composite adapter satisfies TaskProvider interface.
// The composite adapter aggregates tasks from multiple providers. Consider:
//   - Task ID uniqueness across providers
//   - Write routing (which provider receives SaveTask calls)
//   - Partial failure handling (one provider down, others healthy)
func TestCompositeAdapterContract(t *testing.T) {
	t.Skip("scaffold: composite adapter not yet implemented (Epic 9)")
}
