package tasks

import "testing"

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		from     TaskStatus
		to       TaskStatus
		expected bool
	}{
		// Valid transitions from todo
		{StatusTodo, StatusInProgress, true},
		{StatusTodo, StatusBlocked, true},
		{StatusTodo, StatusComplete, true},
		{StatusTodo, StatusTodo, true},

		// Valid transitions from blocked
		{StatusBlocked, StatusTodo, true},
		{StatusBlocked, StatusInProgress, true},
		{StatusBlocked, StatusComplete, true},
		{StatusBlocked, StatusBlocked, true},

		// Valid transitions from in-progress
		{StatusInProgress, StatusBlocked, true},
		{StatusInProgress, StatusInReview, true},
		{StatusInProgress, StatusComplete, true},
		{StatusInProgress, StatusInProgress, true},

		// Valid transitions from in-review
		{StatusInReview, StatusInProgress, true},
		{StatusInReview, StatusComplete, true},
		{StatusInReview, StatusInReview, true},

		// Invalid transitions
		{StatusTodo, StatusInReview, false},
		{StatusBlocked, StatusInReview, false},
		{StatusInReview, StatusTodo, false},
		{StatusInReview, StatusBlocked, false},
		{StatusComplete, StatusTodo, false},
		{StatusComplete, StatusInProgress, false},
		{StatusComplete, StatusBlocked, false},
		{StatusComplete, StatusInReview, false},
	}

	for _, tt := range tests {
		name := string(tt.from) + " -> " + string(tt.to)
		t.Run(name, func(t *testing.T) {
			result := IsValidTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("IsValidTransition(%q, %q) = %v, want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestGetValidTransitions(t *testing.T) {
	transitions := GetValidTransitions(StatusTodo)
	if len(transitions) == 0 {
		t.Error("Expected non-empty transitions for StatusTodo")
	}

	transitions = GetValidTransitions(StatusComplete)
	if len(transitions) != 0 {
		t.Errorf("Expected no transitions from StatusComplete, got %d", len(transitions))
	}
}

func TestValidateStatus(t *testing.T) {
	validStatuses := []string{"todo", "blocked", "in-progress", "in-review", "complete"}
	for _, s := range validStatuses {
		if err := ValidateStatus(s); err != nil {
			t.Errorf("ValidateStatus(%q) returned error: %v", s, err)
		}
	}

	if err := ValidateStatus("invalid"); err == nil {
		t.Error("ValidateStatus(\"invalid\") should return error")
	}
}
