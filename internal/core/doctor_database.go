package core

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	// Pure Go SQLite driver — no CGO required.
	_ "modernc.org/sqlite"

	"github.com/arcavenae/ThreeDoors/internal/enrichment"
)

// checkDatabase runs the Database category checks for enrichment.db.
func (dc *DoctorChecker) checkDatabase() CategoryResult {
	var checks []CheckResult

	dbPath := filepath.Join(dc.configDir, "enrichment.db")

	// Check 1: Database existence and accessibility
	checks = append(checks, dc.checkDBExists(dbPath))

	// Only run deeper checks if the file exists
	if checks[0].Status == CheckOK {
		// Check 2: Schema version
		checks = append(checks, dc.checkDBSchema(dbPath))

		// Check 3: Integrity check
		checks = append(checks, dc.checkDBIntegrity(dbPath))
	}

	return CategoryResult{Checks: checks}
}

// checkDBExists verifies the enrichment database file exists and can be opened.
func (dc *DoctorChecker) checkDBExists(dbPath string) CheckResult {
	result := CheckResult{Name: "Enrichment database"}

	_, err := os.Stat(dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			result.Status = CheckWarn
			result.Message = "Enrichment database not found"
			result.Suggestion = "Will be created on first use"
			return result
		}
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot access enrichment database: %v", err)
		return result
	}

	// Verify we can open it read-only
	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot open enrichment database: %v", err)
		return result
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot connect to enrichment database: %v", err)
		result.Suggestion = "Database may be locked or corrupted"
		return result
	}

	result.Status = CheckOK
	result.Message = "Enrichment database accessible"
	return result
}

// checkDBSchema verifies the enrichment database schema version matches expectations.
func (dc *DoctorChecker) checkDBSchema(dbPath string) CheckResult {
	result := CheckResult{Name: "Schema version"}

	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot open database for schema check: %v", err)
		return result
	}
	defer func() { _ = db.Close() }()

	// Check if schema_version table exists
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_version'").Scan(&tableCount)
	if err != nil {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot query database schema: %v", err)
		return result
	}

	if tableCount == 0 {
		result.Status = CheckWarn
		result.Message = "Schema version table not found"
		result.Suggestion = "Database may need re-initialization"
		return result
	}

	var version int
	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot read schema version: %v", err)
		return result
	}

	if version != enrichment.SchemaVersion {
		result.Status = CheckWarn
		result.Message = fmt.Sprintf("Schema version %d (expected %d)", version, enrichment.SchemaVersion)
		result.Suggestion = "Schema migration may be needed"
		return result
	}

	result.Status = CheckOK
	result.Message = fmt.Sprintf("Schema version: %d", version)
	return result
}

// checkDBIntegrity runs SQLite's PRAGMA integrity_check on the enrichment database.
func (dc *DoctorChecker) checkDBIntegrity(dbPath string) CheckResult {
	result := CheckResult{Name: "Integrity check"}

	db, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Cannot open database for integrity check: %v", err)
		return result
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query("PRAGMA integrity_check")
	if err != nil {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Integrity check failed: %v", err)
		return result
	}
	defer func() { _ = rows.Close() }()

	var issues []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			result.Status = CheckFail
			result.Message = fmt.Sprintf("Cannot read integrity check result: %v", err)
			return result
		}
		if line != "ok" {
			issues = append(issues, line)
		}
	}
	if err := rows.Err(); err != nil {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Integrity check error: %v", err)
		return result
	}

	if len(issues) > 0 {
		result.Status = CheckFail
		result.Message = fmt.Sprintf("Integrity check failed: %s", issues[0])
		result.Suggestion = fmt.Sprintf("Run: sqlite3 %s 'PRAGMA integrity_check'", dbPath)
		return result
	}

	result.Status = CheckOK
	result.Message = "Enrichment DB: healthy"
	return result
}
