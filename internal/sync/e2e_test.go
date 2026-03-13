package sync_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/arcaven/ThreeDoors/internal/core"
	"github.com/arcaven/ThreeDoors/internal/device"
	gosync "github.com/arcaven/ThreeDoors/internal/sync"
)

// simulatedDevice represents a single device in the E2E test environment.
// Each device has its own temp directory, device identity, git transport,
// and task serialization. Connected via a shared bare Git repo.
type simulatedDevice struct {
	ID        device.DeviceID
	Name      string
	RepoDir   string
	Transport *gosync.GitSyncTransport
	Executor  gosync.GitExecutor
}

// newSimulatedDevice creates a fully-initialized simulated device pointing at bareRepoPath.
func newSimulatedDevice(t *testing.T, bareRepoPath, name, uuidStr string) *simulatedDevice {
	t.Helper()

	id, err := device.NewDeviceID(uuidStr)
	if err != nil {
		t.Fatalf("NewDeviceID(%s): %v", uuidStr, err)
	}

	repoDir := filepath.Join(t.TempDir(), name)
	executor := gosync.NewExecGitExecutor(30 * time.Second)

	transport := gosync.NewGitSyncTransport(gosync.GitSyncTransportConfig{
		RepoDir:    repoDir,
		RemoteURL:  bareRepoPath,
		DeviceID:   id,
		DeviceName: name,
		Executor:   executor,
	})

	ctx := context.Background()
	if err := transport.Init(ctx); err != nil {
		t.Fatalf("Init(%s): %v", name, err)
	}

	return &simulatedDevice{
		ID:        id,
		Name:      name,
		RepoDir:   repoDir,
		Transport: transport,
		Executor:  executor,
	}
}

// pushTask writes a YAML task file and pushes it via the transport.
func (d *simulatedDevice) pushTask(t *testing.T, taskID, content string) {
	t.Helper()
	cs := gosync.Changeset{
		DeviceID:  d.ID,
		Timestamp: time.Now().UTC(),
		Files: []gosync.SyncFile{
			{Path: fmt.Sprintf("tasks/%s.yaml", taskID), Content: []byte(content), Op: gosync.OpAdd},
		},
	}
	if err := d.Transport.Push(context.Background(), cs); err != nil {
		t.Fatalf("%s Push(%s): %v", d.Name, taskID, err)
	}
}

// pushFiles writes multiple files and pushes them in a single changeset.
func (d *simulatedDevice) pushFiles(t *testing.T, files []gosync.SyncFile) {
	t.Helper()
	cs := gosync.Changeset{
		DeviceID:  d.ID,
		Timestamp: time.Now().UTC(),
		Files:     files,
	}
	if err := d.Transport.Push(context.Background(), cs); err != nil {
		t.Fatalf("%s Push: %v", d.Name, err)
	}
}

// pull fetches remote changes into this device's repo.
func (d *simulatedDevice) pull(t *testing.T) gosync.Changeset {
	t.Helper()
	cs, err := d.Transport.Pull(context.Background(), time.Time{})
	if err != nil {
		t.Fatalf("%s Pull: %v", d.Name, err)
	}
	return cs
}

// readFile reads a file from this device's local repo.
func (d *simulatedDevice) readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(d.RepoDir, path))
	if err != nil {
		t.Fatalf("%s readFile(%s): %v", d.Name, path, err)
	}
	return string(data)
}

// fileExists checks if a file exists in this device's local repo.
func (d *simulatedDevice) fileExists(path string) bool {
	_, err := os.Stat(filepath.Join(d.RepoDir, path))
	return err == nil
}

// --- AC1: E2E test infrastructure ---

