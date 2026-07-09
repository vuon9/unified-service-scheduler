// Package testutil provides helpers for tests (migration runner, etc).
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
)

// RunMigrations discovers and runs all .up.sql files from the given directory,
// skipping specified version numbers. Creates a schema_migrations tracking table
// to avoid re-running already-applied migrations.
func RunMigrations(db *sqlx.DB, dir string, skipVersions ...string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	// Create schema_migrations table
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("schema_migrations: %w", err)
	}

	skip := toSet(skipVersions)

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		version := strings.SplitN(name, "_", 2)[0]
		if skip[version] {
			continue
		}

		var count int
		db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if count > 0 {
			continue
		}

		sqlBytes, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		statements := splitSQLStatements(string(sqlBytes))
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := tx.Exec(stmt); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration %s: %w\nSQL: %s", version, err, stmt)
			}
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
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

		if depth == 0 && strings.Contains(current.String(), ";") {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		}
	}

	if remaining := strings.TrimSpace(current.String()); remaining != "" {
		statements = append(statements, remaining)
	}
	return statements
}

func toSet(vals []string) map[string]bool {
	m := make(map[string]bool, len(vals))
	for _, v := range vals {
		m[v] = true
	}
	return m
}
