package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"tokman/backend/gateway"
	"tokman/backend/storage"
	"tokman/gui"
)

func loadDotEnv(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, "\"'")
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

func main() {
	loadDotEnv(".env")

	headless := flag.Bool("headless", false, "Run in headless daemon mode (no GUI window)")
	guiFlag := flag.Bool("gui", false, "Force GUI window mode")
	port := flag.Int("port", 8000, "HTTP gateway listen port")
	dbPath := flag.String("db", "data/tokman.db", "SQLite persistence database path")
	tabFlag := flag.Int("tab", 0, "Initial tab index to display (0: Endpoints, 1: Console, 2: Cache, 3: Settings)")
	flag.Parse()

	// Load keys from environment
	groqKey := os.Getenv("GROQ_API_KEY")
	masterKey := os.Getenv("LITELLM_MASTER_KEY")
	if masterKey == "" {
		masterKey = "sk-master-internal-network-key"
	}

	// Initialize Embedded Storage and In-Memory Cache
	db, err := storage.InitDB(*dbPath)
	if err != nil {
		log.Fatalf("[FATAL] Failed to initialize SQLite persistence: %v", err)
	}
	defer db.Close()

	cache := storage.NewLRUCache(2000, 24*time.Hour)

	// Initialize and launch native Go HTTP Gateway
	gwServer := gateway.NewServer(gateway.Config{
		Port:        *port,
		MasterKey:   masterKey,
		GroqAPIKey:  groqKey,
		Cache:       cache,
		DB:          db,
	})

	go func() {
		log.Printf("[TOKMAN] Starting native Gateway on http://127.0.0.1:%d", *port)
		if err := gwServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("[WARN] Gateway listener on :%d: %v (connecting to running instance)", *port, err)
		}
	}()

	// Determine if running in a headless environment
	hasDisplay := os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
	runHeadless := *headless || (!hasDisplay && !*guiFlag)

	if runHeadless {
		if !hasDisplay && !*headless {
			log.Printf("[TOKMAN] No display server detected (headless environment). Running in pure CLI daemon mode.")
		} else {
			log.Printf("[TOKMAN] Running in headless daemon mode.")
		}
		log.Printf("[TOKMAN] Gateway listening at: http://localhost:%d/v1", *port)
		log.Printf("[TOKMAN] Health Check: http://localhost:%d/health/readiness", *port)
		log.Printf("[TOKMAN] Press Ctrl+C to terminate.")

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Printf("[TOKMAN] Shutting down cleanly...")
		return
	}

	// Launch Pure-Native Fyne Desktop GUI
	fmt.Println("[TOKMAN] Launching Fyne Native Desktop GUI...")
	gui.Run(gwServer, db, cache, *tabFlag)
}
