package core

import (
	"fmt"
	"time"
)

// MergeResult holds the outcome of a three-way cross-machine merge for a single task.
type MergeResult struct {
	Task         *Task
	Conflicts    []FieldConflictDetail
	AutoMerged   bool
	ManualNeeded bool
}

// FieldConflictDetail records per-field conflict information for logging.
type FieldConflictDetail struct {
	Field        string `json:"field"`
	BaseValue    string `json:"base_value"`
	LocalValue   string `json:"local_value"`
	RemoteValue  string `json:"remote_value"`
	Winner       string `json:"winner"` // "local" or "remote"
	WinnerDevice string `json:"winner_device"`
	Reason       string `json:"reason"` // "causal", "timestamp", "tiebreaker"
}

// TaskMergeOutcome represents the result of merging entire task lists.
type TaskMergeOutcome struct {
	MergedTasks []*Task
	Conflicts   []ConflictRecord
}

// ConflictRecord captures a conflict for logging to conflicts.jsonl.
type ConflictRecord struct {
	ConflictID        string                `json:"conflict_id"`
	Timestamp         time.Time             `json:"timestamp"`
	TaskID            string                `json:"task_id"`
	DeviceIDs         []string              `json:"device_ids"`
	Fields            []FieldConflictDetail `json:"fields"`
	ResolutionOutcome string                `json:"resolution_outcome"`
	RejectedVersion   *Task                 `json:"rejected_version,omitempty"`
}

// CrossMachineConflictResolver resolves field-level conflicts using vector clocks,
// timestamps, and device ID tiebreaker for cross-machine sync scenarios.
type CrossMachineConflictResolver struct{}

// NewCrossMachineConflictResolver creates a new resolver for cross-machine conflicts.
func NewCrossMachineConflictResolver() *CrossMachineConflictResolver {
	return &CrossMachineConflictResolver{}
}

// ResolveField returns the winner for a given field based on the cross-machine strategy.
// For cross-machine resolution, this method is not used directly — use ResolveFieldWithContext instead.
// Falls back to "remote" for interface compliance.
func (r *CrossMachineConflictResolver) ResolveField(_ string) string {
	return "remote"
}

// Category returns MetadataField for all fields in cross-machine context.
// In cross-machine sync, field categories don't apply — all conflicts use vector clock strategy.
func (r *CrossMachineConflictResolver) Category(_ string) FieldCategory {
	return MetadataField
}

// ResolveFieldConflict determines the winner for a specific field using the cascade:
// 1. Vector clock causal ordering
// 2. UTC timestamp comparison
// 3. Device ID lexicographic tiebreaker
func ResolveFieldConflict(
	localVersion, remoteVersion FieldVersion,
	localClock, remoteClock VectorClock,
) (winner string, reason string) {
	// Step 1: Vector clock causal ordering
	ordering := localClock.Compare(remoteClock)
	switch ordering {
	case HappenedAfter:
		return "local", "causal"
	case HappenedBefore:
		return "remote", "causal"
	case Equal:
		// Identical clocks — fall through to timestamp
	case Concurrent:
		// Concurrent — fall through to timestamp
	}

	// Step 2: UTC timestamp comparison
	if localVersion.UpdatedAt.After(remoteVersion.UpdatedAt) {
		return "local", "timestamp"
	}
	if remoteVersion.UpdatedAt.After(localVersion.UpdatedAt) {
		return "remote", "timestamp"
	}

	// Step 3: Device ID lexicographic tiebreaker (higher wins)
	if localVersion.DeviceID > remoteVersion.DeviceID {
		return "local", "tiebreaker"
	}
	return "remote", "tiebreaker"
}

