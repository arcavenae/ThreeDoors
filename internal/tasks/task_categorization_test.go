package tasks

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestValidateTaskType(t *testing.T) {
	tests := []struct {
		name    string
		value   TaskType
		wantErr bool
	}{
		{"empty is valid", "", false},
		{"creative", TypeCreative, false},
		{"administrative", TypeAdministrative, false},
		{"technical", TypeTechnical, false},
		{"physical", TypePhysical, false},
		{"invalid", TaskType("banana"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTaskType(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTaskType(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTaskEffort(t *testing.T) {
	tests := []struct {
		name    string
		value   TaskEffort
		wantErr bool
	}{
		{"empty is valid", "", false},
		{"quick-win", EffortQuickWin, false},
		{"medium", EffortMedium, false},
		{"deep-work", EffortDeepWork, false},
		{"invalid", TaskEffort("huge"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTaskEffort(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTaskEffort(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTaskLocation(t *testing.T) {
	tests := []struct {
		name    string
		value   TaskLocation
		wantErr bool
	}{
		{"empty is valid", "", false},
		{"home", LocationHome, false},
		{"work", LocationWork, false},
		{"errands", LocationErrands, false},
		{"anywhere", LocationAnywhere, false},
		{"invalid", TaskLocation("mars"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTaskLocation(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTaskLocation(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestTask_Validate_WithCategories(t *testing.T) {
	task := NewTask("Test task")
	task.Type = TypeTechnical
	task.Effort = EffortMedium
	task.Location = LocationWork
	if err := task.Validate(); err != nil {
		t.Errorf("Validate() with valid categories returned error: %v", err)
	}

	task2 := NewTask("Test task")
	task2.Type = TaskType("invalid")
	if err := task2.Validate(); err == nil {
		t.Error("Validate() with invalid type should return error")
	}
}

func TestTask_YAML_RoundTrip_WithCategories(t *testing.T) {
	original := &TasksFile{
		Tasks: []*Task{
			{
				ID:       "test-1",
				Text:     "Categorized task",
				Status:   StatusTodo,
				Type:     TypeCreative,
				Effort:   EffortDeepWork,
				Location: LocationHome,
			},
		},
	}

	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var loaded TasksFile
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(loaded.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(loaded.Tasks))
	}
	task := loaded.Tasks[0]
	if task.Type != TypeCreative {
		t.Errorf("expected type %q, got %q", TypeCreative, task.Type)
	}
	if task.Effort != EffortDeepWork {
		t.Errorf("expected effort %q, got %q", EffortDeepWork, task.Effort)
	}
	if task.Location != LocationHome {
		t.Errorf("expected location %q, got %q", LocationHome, task.Location)
	}
}

func TestTask_YAML_RoundTrip_Uncategorized(t *testing.T) {
	original := &TasksFile{
		Tasks: []*Task{
			{
				ID:     "test-2",
				Text:   "Uncategorized task",
				Status: StatusTodo,
			},
		},
	}

	data, err := yaml.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify omitempty: category fields should not appear in YAML
	yamlStr := string(data)
	if contains(yamlStr, "type:") {
		t.Error("omitempty: 'type' should not appear in YAML for uncategorized task")
	}
	if contains(yamlStr, "effort:") {
		t.Error("omitempty: 'effort' should not appear in YAML for uncategorized task")
	}
	if contains(yamlStr, "location:") {
		t.Error("omitempty: 'location' should not appear in YAML for uncategorized task")
	}

	var loaded TasksFile
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	task := loaded.Tasks[0]
	if task.Type != "" || task.Effort != "" || task.Location != "" {
		t.Error("uncategorized task should have empty categorization fields after round-trip")
	}
}

func TestTask_YAML_BackwardCompatibility(t *testing.T) {
	// Legacy YAML without categorization fields
	legacyYAML := `tasks:
  - id: "old-111"
    text: "Learn Go"
    status: todo
  - id: "old-222"
    text: "Build a TUI app"
    status: in-progress
`
	var loaded TasksFile
	if err := yaml.Unmarshal([]byte(legacyYAML), &loaded); err != nil {
		t.Fatalf("Unmarshal legacy YAML failed: %v", err)
	}
	if len(loaded.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(loaded.Tasks))
	}
	for _, task := range loaded.Tasks {
		if task.Type != "" || task.Effort != "" || task.Location != "" {
			t.Errorf("task %q should have empty categorization from legacy YAML", task.ID)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
