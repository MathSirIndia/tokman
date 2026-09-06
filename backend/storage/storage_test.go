package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLRUCache_BasicAndEviction(t *testing.T) {
	cache := NewLRUCache(2, time.Minute)

	cache.Set("k1", "v1")
	cache.Set("k2", "v2")

	if v, ok := cache.Get("k1"); !ok || v != "v1" {
		t.Fatalf("expected k1 to be v1, got %v (ok=%v)", v, ok)
	}

	// Adding 3rd item should evict k2 (since k1 was accessed most recently)
	cache.Set("k3", "v3")

	if _, ok := cache.Get("k2"); ok {
		t.Fatalf("expected k2 to be evicted")
	}
	if v, ok := cache.Get("k1"); !ok || v != "v1" {
		t.Fatalf("expected k1 to remain, got %v", v)
	}
	if v, ok := cache.Get("k3"); !ok || v != "v3" {
		t.Fatalf("expected k3 to be v3, got %v", v)
	}
	if cache.Len() != 2 {
		t.Fatalf("expected len 2, got %d", cache.Len())
	}
}

func TestLRUCache_TTL(t *testing.T) {
	cache := NewLRUCache(10, 20*time.Millisecond)
	cache.Set("expire_me", "temp")

	if v, ok := cache.Get("expire_me"); !ok || v != "temp" {
		t.Fatalf("expected item to be present before expiration")
	}

	time.Sleep(30 * time.Millisecond)

	if _, ok := cache.Get("expire_me"); ok {
		t.Fatalf("expected item to expire after TTL")
	}
}

func TestDB_LogSpendAndSummary(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_tokman.db")
	defer os.Remove(dbPath)

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	rec1 := SpendRecord{
		Model:            "pool/general",
		PromptTokens:     10,
		CompletionTokens: 20,
		LatencyMs:        15.5,
		Cached:           false,
		ClientID:         "user-1",
	}
	if err := db.LogSpend(rec1); err != nil {
		t.Fatalf("failed to log spend 1: %v", err)
	}

	rec2 := SpendRecord{
		Model:            "pool/general",
		PromptTokens:     10,
		CompletionTokens: 20,
		LatencyMs:        0.8,
		Cached:           true,
		ClientID:         "user-1",
	}
	if err := db.LogSpend(rec2); err != nil {
		t.Fatalf("failed to log spend 2: %v", err)
	}

	logs, err := db.GetRecentSpendLogs(10)
	if err != nil {
		t.Fatalf("failed to get spend logs: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 spend records, got %d", len(logs))
	}

	summary, err := db.GetSpendSummary()
	if err != nil {
		t.Fatalf("failed to get summary: %v", err)
	}

	if summary.TotalRequests != 2 {
		t.Fatalf("expected 2 requests, got %d", summary.TotalRequests)
	}
	if summary.TotalTokens != 60 {
		t.Fatalf("expected 60 total tokens, got %d", summary.TotalTokens)
	}
	if summary.CacheHitRatio != 0.5 {
		t.Fatalf("expected 0.5 cache hit ratio, got %f", summary.CacheHitRatio)
	}
}
