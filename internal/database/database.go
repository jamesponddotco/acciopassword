// Package database holds the database connection and functions for the service.
package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"sync"

	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
	_ "github.com/mattn/go-sqlite3" //nolint:revive // SQLite3 driver
	"go.uber.org/zap"
)

const (
	// ErrNilLogger is returned when a nil logger is passed to a function.
	ErrNilLogger xerrors.Error = "logger cannot be nil"

	// ErrEmptyDSN is returned when an empty DSN is passed to a function.
	ErrEmptyDSN xerrors.Error = "dsn cannot be empty"
)

const (
	// CounterTypeDiceware is the counter type for diceware passwords.
	CounterTypeDiceware = "Diceware"

	// CounterTypeRandom is the counter type for random passwords.
	CounterTypeRandom = "Random"

	// CounterTypePIN is the counter type for PINs.
	CounterTypePIN = "PIN"
)

//go:embed schema.sql
var schema string

// DB wraps the database connection and stores the access counter.
type DB struct {
	db     *sql.DB
	logger *zap.Logger
	count  map[string]uint64
	mu     sync.Mutex
}

// Open opens a database connection and returns a DB instance.
func Open(logger *zap.Logger, dsn string) (*DB, error) {
	if logger == nil {
		return nil, ErrNilLogger
	}

	if dsn == "" {
		return nil, ErrEmptyDSN
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	counters := make(map[string]uint64)

	rows, err := db.Query("SELECT type, count FROM counter")
	if err != nil {
		return nil, fmt.Errorf("failed to get access counters: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			counterType string
			count       uint64
		)

		if err := rows.Scan(&counterType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan access counters: %w", err)
		}

		counters[counterType] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get access counters: %w", err)
	}

	return &DB{
		db:     db,
		logger: logger,
		count:  counters,
	}, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	if err := d.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}

// Ping checks if the PostgreSQL database is accessible by executing a simple query.
func (d *DB) Ping() error {
	var result int

	err := d.db.QueryRow("SELECT 1").Scan(&result)
	if err != nil || result != 1 {
		return fmt.Errorf("failed to ping the database: %w", err)
	}

	return nil
}

// Count returns the current access counter for the given type.
func (d *DB) Count(counterType string) uint64 {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.count[counterType]
}

// Increment increments the access counter for the given type and stores the access in the database.
func (d *DB) Increment(counterType string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if err = tx.Rollback(); err != nil {
				d.logger.Error("failed to rollback after panic", zap.Error(err))
			}

			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				d.logger.Error("failed to rollback transaction", zap.Error(rbErr))
			}
		}

		err = tx.Commit()
		if err != nil {
			d.logger.Error("failed to commit transaction", zap.Error(err))
		}
	}()

	stmt, err := tx.Prepare("UPDATE counter SET count = count + 1 WHERE type = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(counterType)
	if err != nil {
		return fmt.Errorf("failed to increment access counter: %w", err)
	}

	d.count[counterType]++

	return nil
}
