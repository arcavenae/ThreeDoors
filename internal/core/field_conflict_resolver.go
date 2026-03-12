package core

import "time"

// FieldCategory classifies task fields for conflict resolution strategy.
type FieldCategory int

const (
	// MetadataField marks fields where the remote source is authoritative
	// (e.g., text, status, context). Remote wins on conflict.
	MetadataField FieldCategory = iota

	// ThreeDoorsField marks fields specific to the ThreeDoors application
	// (e.g., effort, type, location). Local wins on conflict.
	ThreeDoorsField
)

// FieldResolution records the outcome of resolving a single field conflict.
type FieldResolution struct {
	Field       string `json:"field"`
	LocalValue  string `json:"local_value"`
	RemoteValue string `json:"remote_value"`
	Winner      string `json:"winner"` // "local" or "remote"
}

// ConflictResolver resolves field-level conflicts between local and remote tasks.
type ConflictResolver interface {
	// ResolveField returns which side wins for the given field name.
	// Returns "local" or "remote".
	ResolveField(fieldName string) string

	// Category returns the FieldCategory for the given field name.
	Category(fieldName string) FieldCategory
}

// defaultFieldCategories maps task field names to their conflict resolution category.
// Metadata fields → remote-wins (the source system is authoritative).
// ThreeDoors fields → local-wins (ThreeDoors-specific data stays local).
var defaultFieldCategories = map[string]FieldCategory{
	// Metadata fields (remote-wins)
	"text":         MetadataField,
	"status":       MetadataField,
	"context":      MetadataField,
	"notes":        MetadataField,
	"blocker":      MetadataField,
	"depends_on":   MetadataField,
	"completed_at": MetadataField,
	"defer_until":  MetadataField,

	// ThreeDoors-specific fields (local-wins)
	"effort":       ThreeDoorsField,
	"type":         ThreeDoorsField,
	"location":     ThreeDoorsField,
	"dev_dispatch": ThreeDoorsField,
}

// DefaultConflictResolver implements ConflictResolver with the standard
// field category mapping: metadata → remote-wins, ThreeDoors → local-wins.
type DefaultConflictResolver struct {
	categories map[string]FieldCategory
}

// NewDefaultConflictResolver creates a resolver with the default field category map.
func NewDefaultConflictResolver() *DefaultConflictResolver {
	cats := make(map[string]FieldCategory, len(defaultFieldCategories))
	for k, v := range defaultFieldCategories {
		cats[k] = v
	}
	return &DefaultConflictResolver{categories: cats}
}

// ResolveField returns "remote" for metadata fields and "local" for ThreeDoors fields.
// Unknown fields default to "remote" (safer to accept authoritative source).
func (r *DefaultConflictResolver) ResolveField(fieldName string) string {
	cat, ok := r.categories[fieldName]
	if !ok {
		return "remote"
	}
	switch cat {
	case ThreeDoorsField:
		return "local"
	default:
		return "remote"
	}
}

// Category returns the FieldCategory for the given field name.
// Unknown fields are classified as MetadataField.
func (r *DefaultConflictResolver) Category(fieldName string) FieldCategory {
	cat, ok := r.categories[fieldName]
	if !ok {
		return MetadataField
	}
	return cat
}

// ResolveTaskConflict applies field-level resolution to produce a merged task
// and a list of per-field resolutions. It compares each field that differs
// between local and remote, and picks the winner per the resolver's strategy.
func ResolveTaskConflict(resolver ConflictResolver, local, remote *Task) (*Task, []FieldResolution) {
	merged := *remote // start with remote as base (metadata-authoritative)
	var resolutions []FieldResolution

	// Text
	if local.Text != remote.Text {
		winner := resolver.ResolveField("text")
		resolutions = append(resolutions, FieldResolution{
			Field:       "text",
			LocalValue:  local.Text,
			RemoteValue: remote.Text,
			Winner:      winner,
		})
		if winner == "local" {
			merged.Text = local.Text
		}
	}

	// Status
	if local.Status != remote.Status {
		winner := resolver.ResolveField("status")
		resolutions = append(resolutions, FieldResolution{
			Field:       "status",
			LocalValue:  string(local.Status),
			RemoteValue: string(remote.Status),
			Winner:      winner,
		})
		if winner == "local" {
			merged.Status = local.Status
			merged.CompletedAt = local.CompletedAt
		}
	}

	// Context
	if local.Context != remote.Context {
		winner := resolver.ResolveField("context")
		resolutions = append(resolutions, FieldResolution{
			Field:       "context",
			LocalValue:  local.Context,
			RemoteValue: remote.Context,
			Winner:      winner,
		})
		if winner == "local" {
			merged.Context = local.Context
		}
	}

	// Effort (ThreeDoors field)
	if local.Effort != remote.Effort {
		winner := resolver.ResolveField("effort")
		resolutions = append(resolutions, FieldResolution{
			Field:       "effort",
			LocalValue:  string(local.Effort),
			RemoteValue: string(remote.Effort),
			Winner:      winner,
		})
		if winner == "local" {
			merged.Effort = local.Effort
		}
	}

	// Type (ThreeDoors field)
	if local.Type != remote.Type {
		winner := resolver.ResolveField("type")
		resolutions = append(resolutions, FieldResolution{
			Field:       "type",
			LocalValue:  string(local.Type),
			RemoteValue: string(remote.Type),
			Winner:      winner,
		})
		if winner == "local" {
			merged.Type = local.Type
		}
	}

	// Location (ThreeDoors field)
	if local.Location != remote.Location {
		winner := resolver.ResolveField("location")
		resolutions = append(resolutions, FieldResolution{
			Field:       "location",
			LocalValue:  string(local.Location),
			RemoteValue: string(remote.Location),
			Winner:      winner,
		})
		if winner == "local" {
			merged.Location = local.Location
		}
	}

	// Blocker
	if local.Blocker != remote.Blocker {
		winner := resolver.ResolveField("blocker")
		resolutions = append(resolutions, FieldResolution{
			Field:       "blocker",
			LocalValue:  local.Blocker,
			RemoteValue: remote.Blocker,
			Winner:      winner,
		})
		if winner == "local" {
			merged.Blocker = local.Blocker
		}
	}

	// DeferUntil
	localDefer := formatTimePtr(local.DeferUntil)
	remoteDefer := formatTimePtr(remote.DeferUntil)
	if localDefer != remoteDefer {
		winner := resolver.ResolveField("defer_until")
		resolutions = append(resolutions, FieldResolution{
			Field:       "defer_until",
			LocalValue:  localDefer,
			RemoteValue: remoteDefer,
			Winner:      winner,
		})
		if winner == "local" {
			merged.DeferUntil = local.DeferUntil
		}
	}

	// Preserve the most recent UpdatedAt
	if local.UpdatedAt.After(remote.UpdatedAt) {
		merged.UpdatedAt = local.UpdatedAt
	}

	return &merged, resolutions
}

// formatTimePtr formats a *time.Time to a string, returning "" for nil.
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
