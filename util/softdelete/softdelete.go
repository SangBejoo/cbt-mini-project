// Package softdelete provides shared utilities for soft-delete operations.
package softdelete

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrAlreadyDeleted is returned when attempting to soft-delete a row
// that has already been soft-deleted.
var ErrAlreadyDeleted = errors.New("record already deleted")

// Executor abstracts *sql.DB and *sql.Tx so callers can use either.
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// PickExecutor returns tx if non-nil, otherwise db.
func PickExecutor(tx *sql.Tx, db *sql.DB) Executor {
	if tx != nil {
		return tx
	}
	return db
}

// SoftDelete marks a single row as deleted by setting deleted_at = NOW().
func SoftDelete(ctx context.Context, exec Executor, table string, id int64) error {
	query := fmt.Sprintf(
		`UPDATE %s SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		table,
	)

	result, err := exec.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("soft delete %s id=%d: %w", table, id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("soft delete %s rows affected: %w", table, err)
	}

	if rows == 0 {
		var deletedAt *time.Time
		checkQuery := fmt.Sprintf(`SELECT deleted_at FROM %s WHERE id = $1`, table)
		checkErr := exec.QueryRowContext(ctx, checkQuery, id).Scan(&deletedAt)
		if checkErr == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		if checkErr == nil && deletedAt != nil {
			return ErrAlreadyDeleted
		}
		return sql.ErrNoRows
	}

	return nil
}

// SoftDeleteWhere marks multiple rows as deleted using a custom WHERE clause.
func SoftDeleteWhere(ctx context.Context, exec Executor, table, whereClause string, args ...interface{}) (int64, error) {
	query := fmt.Sprintf(
		`UPDATE %s SET deleted_at = NOW() WHERE %s AND deleted_at IS NULL`,
		table, whereClause,
	)

	result, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("soft delete %s where %s: %w", table, whereClause, err)
	}

	return result.RowsAffected()
}

// CountActive counts non-deleted rows matching the WHERE clause.
func CountActive(ctx context.Context, exec Executor, table, whereClause string, args ...interface{}) (int, error) {
	query := fmt.Sprintf(
		`SELECT COUNT(*) FROM %s WHERE %s AND deleted_at IS NULL`,
		table, whereClause,
	)

	var count int
	err := exec.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active %s where %s: %w", table, whereClause, err)
	}

	return count, nil
}

// ActiveFilter returns "alias.deleted_at IS NULL" for query embedding.
func ActiveFilter(alias string) string {
	if alias != "" {
		return alias + ".deleted_at IS NULL"
	}
	return "deleted_at IS NULL"
}
