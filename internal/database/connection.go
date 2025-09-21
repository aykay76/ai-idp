package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds database configuration
type Config struct {
	URL             string
	MaxConnections  int32
	MinConnections  int32
	MaxIdleTime     time.Duration
	MaxConnLifetime time.Duration
	ConnectTimeout  time.Duration
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig(databaseURL string) *Config {
	return &Config{
		URL:             databaseURL,
		MaxConnections:  25,
		MinConnections:  5,
		MaxIdleTime:     30 * time.Minute,
		MaxConnLifetime: 1 * time.Hour,
		ConnectTimeout:  10 * time.Second,
	}
}

// Pool wraps pgxpool.Pool with additional utilities
type Pool struct {
	*pgxpool.Pool
	config *Config
}

// NewPool creates a new database connection pool
func NewPool(ctx context.Context, config *Config) (*Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Apply configuration
	poolConfig.MaxConns = config.MaxConnections
	poolConfig.MinConns = config.MinConnections
	poolConfig.MaxConnIdleTime = config.MaxIdleTime
	poolConfig.MaxConnLifetime = config.MaxConnLifetime

	// Connection timeout
	connectCtx, cancel := context.WithTimeout(ctx, config.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{
		Pool:   pool,
		config: config,
	}, nil
}

// HealthCheck returns database health information
func (p *Pool) HealthCheck(ctx context.Context) error {
	// Check if we can ping the database
	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check connection pool stats
	stats := p.Stat()
	if stats.TotalConns() == 0 {
		return fmt.Errorf("no database connections available")
	}

	return nil
}

// Stats returns connection pool statistics
func (p *Pool) Stats() ConnectionStats {
	stats := p.Stat()
	return ConnectionStats{
		TotalConnections:     stats.TotalConns(),
		IdleConnections:      stats.IdleConns(),
		UsedConnections:      stats.AcquiredConns(),
		NewConnectionsCount:  stats.NewConnsCount(),
		MaxLifetimeDestroys:  stats.MaxLifetimeDestroyCount(),
		MaxIdleDestroys:      stats.MaxIdleDestroyCount(),
		AcquireCount:         stats.AcquireCount(),
		AcquireDuration:      stats.AcquireDuration(),
		EmptyAcquireCount:    stats.EmptyAcquireCount(),
		CanceledAcquireCount: stats.CanceledAcquireCount(),
	}
}

// ConnectionStats provides readable connection pool statistics
type ConnectionStats struct {
	TotalConnections     int32         `json:"total_connections"`
	IdleConnections      int32         `json:"idle_connections"`
	UsedConnections      int32         `json:"used_connections"`
	NewConnectionsCount  int64         `json:"new_connections_count"`
	MaxLifetimeDestroys  int64         `json:"max_lifetime_destroys"`
	MaxIdleDestroys      int64         `json:"max_idle_destroys"`
	AcquireCount         int64         `json:"acquire_count"`
	AcquireDuration      time.Duration `json:"acquire_duration"`
	EmptyAcquireCount    int64         `json:"empty_acquire_count"`
	CanceledAcquireCount int64         `json:"canceled_acquire_count"`
}

// Transaction represents a database transaction with utilities
type Transaction struct {
	pgx.Tx
	committed  bool
	rolledBack bool
}

// BeginTx starts a new transaction
func (p *Pool) BeginTx(ctx context.Context) (*Transaction, error) {
	tx, err := p.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &Transaction{Tx: tx}, nil
}

// Commit commits the transaction
func (tx *Transaction) Commit(ctx context.Context) error {
	if tx.committed || tx.rolledBack {
		return fmt.Errorf("transaction already completed")
	}

	if err := tx.Tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	tx.committed = true
	return nil
}

// Rollback rolls back the transaction
func (tx *Transaction) Rollback(ctx context.Context) error {
	if tx.committed || tx.rolledBack {
		return fmt.Errorf("transaction already completed")
	}

	if err := tx.Tx.Rollback(ctx); err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	tx.rolledBack = true
	return nil
}

// WithTransaction executes a function within a transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func (p *Pool) WithTransaction(ctx context.Context, fn func(*Transaction) error) error {
	tx, err := p.BeginTx(ctx)
	if err != nil {
		return err
	}

	// Ensure transaction is properly closed
	defer func() {
		if !tx.committed && !tx.rolledBack {
			_ = tx.Rollback(ctx) // Best effort rollback
		}
	}()

	// Execute the function
	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("transaction failed: %w (rollback error: %v)", err, rollbackErr)
		}
		return err
	}

	// Commit the transaction
	return tx.Commit(ctx)
}

// QueryBuilder provides utilities for building dynamic queries
type QueryBuilder struct {
	query    string
	args     []interface{}
	ArgIndex int
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		query:    baseQuery,
		args:     make([]interface{}, 0),
		ArgIndex: 1,
	}
}

// AddCondition adds a WHERE condition to the query
func (qb *QueryBuilder) AddCondition(condition string, arg interface{}) *QueryBuilder {
	if qb.ArgIndex == 1 {
		qb.query += " WHERE " + condition
	} else {
		qb.query += " AND " + condition
	}
	qb.args = append(qb.args, arg)
	qb.ArgIndex++
	return qb
}

// AddOptionalCondition adds a condition only if the argument is not nil/empty
func (qb *QueryBuilder) AddOptionalCondition(condition string, arg interface{}) *QueryBuilder {
	if arg == nil {
		return qb
	}

	// Check for empty strings
	if str, ok := arg.(string); ok && str == "" {
		return qb
	}

	// Check for empty slices
	if slice, ok := arg.([]string); ok && len(slice) == 0 {
		return qb
	}

	return qb.AddCondition(condition, arg)
}

// AddOrderBy adds ORDER BY clause
func (qb *QueryBuilder) AddOrderBy(orderBy string) *QueryBuilder {
	qb.query += " ORDER BY " + orderBy
	return qb
}

// AddLimit adds LIMIT clause
func (qb *QueryBuilder) AddLimit(limit int) *QueryBuilder {
	if limit > 0 {
		qb.query += fmt.Sprintf(" LIMIT %d", limit)
	}
	return qb
}

// AddOffset adds OFFSET clause
func (qb *QueryBuilder) AddOffset(offset int) *QueryBuilder {
	if offset > 0 {
		qb.query += fmt.Sprintf(" OFFSET %d", offset)
	}
	return qb
}

// Build returns the final query and arguments
func (qb *QueryBuilder) Build() (string, []interface{}) {
	return qb.query, qb.args
}

// BuildAndQuery executes the query and returns rows
func (qb *QueryBuilder) BuildAndQuery(ctx context.Context, conn Querier) (pgx.Rows, error) {
	query, args := qb.Build()
	return conn.Query(ctx, query, args...)
}

// BuildAndQueryRow executes the query and returns a single row
func (qb *QueryBuilder) BuildAndQueryRow(ctx context.Context, conn Querier) pgx.Row {
	query, args := qb.Build()
	return conn.QueryRow(ctx, query, args...)
}

// Querier interface for database connections and transactions
type Querier interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
}

// Ensure Pool and Transaction implement Querier
var _ Querier = (*Pool)(nil)
var _ Querier = (*Transaction)(nil)
