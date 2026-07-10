package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// Repository holds the database connection and provides data access methods.
type Repository struct {
	DB *sqlx.DB
}

// New opens a SQLite database at the given path, enables WAL mode,
// runs pending migrations, and returns a Repository.
func New(dbPath string, migrationsDir string) (*Repository, error) {
	// Ensure the directory for the database exists
	if dir := filepath.Dir(dbPath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create db directory: %w", err)
		}
	}

	// Enable WAL mode, foreign keys, and immediate transaction locking via pragmas in the DSN
	dsn := dbPath + "?_journal_mode=WAL&_foreign_keys=on&_txlock=immediate"
	db, err := sqlx.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite serializes writes; 1 is safest for MVP

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("database connected", "path", dbPath)

	if err := runMigrations(db.DB, migrationsDir); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &Repository{DB: db}, nil
}

// Close closes the database connection.
func (r *Repository) Close() error {
	return r.DB.Close()
}

// runMigrations applies SQL migration files from the given directory in order.
// It uses a simple version-tracking approach with a schema_migrations table.
func runMigrations(db *sql.DB, migrationsDir string) error {
	// Ensure migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		slog.Warn("migrations directory not found, skipping", "dir", migrationsDir)
		return nil
	}

	// Create schema_migrations table if not exists
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("failed to create schema_migrations: %w", err)
	}

	// Read all .up.sql files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations dir: %w", err)
	}

	// Collect up migration files
	type migration struct {
		version string
		path    string
	}
	var migrations []migration
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		// Extract version prefix (e.g., "001" from "001_initial_schema.up.sql")
		version := strings.SplitN(name, "_", 2)[0]
		migrations = append(migrations, migration{
			version: version,
			path:    filepath.Join(migrationsDir, name),
		})
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	// Apply each pending migration
	for _, m := range migrations {
		// Check if already applied
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.version).Scan(&count); err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}
		if count > 0 {
			slog.Info("migration already applied", "version", m.version)
			continue
		}

		// Read, split by semicolons, and execute each statement
		sqlBytes, err := os.ReadFile(m.path)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", m.version, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin tx for migration %s: %w", m.version, err)
		}

		statements := splitSQLStatements(string(sqlBytes))
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := tx.Exec(stmt); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to apply migration %s: %w\nSQL: %s", m.version, err, stmt)
			}
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", m.version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", m.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", m.version, err)
		}

		slog.Info("migration applied", "version", m.version)
	}

	return nil
}

// splitSQLStatements splits a SQL script into individual statements,
// handling semicolons inside CREATE TRIGGER/VIEW/PROCEDURE blocks.
func splitSQLStatements(sql string) []string {
	var statements []string
	depth := 0
	current := strings.Builder{}

	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "BEGIN") {
			depth++
		} else if strings.HasPrefix(trimmed, "END") {
			depth--
		}

		if current.Len() > 0 {
			current.WriteString("\n")
		}
		current.WriteString(line)

		// Only split on ; when not inside a BEGIN..END block
		if depth == 0 && strings.Contains(current.String(), ";") {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	// Flush any remaining
	if remaining := strings.TrimSpace(current.String()); remaining != "" {
		statements = append(statements, remaining)
	}

	return statements
}

// DBTX is an interface satisfied by both *sqlx.DB and *sqlx.Tx.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
}
