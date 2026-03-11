package core

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/arcaven/ThreeDoors/internal/enrichment"
)

func TestDoctorChecker_DatabaseMissing(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkDatabase()

	if len(result.Checks) != 1 {
		t.Fatalf("expected 1 check (existence only), got %d", len(result.Checks))
	}

	check := result.Checks[0]
	if check.Status != CheckWarn {
		t.Errorf("status = %v, want %v", check.Status, CheckWarn)
	}
	if check.Message != "Enrichment database not found" {
		t.Errorf("message = %q, want %q", check.Message, "Enrichment database not found")
	}
	if check.Suggestion != "Will be created on first use" {
		t.Errorf("suggestion = %q, want %q", check.Suggestion, "Will be created on first use")
	}
}

func TestDoctorChecker_DatabaseHealthy(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "enrichment.db")

	// Create a valid enrichment database
	edb, err := enrichment.Open(dbPath)
	if err != nil {
		t.Fatalf("create test db: %v", err)
	}
	_ = edb.Close()

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkDatabase()

	if len(result.Checks) != 3 {
		t.Fatalf("expected 3 checks, got %d", len(result.Checks))
	}

	// Check 1: existence
	if result.Checks[0].Status != CheckOK {
		t.Errorf("existence check = %v, want %v (msg: %s)", result.Checks[0].Status, CheckOK, result.Checks[0].Message)
	}

	// Check 2: schema version
	if result.Checks[1].Status != CheckOK {
		t.Errorf("schema check = %v, want %v (msg: %s)", result.Checks[1].Status, CheckOK, result.Checks[1].Message)
	}
	wantSchemaMsg := "Schema version: 1"
	if result.Checks[1].Message != wantSchemaMsg {
		t.Errorf("schema message = %q, want %q", result.Checks[1].Message, wantSchemaMsg)
	}

	// Check 3: integrity
	if result.Checks[2].Status != CheckOK {
		t.Errorf("integrity check = %v, want %v (msg: %s)", result.Checks[2].Status, CheckOK, result.Checks[2].Message)
	}
	if result.Checks[2].Message != "Enrichment DB: healthy" {
		t.Errorf("integrity message = %q, want %q", result.Checks[2].Message, "Enrichment DB: healthy")
	}
}

func TestDoctorChecker_DatabaseSchemaVersionMismatch(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "enrichment.db")

	// Create a database with wrong schema version
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE schema_version (version INTEGER NOT NULL, applied_at TEXT NOT NULL)`)
	if err != nil {
		t.Fatal(err)
	}
	// Insert a version that doesn't match enrichment.SchemaVersion
	wrongVersion := enrichment.SchemaVersion + 1
	_, err = db.Exec("INSERT INTO schema_version (version, applied_at) VALUES (?, '2026-01-01T00:00:00Z')", wrongVersion)
	if err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkDatabase()

	// Find schema version check
	if len(result.Checks) < 2 {
		t.Fatalf("expected at least 2 checks, got %d", len(result.Checks))
	}

	schemaCheck := result.Checks[1]
	if schemaCheck.Status != CheckWarn {
		t.Errorf("schema version check = %v, want %v (msg: %s)", schemaCheck.Status, CheckWarn, schemaCheck.Message)
	}
}

func TestDoctorChecker_DatabaseNoSchemaTable(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "enrichment.db")

	// Create an empty SQLite database (no schema_version table)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	// Force creation by pinging
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkDatabase()

	if len(result.Checks) < 2 {
		t.Fatalf("expected at least 2 checks, got %d", len(result.Checks))
	}

	schemaCheck := result.Checks[1]
	if schemaCheck.Status != CheckWarn {
		t.Errorf("no schema table check = %v, want %v (msg: %s)", schemaCheck.Status, CheckWarn, schemaCheck.Message)
	}
	if schemaCheck.Message != "Schema version table not found" {
		t.Errorf("message = %q, want %q", schemaCheck.Message, "Schema version table not found")
	}
}

func TestDoctorChecker_DatabaseCorrupt(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "enrichment.db")

	// Write garbage to simulate corruption
	if err := os.WriteFile(dbPath, []byte("this is not a sqlite database at all"), 0o644); err != nil {
		t.Fatal(err)
	}

	dc := &DoctorChecker{configDir: tmpDir}
	result := dc.checkDatabase()

	// Existence check should fail because DB can't be opened/pinged
	if len(result.Checks) < 1 {
		t.Fatal("expected at least 1 check")
	}

	existCheck := result.Checks[0]
	if existCheck.Status != CheckFail {
		t.Errorf("corrupt db existence check = %v, want %v (msg: %s)", existCheck.Status, CheckFail, existCheck.Message)
	}
}

func TestDoctorChecker_DatabaseRegistered(t *testing.T) {
	tmpDir := t.TempDir()
	dc := NewDoctorChecker(tmpDir)
	result := dc.Run()

	// Should have a Database category registered
	var found bool
	for _, cat := range result.Categories {
		if cat.Name == "Database" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Database category not found in registered categories")
	}
}

func TestDoctorChecker_DatabaseIntegrationWithDoctor(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "enrichment.db")

	// Create a healthy database
	edb, err := enrichment.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_ = edb.Close()

	dc := NewDoctorChecker(tmpDir)
	result := dc.Run()

	// Database category should be OK
	dbCat := findCategory(result, "Database")
	if dbCat == nil {
		t.Fatal("Database category not found")
	}
	if dbCat.Status != CheckOK {
		t.Errorf("database category status = %v, want %v", dbCat.Status, CheckOK)
	}

	// All 3 checks should pass
	for _, check := range dbCat.Checks {
		if check.Status != CheckOK {
			t.Errorf("check %q: status = %v, want %v (msg: %s)", check.Name, check.Status, CheckOK, check.Message)
		}
	}
}

func findCategory(result DoctorResult, name string) *CategoryResult {
	for i := range result.Categories {
		if result.Categories[i].Name == name {
			return &result.Categories[i]
		}
	}
	return nil
}
