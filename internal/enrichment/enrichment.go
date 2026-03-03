package enrichment

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	// Pure Go SQLite driver — no CGO required.
	_ "modernc.org/sqlite"
)

// SchemaVersion tracks the current database schema version for migrations.
const SchemaVersion = 1

// DB wraps a SQLite database providing enrichment storage for task metadata,
// cross-references, learning patterns, and feedback history.
type DB struct {
	db   *sql.DB
	path string
}

// Open creates or opens a SQLite enrichment database at the given path.
// It runs any pending schema migrations automatically.
func Open(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create enrichment dir: %w", err)
	}

	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open enrichment db: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	// Enable foreign keys.
	if _, err := sqlDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	edb := &DB{db: sqlDB, path: path}
	if err := edb.migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("migrate enrichment db: %w", err)
	}

	return edb, nil
}

// Close closes the underlying database connection.
func (edb *DB) Close() error {
	return edb.db.Close()
}

// Path returns the file path of the database.
func (edb *DB) Path() string {
	return edb.path
}

// migrate applies schema migrations up to SchemaVersion.
func (edb *DB) migrate() error {
	// Create the schema_version table if it doesn't exist.
	if _, err := edb.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER NOT NULL,
			applied_at TEXT NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create schema_version table: %w", err)
	}

	current, err := edb.currentVersion()
	if err != nil {
		return fmt.Errorf("get current version: %w", err)
	}

	for v := current + 1; v <= SchemaVersion; v++ {
		if err := edb.applyMigration(v); err != nil {
			return fmt.Errorf("apply migration %d: %w", v, err)
		}
	}

	return nil
}

func (edb *DB) currentVersion() (int, error) {
	var version int
	err := edb.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func (edb *DB) applyMigration(version int) error {
	tx, err := edb.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	switch version {
	case 1:
		if err := edb.migrateV1(tx); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown migration version: %d", version)
	}

	if _, err := tx.Exec(
		"INSERT INTO schema_version (version, applied_at) VALUES (?, ?)",
		version, time.Now().UTC().Format(time.RFC3339),
	); err != nil {
		return fmt.Errorf("record migration version: %w", err)
	}

	return tx.Commit()
}

func (edb *DB) migrateV1(tx *sql.Tx) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS task_metadata (
			task_id TEXT PRIMARY KEY,
			category TEXT NOT NULL DEFAULT '',
			enrichment_tags TEXT NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS cross_references (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_task_id TEXT NOT NULL,
			target_task_id TEXT NOT NULL,
			source_system TEXT NOT NULL DEFAULT '',
			relationship TEXT NOT NULL DEFAULT 'related',
			created_at TEXT NOT NULL,
			UNIQUE(source_task_id, target_task_id)
		)`,
		`CREATE TABLE IF NOT EXISTS learning_patterns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pattern_type TEXT NOT NULL,
			pattern_key TEXT NOT NULL,
			weight REAL NOT NULL DEFAULT 0.0,
			data TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL,
			UNIQUE(pattern_type, pattern_key)
		)`,
		`CREATE TABLE IF NOT EXISTS feedback_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL,
			feedback_type TEXT NOT NULL,
			mood TEXT NOT NULL DEFAULT '',
			comment TEXT NOT NULL DEFAULT '',
			session_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cross_refs_source ON cross_references(source_task_id)`,
		`CREATE INDEX IF NOT EXISTS idx_cross_refs_target ON cross_references(target_task_id)`,
		`CREATE INDEX IF NOT EXISTS idx_feedback_task ON feedback_history(task_id)`,
		`CREATE INDEX IF NOT EXISTS idx_feedback_session ON feedback_history(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_learning_type ON learning_patterns(pattern_type)`,
	}

	for _, stmt := range statements {
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("exec %q: %w", stmt[:40], err)
		}
	}

	return nil
}