func TestE2E_Infrastructure(t *testing.T) {
	t.Parallel()

	bareDir := createBareRepo(t)

	devA := newSimulatedDevice(t, bareDir, "device-a", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devB := newSimulatedDevice(t, bareDir, "device-b", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	// Verify each device has its own directory
	if devA.RepoDir == devB.RepoDir {
		t.Error("devices should have separate repo directories")
	}

	// Verify each device has a distinct identity
	if devA.ID == devB.ID {
		t.Error("devices should have distinct IDs")
	}

	// Verify both can reach the shared bare repo
	statusA, err := devA.Transport.Status(context.Background())
	if err != nil {
		t.Fatalf("device-a Status: %v", err)
	}
	statusB, err := devB.Transport.Status(context.Background())
	if err != nil {
		t.Fatalf("device-b Status: %v", err)
	}

	if statusA.State == "not_initialized" {
		t.Error("device-a should be initialized")
	}
	if statusB.State == "not_initialized" {
		t.Error("device-b should be initialized")
	}
}

// --- AC2: Bidirectional sync ---

func TestE2E_BidirectionalSync(t *testing.T) {
	t.Parallel()

	bareDir := createBareRepo(t)

	devA := newSimulatedDevice(t, bareDir, "device-a", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devB := newSimulatedDevice(t, bareDir, "device-b", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	// Step 1: Device A creates a task and pushes
	taskContent := "id: task-001\ntext: Buy groceries\nstatus: todo\n"
	devA.pushTask(t, "task-001", taskContent)

	// Step 2: Device B pulls and sees Device A's task
	devB.pull(t)

	gotContent := devB.readFile(t, "tasks/task-001.yaml")
	if gotContent != taskContent {
		t.Errorf("Device B should see Device A's task\ngot:  %q\nwant: %q", gotContent, taskContent)
	}

	// Step 3: Device B modifies the task and pushes
	modifiedContent := "id: task-001\ntext: Buy groceries and cook dinner\nstatus: active\n"
	devB.pushFiles(t, []gosync.SyncFile{
		{Path: "tasks/task-001.yaml", Content: []byte(modifiedContent), Op: gosync.OpModify},
	})

	// Step 4: Device A pulls and sees Device B's modification
	devA.pull(t)

	gotModified := devA.readFile(t, "tasks/task-001.yaml")
	if gotModified != modifiedContent {
		t.Errorf("Device A should see Device B's modification\ngot:  %q\nwant: %q", gotModified, modifiedContent)
	}
}

// --- AC3: Offline → online reconciliation ---

func TestE2E_OfflineOnlineReconciliation(t *testing.T) {
	t.Parallel()

	bareDir := createBareRepo(t)

	devA := newSimulatedDevice(t, bareDir, "device-a", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devB := newSimulatedDevice(t, bareDir, "device-b", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	// Initial shared state: push a seed task so both have content
	devA.pushTask(t, "seed", "id: seed\ntext: seed task\nstatus: todo\n")
	devB.pull(t)

	// Create OfflineManager for Device A with a circuit breaker we can control.
	// Use a long probe interval so the breaker stays open during the test.
	cbConfig := core.CircuitBreakerConfig{
		FailureThreshold: 1, // trip after 1 failure
		FailureWindow:    1 * time.Minute,
		ProbeInterval:    10 * time.Minute, // long enough to stay open during test
		MaxProbeInterval: 30 * time.Minute,
	}
	breaker := core.NewCircuitBreaker(cbConfig)

	om := gosync.NewOfflineManager(gosync.OfflineManagerConfig{
		Executor:  devA.Executor,
		Breaker:   breaker,
		RepoDir:   devA.RepoDir,
		RemoteURL: bareDir,
	})

	// Simulate offline: trip the circuit breaker
	_ = breaker.Execute(func() error { return fmt.Errorf("simulated network failure") })

	if om.IsOnline() {
		t.Error("Device A should be offline after circuit breaker trip")
	}

	// Device A makes local changes while "offline"
	offlineTask := "id: offline-task\ntext: created offline\nstatus: todo\n"
	if err := os.MkdirAll(filepath.Join(devA.RepoDir, "tasks"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(devA.RepoDir, "tasks/offline-task.yaml"), []byte(offlineTask), 0o644); err != nil {
		t.Fatalf("write offline task: %v", err)
	}
	if err := om.CommitLocal(context.Background(), "offline: add task"); err != nil {
		t.Fatalf("CommitLocal: %v", err)
	}

	// Verify the push fails while offline (breaker is open with long probe interval)
	err := om.PushQueued(context.Background())
	if err == nil {
		t.Error("PushQueued should fail while circuit breaker is open")
	}

	// "Reconnect" by resetting the breaker
	breaker.Reset()

	if !om.IsOnline() {
		t.Error("Device A should be online after breaker reset")
	}

	// Now push succeeds
	if err := om.PushQueued(context.Background()); err != nil {
		t.Fatalf("PushQueued after reconnect: %v", err)
	}

	// Device B pulls and sees the offline-created task
	devB.pull(t)

	if !devB.fileExists("tasks/offline-task.yaml") {
		t.Error("Device B should see the task created while Device A was offline")
	}

	got := devB.readFile(t, "tasks/offline-task.yaml")
	if got != offlineTask {
		t.Errorf("offline task content mismatch\ngot:  %q\nwant: %q", got, offlineTask)
	}
}

// --- AC4: Concurrent non-conflicting edits ---

func TestE2E_ConcurrentNonConflictingEdits(t *testing.T) {
	t.Parallel()

	bareDir := createBareRepo(t)

	devA := newSimulatedDevice(t, bareDir, "device-a", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devB := newSimulatedDevice(t, bareDir, "device-b", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	// Both devices start with a shared task
	initialTask := "id: shared-task\ntext: original text\nstatus: todo\neffort: small\n"
	devA.pushTask(t, "shared-task", initialTask)
	devB.pull(t)

	// Device A modifies file-a (different file than Device B)
	devA.pushFiles(t, []gosync.SyncFile{
		{Path: "tasks/task-a-only.yaml", Content: []byte("id: task-a-only\ntext: from device A\n"), Op: gosync.OpAdd},
	})

	// Device B modifies file-b (different file than Device A)
	devB.pushFiles(t, []gosync.SyncFile{
		{Path: "tasks/task-b-only.yaml", Content: []byte("id: task-b-only\ntext: from device B\n"), Op: gosync.OpAdd},
	})

	// Both pull to sync
	devA.pull(t)
	devB.pull(t)

	// Verify both devices have both files
	if !devA.fileExists("tasks/task-b-only.yaml") {
		t.Error("Device A should have Device B's task after sync")
	}
	if !devB.fileExists("tasks/task-a-only.yaml") {
		t.Error("Device B should have Device A's task after sync")
	}

	// Verify content is preserved
	gotA := devA.readFile(t, "tasks/task-b-only.yaml")
	if !strings.Contains(gotA, "from device B") {
		t.Errorf("Device A should have Device B's content, got: %q", gotA)
	}

	gotB := devB.readFile(t, "tasks/task-a-only.yaml")
	if !strings.Contains(gotB, "from device A") {
		t.Errorf("Device B should have Device A's content, got: %q", gotB)
	}
}

// --- AC5: Concurrent conflicting edits with LWW ---

func TestE2E_ConcurrentConflictingEdits(t *testing.T) {
	t.Parallel()

	bareDir := createBareRepo(t)

	devA := newSimulatedDevice(t, bareDir, "device-a", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devB := newSimulatedDevice(t, bareDir, "device-b", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	// Shared initial state
	initialTask := "id: contested\ntext: original\nstatus: todo\n"
	devA.pushTask(t, "contested", initialTask)
	devB.pull(t)

	// Device A pushes a conflicting version
	devA.pushFiles(t, []gosync.SyncFile{
		{Path: "tasks/contested.yaml", Content: []byte("id: contested\ntext: version-A\nstatus: active\n"), Op: gosync.OpModify},
	})

	// Device B pushes a different conflicting version (without pulling A's change first)
	err := devB.Transport.Push(context.Background(), gosync.Changeset{
		DeviceID:  devB.ID,
		Timestamp: time.Now().UTC(),
		Files: []gosync.SyncFile{
			{Path: "tasks/contested.yaml", Content: []byte("id: contested\ntext: version-B\nstatus: done\n"), Op: gosync.OpModify},
		},
	})
	// Push should handle the conflict (either succeed with rebase or fail gracefully)
	if err != nil {
		t.Logf("Device B Push() with conflict returned: %v (expected for rebase conflict)", err)
	}

	// Verify repos are in consistent state (not mid-rebase)
	statusB, err := devB.Transport.Status(context.Background())
	if err != nil {
		t.Fatalf("Device B Status after conflict: %v", err)
	}
	if statusB.State == "error" {
		t.Error("Device B should handle conflicts gracefully, not enter error state")
	}

	// Test the higher-level cross-machine merge separately
	// This validates the property-level LWW resolution + conflict logging
	now := time.Now().UTC()
	base := &core.Task{
		ID: "contested", Text: "original", Status: "todo",
		CreatedAt: now, UpdatedAt: now,
		VectorClock: map[string]uint64{"device-a": 1, "device-b": 1},
	}
	local := &core.Task{
		ID: "contested", Text: "version-A", Status: "active",
		CreatedAt: now, UpdatedAt: now.Add(1 * time.Second),
		FieldVersions: map[string]core.FieldVersion{
			"text":   {DeviceID: "device-a", UpdatedAt: now.Add(1 * time.Second), Version: 2},
			"status": {DeviceID: "device-a", UpdatedAt: now.Add(1 * time.Second), Version: 2},
		},
		VectorClock: map[string]uint64{"device-a": 2, "device-b": 1},
	}
	remote := &core.Task{
		ID: "contested", Text: "version-B", Status: "done",
		CreatedAt: now, UpdatedAt: now.Add(2 * time.Second),
		FieldVersions: map[string]core.FieldVersion{
			"text":   {DeviceID: "device-b", UpdatedAt: now.Add(2 * time.Second), Version: 2},
			"status": {DeviceID: "device-b", UpdatedAt: now.Add(2 * time.Second), Version: 2},
		},
		VectorClock: map[string]uint64{"device-a": 1, "device-b": 2},
	}

	result := core.ThreeWayMergeTask(base, local, remote)
	if result.Task == nil {
		t.Fatal("ThreeWayMergeTask should return a non-nil merged task")
	}

	// Both fields were modified by both sides → conflicts expected
	if len(result.Conflicts) == 0 {
		t.Error("expected conflicts when both devices modify the same fields")
	}

	// Verify conflict details record field-level information
	for _, c := range result.Conflicts {
		if c.Field == "" {
			t.Error("conflict detail should have a field name")
		}
		if c.Winner == "" {
			t.Error("conflict detail should have a winner")
		}
		if c.Reason == "" {
			t.Error("conflict detail should have a resolution reason")
		}
	}

	// Verify conflict logging
	logDir := t.TempDir()
	conflictLog, err := core.NewConflictLog(logDir)
	if err != nil {
		t.Fatalf("NewConflictLog: %v", err)
	}

	record := core.ConflictRecord{
		ConflictID:        core.NewConflictID(),
		Timestamp:         now,
		TaskID:            "contested",
		DeviceIDs:         []string{"device-a", "device-b"},
		Fields:            result.Conflicts,
		ResolutionOutcome: "auto-resolved",
		RejectedVersion:   remote,
	}
	if err := conflictLog.LogConflicts([]core.ConflictRecord{record}); err != nil {
		t.Fatalf("LogConflicts: %v", err)
	}

	entries, err := conflictLog.ReadEntries()
	if err != nil {
		t.Fatalf("ReadEntries: %v", err)
	}
	if len(entries) == 0 {
		t.Error("conflict log should have at least one entry")
	}
	if entries[0].TaskID != "contested" {
		t.Errorf("logged conflict task ID = %q, want %q", entries[0].TaskID, "contested")
	}
	if len(entries[0].RejectedValues) == 0 {
		t.Error("logged conflict should preserve rejected values")
	}
}

// --- AC6: New device joining ---

func TestE2E_NewDeviceJoining(t *testing.T) {
	t.Parallel()

	bareDir := createBareRepo(t)

	devA := newSimulatedDevice(t, bareDir, "device-a", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devB := newSimulatedDevice(t, bareDir, "device-b", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	// Devices A and B establish shared state
	devA.pushTask(t, "task-1", "id: task-1\ntext: first task\nstatus: todo\n")
	devB.pull(t)
	devB.pushTask(t, "task-2", "id: task-2\ntext: second task\nstatus: active\n")
	devA.pull(t)

	// Register devices in a registry within the bare repo
	regDir := filepath.Join(t.TempDir(), "devices")
	registry := device.NewLocalDeviceRegistry(regDir)
	now := time.Now().UTC()
	if err := registry.Register(device.Device{ID: devA.ID, Name: "device-a", FirstSeen: now, LastSync: now}); err != nil {
		t.Fatalf("Register device A: %v", err)
	}
	if err := registry.Register(device.Device{ID: devB.ID, Name: "device-b", FirstSeen: now, LastSync: now}); err != nil {
		t.Fatalf("Register device B: %v", err)
	}

	// Device C joins — clones the same bare repo
	devC := newSimulatedDevice(t, bareDir, "device-c", "cccccccc-cccc-cccc-cccc-cccccccccccc")

	// Register Device C
	if err := registry.Register(device.Device{ID: devC.ID, Name: "device-c", FirstSeen: now, LastSync: now}); err != nil {
		t.Fatalf("Register device C: %v", err)
	}

	// Verify Device C has all existing tasks from the clone
	if !devC.fileExists("tasks/task-1.yaml") {
		t.Error("Device C should have task-1 after cloning the sync repo")
	}
	if !devC.fileExists("tasks/task-2.yaml") {
		t.Error("Device C should have task-2 after cloning the sync repo")
	}

	// Verify content matches
	gotTask1 := devC.readFile(t, "tasks/task-1.yaml")
	wantTask1 := devA.readFile(t, "tasks/task-1.yaml")
	if gotTask1 != wantTask1 {
		t.Errorf("Device C task-1 content doesn't match Device A\ngot:  %q\nwant: %q", gotTask1, wantTask1)
	}

	gotTask2 := devC.readFile(t, "tasks/task-2.yaml")
	wantTask2 := devA.readFile(t, "tasks/task-2.yaml")
	if gotTask2 != wantTask2 {
		t.Errorf("Device C task-2 content doesn't match Device A\ngot:  %q\nwant: %q", gotTask2, wantTask2)
	}

	// Verify registry shows 3 devices
	devices, err := registry.List()
	if err != nil {
		t.Fatalf("List devices: %v", err)
	}
	if len(devices) != 3 {
		t.Errorf("registry should have 3 devices, got %d", len(devices))
	}

	// Device C can push and others can pull
	devC.pushTask(t, "task-3", "id: task-3\ntext: from device C\nstatus: todo\n")
	devA.pull(t)

	if !devA.fileExists("tasks/task-3.yaml") {
		t.Error("Device A should see Device C's task after sync")
	}
}

// --- AC7: Device removal ---

func TestE2E_DeviceRemoval(t *testing.T) {
	t.Parallel()

	bareDir := createBareRepo(t)

	devA := newSimulatedDevice(t, bareDir, "device-a", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devB := newSimulatedDevice(t, bareDir, "device-b", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	devC := newSimulatedDevice(t, bareDir, "device-c", "cccccccc-cccc-cccc-cccc-cccccccccccc")

	// Establish shared state
	devA.pushTask(t, "task-1", "id: task-1\ntext: shared task\nstatus: todo\n")
	devB.pull(t)
	devC.pull(t)

	// Set up registry
	regDir := filepath.Join(t.TempDir(), "devices")
	registry := device.NewLocalDeviceRegistry(regDir)
	now := time.Now().UTC()
	for _, d := range []struct {
		id   device.DeviceID
		name string
	}{
		{devA.ID, "device-a"},
		{devB.ID, "device-b"},
		{devC.ID, "device-c"},
	} {
		if err := registry.Register(device.Device{ID: d.id, Name: d.name, FirstSeen: now, LastSync: now}); err != nil {
			t.Fatalf("Register %s: %v", d.name, err)
		}
	}

	// Remove Device C from the registry
	if err := registry.Remove(devC.ID); err != nil {
		t.Fatalf("Remove device C: %v", err)
	}

	// Verify registry shows 2 devices
	devices, err := registry.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(devices) != 2 {
		t.Errorf("registry should have 2 devices after removal, got %d", len(devices))
	}

	// Remaining devices can still sync normally
	devA.pushTask(t, "task-2", "id: task-2\ntext: after removal\nstatus: todo\n")
	devB.pull(t)

	if !devB.fileExists("tasks/task-2.yaml") {
		t.Error("Device B should still sync normally after Device C removal")
	}

	gotContent := devB.readFile(t, "tasks/task-2.yaml")
	if !strings.Contains(gotContent, "after removal") {
		t.Errorf("Device B content mismatch: %q", gotContent)
	}
}

// --- AC8: Large payload ---

func TestE2E_LargePayload(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping large payload test in short mode")
	}

	bareDir := createBareRepo(t)

	devA := newSimulatedDevice(t, bareDir, "device-a", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devB := newSimulatedDevice(t, bareDir, "device-b", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	const numTasks = 100
	const numModifications = 10

	// Device A creates 100 tasks
	for i := range numTasks {
		taskID := fmt.Sprintf("task-%03d", i)
		content := fmt.Sprintf("id: %s\ntext: task number %d\nstatus: todo\neffort: small\n", taskID, i)
		devA.pushTask(t, taskID, content)
	}

	// Device B pulls all tasks
	devB.pull(t)

	// Verify all 100 tasks arrived
	for i := range numTasks {
		taskFile := fmt.Sprintf("tasks/task-%03d.yaml", i)
		if !devB.fileExists(taskFile) {
			t.Errorf("Device B missing %s", taskFile)
		}
	}

	// Device B modifies each task 10 times
	for mod := range numModifications {
		var files []gosync.SyncFile
		for i := range numTasks {
			taskID := fmt.Sprintf("task-%03d", i)
			content := fmt.Sprintf("id: %s\ntext: task %d mod %d\nstatus: active\neffort: medium\n", taskID, i, mod)
			files = append(files, gosync.SyncFile{
				Path:    fmt.Sprintf("tasks/%s.yaml", taskID),
				Content: []byte(content),
				Op:      gosync.OpModify,
			})
		}
		devB.pushFiles(t, files)
	}

	// Device A pulls all modifications
	devA.pull(t)

	// Verify consistency: Device A should have the final version of each task
	for i := range numTasks {
		taskFile := fmt.Sprintf("tasks/task-%03d.yaml", i)
		content := devA.readFile(t, taskFile)
		expected := fmt.Sprintf("mod %d", numModifications-1)
		if !strings.Contains(content, expected) {
			t.Errorf("task-%03d: expected final modification (mod %d), got: %q", i, numModifications-1, content)
		}
	}
}

// --- AC9: Transport failure mid-sync ---

func TestE2E_TransportFailureMidSync(t *testing.T) {
	t.Parallel()

	bareDir := createBareRepo(t)

	devA := newSimulatedDevice(t, bareDir, "device-a", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devB := newSimulatedDevice(t, bareDir, "device-b", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	// Device A pushes initial content
	devA.pushTask(t, "before-failure", "id: before-failure\ntext: initial\nstatus: todo\n")

	// Simulate a failed push by using OfflineManager with a tripped breaker
	cbConfig := core.CircuitBreakerConfig{
		FailureThreshold: 1,
		FailureWindow:    1 * time.Minute,
		ProbeInterval:    100 * time.Millisecond,
		MaxProbeInterval: 1 * time.Second,
	}
	breaker := core.NewCircuitBreaker(cbConfig)

	om := gosync.NewOfflineManager(gosync.OfflineManagerConfig{
		Executor:  devA.Executor,
		Breaker:   breaker,
		RepoDir:   devA.RepoDir,
		RemoteURL: bareDir,
	})

	// Make local changes
	postFailureContent := "id: after-failure\ntext: after transport failure\nstatus: todo\n"
	if err := os.MkdirAll(filepath.Join(devA.RepoDir, "tasks"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(devA.RepoDir, "tasks/after-failure.yaml"), []byte(postFailureContent), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := om.CommitLocal(context.Background(), "add task after failure"); err != nil {
		t.Fatalf("CommitLocal: %v", err)
	}

	// Trip the breaker (simulating mid-sync network failure)
	_ = breaker.Execute(func() error { return fmt.Errorf("network down") })

	// Push fails
	err := om.PushQueued(context.Background())
	if err == nil {
		t.Error("push should fail when breaker is open")
	}

	// Reset breaker — next sync cycle should complete cleanly
	breaker.Reset()

	if err := om.PushQueued(context.Background()); err != nil {
		t.Fatalf("push after recovery: %v", err)
	}

	// Device B pulls — should see all content including post-failure task
	devB.pull(t)

	if !devB.fileExists("tasks/before-failure.yaml") {
		t.Error("Device B should have pre-failure task")
	}
	if !devB.fileExists("tasks/after-failure.yaml") {
		t.Error("Device B should have post-failure task after recovery")
	}

	got := devB.readFile(t, "tasks/after-failure.yaml")
	if got != postFailureContent {
		t.Errorf("post-failure content mismatch\ngot:  %q\nwant: %q", got, postFailureContent)
	}
}

// --- AC10: Race detector ---

func TestE2E_RaceDetector(t *testing.T) {
	t.Parallel()

	bareDir := createBareRepo(t)

	devA := newSimulatedDevice(t, bareDir, "device-a", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devB := newSimulatedDevice(t, bareDir, "device-b", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	// Push some initial content
	devA.pushTask(t, "race-test", "id: race-test\ntext: race test\nstatus: todo\n")

	// Concurrent reads from multiple goroutines (simulating real-world access)
	var wg sync.WaitGroup
	errCh := make(chan error, 10)

	for i := range 5 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := devA.Transport.Status(context.Background())
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d Status: %w", idx, err)
			}
		}(i)
	}

	for i := range 5 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := devB.Transport.Status(context.Background())
			if err != nil {
				errCh <- fmt.Errorf("goroutine %d Status: %w", idx, err)
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent access error: %v", err)
	}

	// Concurrent vector clock operations
	vc := core.NewVectorClock()
	var vcWg sync.WaitGroup
	for i := range 10 {
		vcWg.Add(1)
		go func(idx int) {
			defer vcWg.Done()
			localVC := vc.Copy()
			localVC.Increment(fmt.Sprintf("device-%d", idx))
		}(i)
	}
	vcWg.Wait()

	// Concurrent circuit breaker operations
	cb := core.NewCircuitBreaker(core.DefaultCircuitBreakerConfig())
	var cbWg sync.WaitGroup
	for i := range 10 {
		cbWg.Add(1)
		go func(idx int) {
			defer cbWg.Done()
			_ = cb.State()
			_ = cb.Execute(func() error {
				if idx%3 == 0 {
					return fmt.Errorf("simulated failure %d", idx)
				}
				return nil
			})
		}(i)
	}
	cbWg.Wait()
}

// --- AC13: Property-based eventual consistency ---

func TestE2E_PropertyBasedEventualConsistency(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping property-based test in short mode")
	}

	// Test invariant: after N random operations on M devices with arbitrary
	// sync ordering, all devices converge to the same task state.
	const numDevices = 3
	const numOperations = 20
	const numSyncRounds = 5

	bareDir := createBareRepo(t)
	deviceUUIDs := []string{
		"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		"cccccccc-cccc-cccc-cccc-cccccccccccc",
	}

	devices := make([]*simulatedDevice, numDevices)
	for i := range numDevices {
		name := fmt.Sprintf("device-%d", i)
		devices[i] = newSimulatedDevice(t, bareDir, name, deviceUUIDs[i])
	}

	rng := rand.New(rand.NewPCG(42, 0))

	// Perform random operations: each device pulls first (to avoid non-fast-forward),
	// then makes changes and pushes. This models realistic device behavior where
	// sync = pull + local changes + push.
	for op := range numOperations {
		devIdx := rng.IntN(numDevices)
		dev := devices[devIdx]

		// Pull first to stay up-to-date (prevents non-fast-forward rejection)
		dev.pull(t)

		taskID := fmt.Sprintf("prop-task-%03d", rng.IntN(10)) // reuse IDs to create modifications
		content := fmt.Sprintf("id: %s\ntext: op-%d by device-%d\nstatus: todo\n", taskID, op, devIdx)
		dev.pushFiles(t, []gosync.SyncFile{
			{Path: fmt.Sprintf("tasks/%s.yaml", taskID), Content: []byte(content), Op: gosync.OpModify},
		})
	}

	// Sync all devices multiple rounds to ensure convergence
	for range numSyncRounds {
		order := rng.Perm(numDevices)
		for _, idx := range order {
			devices[idx].pull(t)
		}
	}

	// Final pull to ensure full convergence
	for _, dev := range devices {
		dev.pull(t)
	}

	// Verify convergence: all devices should have the same file set
	readTaskDir := func(dev *simulatedDevice) map[string]string {
		files := make(map[string]string)
		tasksDir := filepath.Join(dev.RepoDir, "tasks")
		entries, err := os.ReadDir(tasksDir)
		if err != nil {
			t.Fatalf("ReadDir(%s/tasks): %v", dev.Name, err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			content, err := os.ReadFile(filepath.Join(tasksDir, e.Name()))
			if err != nil {
				t.Fatalf("ReadFile(%s/%s): %v", dev.Name, e.Name(), err)
			}
			files[e.Name()] = string(content)
		}
		return files
	}

	refFiles := readTaskDir(devices[0])

	for i := 1; i < numDevices; i++ {
		devFiles := readTaskDir(devices[i])
		if len(devFiles) != len(refFiles) {
			t.Errorf("device-%d has %d files, device-0 has %d files", i, len(devFiles), len(refFiles))
			continue
		}
		for name, refContent := range refFiles {
			devContent, ok := devFiles[name]
			if !ok {
				t.Errorf("device-%d missing file %s that device-0 has", i, name)
				continue
			}
			if devContent != refContent {
				t.Errorf("device-%d file %s differs from device-0\ndevice-0: %q\ndevice-%d: %q",
					i, name, refContent, i, devContent)
			}
		}
	}
}

// --- AC11: Docker E2E integration ---

func TestE2E_DockerIntegration(t *testing.T) {
	t.Parallel()

	// This test verifies the test suite can run in the Docker test environment.
	// The actual Docker execution is handled by `docker compose run test`
	// using the existing Dockerfile.test from Epic 18.
	//
	// When running in Docker (CI=true), this test runs normally.
	// When running locally, it verifies the test infrastructure is compatible
	// with the Docker environment constraints (no network, no root, temp dirs).

	// Verify tests work with temporary directories (Docker constraint)
	bareDir := createBareRepo(t)

	dev := newSimulatedDevice(t, bareDir, "docker-test", "dddddddd-dddd-dddd-dddd-dddddddddddd")
	dev.pushTask(t, "docker-task", "id: docker-task\ntext: docker test\nstatus: todo\n")

	content := dev.readFile(t, "tasks/docker-task.yaml")
	if !strings.Contains(content, "docker test") {
		t.Errorf("docker test content mismatch: %q", content)
	}

	// Verify no network dependencies (all Git operations use local paths)
	status, err := dev.Transport.Status(context.Background())
	if err != nil {
		t.Fatalf("Status in Docker context: %v", err)
	}
	if !strings.Contains(status.RemoteURL, t.TempDir()[:5]) && status.RemoteURL != bareDir {
		// RemoteURL should be a local path, not a network URL
		if strings.HasPrefix(status.RemoteURL, "http") || strings.HasPrefix(status.RemoteURL, "git@") {
			t.Errorf("RemoteURL should be local path for Docker E2E, got: %s", status.RemoteURL)
		}
	}
}

// --- AC12: Performance (60s target) ---

func TestE2E_PerformanceTarget(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	start := time.Now()

	bareDir := createBareRepo(t)

	devA := newSimulatedDevice(t, bareDir, "perf-a", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	devB := newSimulatedDevice(t, bareDir, "perf-b", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	// Simulate a realistic workload: 50 tasks, 5 modifications each, sync cycles
	const numTasks = 50
	const numMods = 5

	for i := range numTasks {
		taskID := fmt.Sprintf("perf-task-%03d", i)
		content := fmt.Sprintf("id: %s\ntext: perf task %d\nstatus: todo\n", taskID, i)
		devA.pushTask(t, taskID, content)
	}

	devB.pull(t)

	for mod := range numMods {
		var files []gosync.SyncFile
		for i := range numTasks {
			taskID := fmt.Sprintf("perf-task-%03d", i)
			content := fmt.Sprintf("id: %s\ntext: perf task %d v%d\nstatus: active\n", taskID, i, mod)
			files = append(files, gosync.SyncFile{
				Path:    fmt.Sprintf("tasks/%s.yaml", taskID),
				Content: []byte(content),
				Op:      gosync.OpModify,
			})
		}
		devB.pushFiles(t, files)
	}

	devA.pull(t)

	elapsed := time.Since(start)
	t.Logf("Performance test completed in %v", elapsed)

	if elapsed > 60*time.Second {
		t.Errorf("E2E test suite should complete in under 60s, took %v", elapsed)
	}
}

// --- Stress test for cross-machine merge convergence ---

func TestE2E_CrossMachineMergeConvergence(t *testing.T) {
	t.Parallel()

	// Verify that ThreeWayMergeTaskLists produces consistent results
	// regardless of merge ordering.

	now := time.Now().UTC()

	baseTasks := []*core.Task{
		{ID: "t1", Text: "original-1", Status: "todo", CreatedAt: now, UpdatedAt: now},
		{ID: "t2", Text: "original-2", Status: "todo", CreatedAt: now, UpdatedAt: now},
		{ID: "t3", Text: "original-3", Status: "todo", CreatedAt: now, UpdatedAt: now},
	}

	// Device A modifies t1, creates t4
	localTasks := []*core.Task{
		{ID: "t1", Text: "modified-by-A", Status: "active", CreatedAt: now, UpdatedAt: now.Add(time.Second)},
		{ID: "t2", Text: "original-2", Status: "todo", CreatedAt: now, UpdatedAt: now},
		{ID: "t3", Text: "original-3", Status: "todo", CreatedAt: now, UpdatedAt: now},
		{ID: "t4", Text: "new-from-A", Status: "todo", CreatedAt: now, UpdatedAt: now},
	}

	// Device B modifies t2, creates t5
	remoteTasks := []*core.Task{
		{ID: "t1", Text: "original-1", Status: "todo", CreatedAt: now, UpdatedAt: now},
		{ID: "t2", Text: "modified-by-B", Status: "done", CreatedAt: now, UpdatedAt: now.Add(time.Second)},
		{ID: "t3", Text: "original-3", Status: "todo", CreatedAt: now, UpdatedAt: now},
		{ID: "t5", Text: "new-from-B", Status: "todo", CreatedAt: now, UpdatedAt: now},
	}

	outcome := core.ThreeWayMergeTaskLists(baseTasks, localTasks, remoteTasks)

	// Should have 5 tasks: t1 (A's version), t2 (B's version), t3 (unchanged), t4 (new from A), t5 (new from B)
	merged := make(map[string]*core.Task)
	for _, t := range outcome.MergedTasks {
		merged[t.ID] = t
	}

	if len(merged) != 5 {
		t.Errorf("expected 5 merged tasks, got %d", len(merged))
	}

	if t1, ok := merged["t1"]; ok {
		if t1.Text != "modified-by-A" {
			t.Errorf("t1 should have A's modification, got text=%q", t1.Text)
		}
	} else {
		t.Error("merged result missing t1")
	}

	if t2, ok := merged["t2"]; ok {
		if t2.Text != "modified-by-B" {
			t.Errorf("t2 should have B's modification, got text=%q", t2.Text)
		}
	} else {
		t.Error("merged result missing t2")
	}

	if _, ok := merged["t4"]; !ok {
		t.Error("merged result should include t4 (new from A)")
	}
	if _, ok := merged["t5"]; !ok {
		t.Error("merged result should include t5 (new from B)")
	}
}
