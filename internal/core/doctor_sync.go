package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// syncStalenessThreshold is how long since the last sync before we warn.
	syncStalenessThreshold = 24 * time.Hour

	// walStuckRetries is the retry count at which a WAL entry is considered stuck.
	walStuckRetries = 10

	// walExcessiveBacklog is the entry count at which the WAL queue is considered excessive.
	walExcessiveBacklog = 10000

	// orphanedTmpAge is how old a .tmp file must be to be considered orphaned.
	orphanedTmpAge = 1 * time.Hour
)

// checkSync runs the Sync & Offline Queue category checks.
func (dc *DoctorChecker) checkSync() CategoryResult {
	var checks []CheckResult

	checks = append(checks, dc.checkSyncState())
	checks = append(checks, dc.checkWALQueue())
	checks = append(checks, dc.checkOrphanedTmpFiles())

	return CategoryResult{Checks: checks}
}

// checkSyncState validates sync_state.yaml existence, YAML validity, and staleness.
func (dc *DoctorChecker) checkSyncState() CheckResult {
	result := CheckResult{Name: "Sync state"}

	statePath := filepath.Join(dc.configDir, syncStateFile)
	data, err := ReadFileWithLimit(statePath, MaxConfigFileSize)
	if err != nil {
		if os.IsNotExist(err) {
			result.Status = CheckInfo
			result.Message = "No sync history"
			return result
		}
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot read sync state: %v", err)
		return result
	}

	var state SyncState
	if err := yaml.Unmarshal(data, &state); err != nil {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Sync state is not valid YAML: %v", err)
		return result
	}

	if state.LastSyncTime.IsZero() {
		result.Status = CheckInfo
		result.Message = "No sync history"
		return result
	}

	age := time.Since(state.LastSyncTime)
	if age > syncStalenessThreshold {
		result.Status = CheckWarn
		result.Message = fmt.Sprintf("Last sync: %s ago", formatDuration(age))
		result.Suggestion = "Press S in doors view to trigger sync"
		return result
	}

	result.Status = CheckOK
	result.Message = fmt.Sprintf("Last sync: %s ago", formatDuration(age))
	return result
}

// checkWALQueue validates sync-queue.jsonl for entry count, parse errors, and stuck operations.
func (dc *DoctorChecker) checkWALQueue() CheckResult {
	result := CheckResult{Name: "WAL queue"}

	walPath := filepath.Join(dc.configDir, walFile)
	f, err := os.Open(walPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Status = CheckOK
			result.Message = "No pending WAL entries"
			return result
		}
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot read WAL queue: %v", err)
		return result
	}
	defer f.Close() //nolint:errcheck // best-effort close on read

	scanner := NewLimitedScanner(f, MaxJSONLLineSize)
	var totalEntries int
	var stuckEntries int
	var parseErrors int

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		totalEntries++

		var entry WALEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			parseErrors++
			continue
		}

		if entry.Retries >= walStuckRetries {
			stuckEntries++
		}
	}

	if err := scanner.Err(); err != nil {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Error reading WAL queue: %v", err)
		return result
	}

	if parseErrors > 0 {
		result.Status = CheckWarn
		result.Message = fmt.Sprintf("%d corrupt entries in WAL queue", parseErrors)
		return result
	}

	if stuckEntries > 0 {
		result.Status = CheckWarn
		result.Message = fmt.Sprintf("%d stuck operations (retries >= %d)", stuckEntries, walStuckRetries)
		return result
	}

	if totalEntries > walExcessiveBacklog {
		result.Status = CheckWarn
		result.Message = fmt.Sprintf("Excessive backlog: %d entries", totalEntries)
		return result
	}

	if totalEntries == 0 {
		result.Status = CheckOK
		result.Message = "No pending WAL entries"
	} else {
		result.Status = CheckOK
		result.Message = fmt.Sprintf("%d pending WAL entries", totalEntries)
	}
	return result
}

// checkOrphanedTmpFiles detects .tmp files older than orphanedTmpAge in the config directory.
func (dc *DoctorChecker) checkOrphanedTmpFiles() CheckResult {
	result := CheckResult{Name: "Temp files"}

	matches, err := filepath.Glob(filepath.Join(dc.configDir, "*.tmp"))
	if err != nil {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot scan for temp files: %v", err)
		return result
	}

	now := time.Now().UTC()
	var orphanedPaths []string
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) > orphanedTmpAge {
			orphanedPaths = append(orphanedPaths, match)
		}
	}

	if len(orphanedPaths) > 0 {
		if dc.fix {
			removed := 0
			for _, p := range orphanedPaths {
				if rmErr := os.Remove(p); rmErr == nil {
					removed++
				}
			}
			if removed > 0 {
				result.Status = CheckFixed
				result.Message = fmt.Sprintf("FIXED: removed %d stale .tmp files", removed)
				return result
			}
		}
		result.Status = CheckWarn
		result.Message = fmt.Sprintf("%d orphaned temp files found", len(orphanedPaths))
		result.Suggestion = "Run threedoors doctor --fix"
		return result
	}

	result.Status = CheckOK
	result.Message = "No orphaned temp files"
	return result
}