// ThreeWayMergeTask performs a three-way merge of a single task.
// base is the common ancestor, local and remote are the diverged versions.
// Returns the merged task and any conflict details.
func ThreeWayMergeTask(base, local, remote *Task) MergeResult {
	if base == nil {
		// Concurrent create: both sides created the task
		// Take remote as base, but this shouldn't normally happen for same-ID tasks
		return MergeResult{Task: remote, AutoMerged: true}
	}

	merged := *local // start with local as base for the merge
	var conflicts []FieldConflictDetail

	localClock := VectorClock(local.VectorClock).Copy()
	if localClock == nil {
		localClock = NewVectorClock()
	}
	remoteClock := VectorClock(remote.VectorClock).Copy()
	if remoteClock == nil {
		remoteClock = NewVectorClock()
	}

	// Helper to get field version or zero value
	getVersion := func(versions map[string]FieldVersion, field string) FieldVersion {
		if versions == nil {
			return FieldVersion{}
		}
		return versions[field]
	}

	// Compare each field: base→local diff vs base→remote diff
	type fieldCheck struct {
		name        string
		baseVal     string
		localVal    string
		remoteVal   string
		applyLocal  func()
		applyRemote func()
	}

	checks := []fieldCheck{
		{
			name: "text", baseVal: base.Text, localVal: local.Text, remoteVal: remote.Text,
			applyLocal:  func() { merged.Text = local.Text },
			applyRemote: func() { merged.Text = remote.Text },
		},
		{
			name: "status", baseVal: string(base.Status), localVal: string(local.Status), remoteVal: string(remote.Status),
			applyLocal:  func() { merged.Status = local.Status; merged.CompletedAt = local.CompletedAt },
			applyRemote: func() { merged.Status = remote.Status; merged.CompletedAt = remote.CompletedAt },
		},
		{
			name: "context", baseVal: base.Context, localVal: local.Context, remoteVal: remote.Context,
			applyLocal:  func() { merged.Context = local.Context },
			applyRemote: func() { merged.Context = remote.Context },
		},
		{
			name: "effort", baseVal: string(base.Effort), localVal: string(local.Effort), remoteVal: string(remote.Effort),
			applyLocal:  func() { merged.Effort = local.Effort },
			applyRemote: func() { merged.Effort = remote.Effort },
		},
		{
			name: "type", baseVal: string(base.Type), localVal: string(local.Type), remoteVal: string(remote.Type),
			applyLocal:  func() { merged.Type = local.Type },
			applyRemote: func() { merged.Type = remote.Type },
		},
		{
			name: "location", baseVal: string(base.Location), localVal: string(local.Location), remoteVal: string(remote.Location),
			applyLocal:  func() { merged.Location = local.Location },
			applyRemote: func() { merged.Location = remote.Location },
		},
		{
			name: "blocker", baseVal: base.Blocker, localVal: local.Blocker, remoteVal: remote.Blocker,
			applyLocal:  func() { merged.Blocker = local.Blocker },
			applyRemote: func() { merged.Blocker = remote.Blocker },
		},
	}

	for _, check := range checks {
		localChanged := check.localVal != check.baseVal
		remoteChanged := check.remoteVal != check.baseVal

		if !localChanged && !remoteChanged {
			// No changes — keep as-is
			continue
		}
		if localChanged && !remoteChanged {
			// Only local changed — take local (already in merged)
			check.applyLocal()
			continue
		}
		if !localChanged && remoteChanged {
			// Only remote changed — take remote
			check.applyRemote()
			continue
		}
		// Both changed — overlapping conflict
		if check.localVal == check.remoteVal {
			// Same change on both sides — no real conflict
			continue
		}

		localVer := getVersion(local.FieldVersions, check.name)
		remoteVer := getVersion(remote.FieldVersions, check.name)

		winner, reason := ResolveFieldConflict(localVer, remoteVer, localClock, remoteClock)

		detail := FieldConflictDetail{
			Field:       check.name,
			BaseValue:   check.baseVal,
			LocalValue:  check.localVal,
			RemoteValue: check.remoteVal,
			Winner:      winner,
			Reason:      reason,
		}

		if winner == "local" {
			check.applyLocal()
			detail.WinnerDevice = localVer.DeviceID
		} else {
			check.applyRemote()
			detail.WinnerDevice = remoteVer.DeviceID
		}

		conflicts = append(conflicts, detail)
	}

	// Merge vector clocks
	mergedClock := localClock.Copy()
	mergedClock.Merge(remoteClock)
	merged.VectorClock = map[string]uint64(mergedClock)

	// Merge field versions — take the winning version for each field
	mergedVersions := make(map[string]FieldVersion)
	for field, ver := range local.FieldVersions {
		mergedVersions[field] = ver
	}
	for field, ver := range remote.FieldVersions {
		existing, ok := mergedVersions[field]
		if !ok || ver.Version > existing.Version {
			mergedVersions[field] = ver
		}
	}
	merged.FieldVersions = mergedVersions

	// Preserve the most recent UpdatedAt
	if remote.UpdatedAt.After(local.UpdatedAt) {
		merged.UpdatedAt = remote.UpdatedAt
	}

	return MergeResult{
		Task:       &merged,
		Conflicts:  conflicts,
		AutoMerged: len(conflicts) > 0,
	}
}

