package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// SpendRecord represents an individual completion transaction logged to SQLite.
type SpendRecord struct {
	ID               int64     `json:"id"`
	Timestamp        time.Time `json:"timestamp"`
	Model            string    `json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	LatencyMs        float64   `json:"latency_ms"`
	Cached           bool      `json:"cached"`
	ClientID         string    `json:"client_id"`
}

// SpendSummary aggregates token spend and latency statistics.
type SpendSummary struct {
	TotalRequests         int     `json:"total_requests"`
	TotalPromptTokens     int     `json:"total_prompt_tokens"`
	TotalCompletionTokens int     `json:"total_completion_tokens"`
	TotalTokens           int     `json:"total_tokens"`
	AverageLatencyMs      float64 `json:"average_latency_ms"`
	CacheHitRatio         float64 `json:"cache_hit_ratio"`
}

// DB wraps the embedded SQLite database connection.
type DB struct {
	mu sync.Mutex
	db *sql.DB
}

// InitDB initializes the SQLite database at the specified path and runs migrations.
func InitDB(dbPath string) (*DB, error) {
	if dbPath == "" {
		dbPath = "data/tokman.db"
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Configure connection pool for embedded usage
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	// Enable WAL mode for high concurrent read performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS spend_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		model TEXT NOT NULL,
		prompt_tokens INTEGER DEFAULT 0,
		completion_tokens INTEGER DEFAULT 0,
		total_tokens INTEGER DEFAULT 0,
		latency_ms REAL DEFAULT 0.0,
		cached BOOLEAN DEFAULT 0,
		client_id TEXT DEFAULT 'local'
	);
	CREATE INDEX IF NOT EXISTS idx_spend_timestamp ON spend_logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_spend_model ON spend_logs(model);
	`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &DB{db: db}, nil
}

// LogSpend records an AI request transaction into SQLite.
func (d *DB) LogSpend(rec SpendRecord) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	query := `
	INSERT INTO spend_logs (model, prompt_tokens, completion_tokens, total_tokens, latency_ms, cached, client_id)
	VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	total := rec.PromptTokens + rec.CompletionTokens
	if rec.TotalTokens > 0 {
		total = rec.TotalTokens
	}
	clientID := rec.ClientID
	if clientID == "" {
		clientID = "local"
	}

	_, err := d.db.Exec(query, rec.Model, rec.PromptTokens, rec.CompletionTokens, total, rec.LatencyMs, rec.Cached, clientID)
	return err
}

// GetRecentSpendLogs returns the N most recent spend logs.
func (d *DB) GetRecentSpendLogs(limit int) ([]SpendRecord, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if limit <= 0 {
		limit = 50
	}

	query := `
	SELECT id, timestamp, model, prompt_tokens, completion_tokens, total_tokens, latency_ms, cached, client_id
	FROM spend_logs
	ORDER BY id DESC
	LIMIT ?;
	`
	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []SpendRecord
	for rows.Next() {
		var rec SpendRecord
		var ts string
		if err := rows.Scan(&rec.ID, &ts, &rec.Model, &rec.PromptTokens, &rec.CompletionTokens, &rec.TotalTokens, &rec.LatencyMs, &rec.Cached, &rec.ClientID); err != nil {
			return nil, err
		}
		rec.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		records = append(records, rec)
	}

	return records, rows.Err()
}

// GetSpendSummary returns aggregated metrics across all recorded spend logs.
func (d *DB) GetSpendSummary() (*SpendSummary, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	query := `
	SELECT 
		COUNT(*),
		COALESCE(SUM(prompt_tokens), 0),
		COALESCE(SUM(completion_tokens), 0),
		COALESCE(SUM(total_tokens), 0),
		COALESCE(AVG(latency_ms), 0.0),
		COALESCE(AVG(CASE WHEN cached = 1 THEN 1.0 ELSE 0.0 END), 0.0)
	FROM spend_logs;
	`
	var summary SpendSummary
	row := d.db.QueryRow(query)
	err := row.Scan(
		&summary.TotalRequests,
		&summary.TotalPromptTokens,
		&summary.TotalCompletionTokens,
		&summary.TotalTokens,
		&summary.AverageLatencyMs,
		&summary.CacheHitRatio,
	)
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

// Close closes the underlying database connection.
func (d *DB) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.db.Close()
}
