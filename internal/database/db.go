package database

import (
	"context"
	"database/sql"
	"fmt"
	"mcloud/internal/config"
	"os"
	"path/filepath"
	"sort"

	_ "modernc.org/sqlite"
)

// DefaultMigrationsDir is the default path to the folder containing SQL migration files
const (
	DefaultMigrationsDir = "internal/database/migrations"
)

// Database wraps the sql.DB connection and provides migration capabilities
type Database struct {
	db *sql.DB // underlying sql.DB connection
}

// Open creates a new Database instance with a connection to the given SQLite file
func Open(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	return &Database{db: db}, nil
}

// ensureMigrationsTable creates the migrations tracking table if it doesn't exist
func (s *Database) ensureMigrationsTable() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// migrationApplied checks if a migration file has already been applied
func (s *Database) migrationApplied(filename string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE filename = ?", filename).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// recordMigration records a migration as applied
func (s *Database) recordMigration(filename string) error {
	_, err := s.db.Exec("INSERT INTO schema_migrations (filename) VALUES (?)", filename)
	return err
}

// Migrate runs all SQL migration files in the migrations directory in order
// It reads all .sql files, sorts them alphabetically, and executes each statement on the database
func (s *Database) Migrate() error {
	// Ensure migrations tracking table exists
	if err := s.ensureMigrationsTable(); err != nil {
		return err
	}

	files, err := os.ReadDir(DefaultMigrationsDir)
	if err != nil {
		return err
	}

	// Collect all .sql files from the migrations directory
	var migrationFiles []string
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".sql" {
			migrationFiles = append(migrationFiles, f.Name())
		}
	}

	// Sort files to ensure migrations run in order (e.g., 001_init.sql, 002_add_users.sql)
	sort.Strings(migrationFiles)
	for _, fname := range migrationFiles {
		// Check if migration file has already been applied
		applied, err := s.migrationApplied(fname)
		if err != nil {
			return err
		}
		if applied {
			fmt.Printf("Skipping already applied migration: %s\n", fname)
			continue
		}

		// Read migration SQL file
		path := filepath.Join(DefaultMigrationsDir, fname)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sqlStmt := string(sqlBytes)
		// Execute migration SQL statement
		if _, err := s.db.Exec(sqlStmt); err != nil {
			return err
		}

		// Record migration as applied
		if err := s.recordMigration(fname); err != nil {
			return err
		}

		// Log successful migration
		fmt.Printf("Applied migration: %s\n", fname)
	}
	fmt.Printf("Migration completed successfully \n")

	return nil
}

// Connect loads config, ensures the database file exists, opens the connection, and runs migrations
// Returns a ready-to-use Database instance with all migrations applied
func Connect() (*sql.DB, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	dbPath := cfg.Store.DBPath
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL&_pragma=synchronous=NORMAL", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Create Database instance
	database := &Database{db: db}

	// Always run migrations to ensure schema is up to date
	if err := database.Migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

// WithTx executes the given function within a database transaction.
// It begins a transaction, calls the function with the transaction,
// and commits or rolls back based on whether an error occurred.
func WithTx(
	ctx context.Context,
	db *sql.DB,
	fn func(tx *sql.Tx) error,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