// ThreeWayMergeTaskLists performs a three-way merge of task lists.
// base is the common ancestor list, local and remote are diverged versions.
// Implements modify-wins semantics for concurrent delete vs modify.
func ThreeWayMergeTaskLists(base, local, remote []*Task) TaskMergeOutcome {
	baseByID := tasksByID(base)
	localByID := tasksByID(local)
	remoteByID := tasksByID(remote)

	seen := make(map[string]bool)
	var merged []*Task
	var conflicts []ConflictRecord

	// Process all tasks that exist in any of the three lists
	allIDs := make(map[string]bool)
	for id := range baseByID {
		allIDs[id] = true
	}
	for id := range localByID {
		allIDs[id] = true
	}
	for id := range remoteByID {
		allIDs[id] = true
	}

	for id := range allIDs {
		if seen[id] {
			continue
		}
		seen[id] = true

		baseTask := baseByID[id]
		localTask := localByID[id]
		remoteTask := remoteByID[id]

		switch {
		case baseTask == nil && localTask != nil && remoteTask == nil:
			// Local-only create
			merged = append(merged, localTask)

		case baseTask == nil && localTask == nil && remoteTask != nil:
			// Remote-only create
			merged = append(merged, remoteTask)

		case baseTask == nil && localTask != nil && remoteTask != nil:
			// Concurrent create — both sides created same ID (unlikely with UUIDs)
			result := ThreeWayMergeTask(nil, localTask, remoteTask)
			merged = append(merged, result.Task)

		case baseTask != nil && localTask == nil && remoteTask == nil:
			// Both deleted — task is gone

		case baseTask != nil && localTask != nil && remoteTask == nil:
			// Remote deleted, local modified or unchanged
			if tasksEqual(baseTask, localTask) {
				// Local didn't change — honor remote delete
				continue
			}
			// Modify-wins: keep the local version
			merged = append(merged, localTask)

		case baseTask != nil && localTask == nil && remoteTask != nil:
			// Local deleted, remote modified or unchanged
			if tasksEqual(baseTask, remoteTask) {
				// Remote didn't change — honor local delete
				continue
			}
			// Modify-wins: keep the remote version
			merged = append(merged, remoteTask)

		case baseTask != nil && localTask != nil && remoteTask != nil:
			// All three exist — three-way merge
			result := ThreeWayMergeTask(baseTask, localTask, remoteTask)
			merged = append(merged, result.Task)
			if len(result.Conflicts) > 0 {
				deviceIDs := collectDeviceIDs(result.Conflicts)
				conflicts = append(conflicts, ConflictRecord{
					ConflictID:        fmt.Sprintf("conflict-%s-%d", id[:8], time.Now().UTC().UnixNano()),
					Timestamp:         time.Now().UTC(),
					TaskID:            id,
					DeviceIDs:         deviceIDs,
					Fields:            result.Conflicts,
					ResolutionOutcome: "auto-resolved",
				})
			}
		}
	}

	return TaskMergeOutcome{
		MergedTasks: merged,
		Conflicts:   conflicts,
	}
}

// tasksByID builds a lookup map from task ID to task.
func tasksByID(tasks []*Task) map[string]*Task {
	m := make(map[string]*Task, len(tasks))
	for _, t := range tasks {
		m[t.ID] = t
	}
	return m
}

// tasksEqual checks if two tasks have the same content (ignoring timestamps).
func tasksEqual(a, b *Task) bool {
	return a.Text == b.Text &&
		a.Status == b.Status &&
		a.Context == b.Context &&
		a.Effort == b.Effort &&
		a.Type == b.Type &&
		a.Location == b.Location &&
		a.Blocker == b.Blocker
}

// collectDeviceIDs extracts unique device IDs from conflict details.
func collectDeviceIDs(details []FieldConflictDetail) []string {
	seen := make(map[string]bool)
	var ids []string
	for _, d := range details {
		if d.WinnerDevice != "" && !seen[d.WinnerDevice] {
			seen[d.WinnerDevice] = true
			ids = append(ids, d.WinnerDevice)
		}
	}
	return ids
}
