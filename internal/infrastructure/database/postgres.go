// Package database provides PostgreSQL database connectivity and management.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// DB wraps the pgxpool.Pool with additional functionality
type DB struct {
	Pool   *pgxpool.Pool
	config config.DatabaseConfig
	logger *logger.Logger
}

// New creates a new database connection pool
func New(cfg config.DatabaseConfig, log *logger.Logger) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build connection string
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&pool_max_conns=%d",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.SSLMode,
		cfg.MaxConnections,
	)

	// Parse connection config
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MaxConnections / 4) // 25% of max as min
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	// Connect to database
	log.Infof("Connecting to PostgreSQL at %s:%d/%s", cfg.Host, cfg.Port, cfg.Database)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("Successfully connected to PostgreSQL")

	return &DB{
		Pool:   pool,
		config: cfg,
		logger: log,
	}, nil
}

// Close gracefully closes the database connection pool
func (db *DB) Close() {
	if db.Pool != nil {
		db.logger.Info("Closing database connection pool")
		db.Pool.Close()
		db.logger.Info("Database connection pool closed")
	}
}

// Health checks the health of the database connection
func (db *DB) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.Pool.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// Stats returns database pool statistics
func (db *DB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}

// LogStats logs current database pool statistics
func (db *DB) LogStats() {
	stats := db.Stats()
	db.logger.GetZerolog().Info().
		Int32("total_conns", stats.TotalConns()).
		Int32("idle_conns", stats.IdleConns()).
		Int32("acquired_conns", stats.AcquiredConns()).
		Int32("max_conns", stats.MaxConns()).
		Msg("Database pool statistics")
}

// Exec executes a query without returning rows
func (db *DB) Exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := db.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}
	return nil
}

// Query executes a query that returns rows
func (db *DB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	rows, err := db.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return rows, nil
}

// QueryRow executes a query that returns at most one row
func (db *DB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return db.Pool.QueryRow(ctx, sql, args...)
}

// Transaction represents a database transaction
type Transaction struct {
	tx     pgx.Tx
	logger *logger.Logger
}

// BeginTx starts a new transaction
func (db *DB) BeginTx(ctx context.Context) (*Transaction, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Transaction{
		tx:     tx,
		logger: db.logger,
	}, nil
}

// Exec executes a query within the transaction
func (t *Transaction) Exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := t.tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("transaction exec error: %w", err)
	}
	return nil
}

// Query executes a query within the transaction
func (t *Transaction) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	rows, err := t.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("transaction query error: %w", err)
	}
	return rows, nil
}

// QueryRow executes a query that returns at most one row within the transaction
func (t *Transaction) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return t.tx.QueryRow(ctx, sql, args...)
}

// Commit commits the transaction
func (t *Transaction) Commit(ctx context.Context) error {
	if err := t.tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	t.logger.Debug("Transaction committed")
	return nil
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback(ctx context.Context) error {
	if err := t.tx.Rollback(ctx); err != nil {
		// Ignore "already closed" errors
		if err.Error() == "tx is closed" {
			return nil
		}
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	t.logger.Debug("Transaction rolled back")
	return nil
}

// WithTransaction executes a function within a transaction
// Automatically commits on success, rolls back on error
func (db *DB) WithTransaction(ctx context.Context, fn func(*Transaction) error) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}

	// Ensure rollback on panic or error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			db.logger.Error(rbErr, "failed to rollback transaction")
		}
		return err
	}

	return tx.Commit(ctx)
}

// MustConnect creates a new database connection or panics
func MustConnect(cfg config.DatabaseConfig, log *logger.Logger) *DB {
	db, err := New(cfg, log)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}
