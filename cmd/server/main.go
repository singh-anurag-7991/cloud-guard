package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/alerting"
	"github.com/singh-anurag-7991/cloud-guard/internal/scheduler"
	"github.com/singh-anurag-7991/cloud-guard/internal/server"
	"github.com/singh-anurag-7991/cloud-guard/internal/storage"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("☁️  Cloud Guard starting...")

	// Root context for background tasks
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	// ── Config ──────────────────────────────────────────────
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Must live inside the mounted Docker volume (/root/data), otherwise the DB is
	// written to the container's ephemeral layer and every redeploy silently wipes
	// all connected accounts and findings.
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/cloudguard.db"
	}
	if dir := filepath.Dir(dbPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatalf("Failed to create database directory %s: %v", dir, err)
		}
	}

	slackWebhook := os.Getenv("SLACK_WEBHOOK_URL")

	// ── Initialize Dependencies ─────────────────────────────
	db, err := storage.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Database initialized:", dbPath)

	slack := alerting.NewSlackWrapper(slackWebhook)
	if slackWebhook != "" {
		log.Println("✅ Slack alerts enabled")
	} else {
		log.Println("⏭️  Slack alerts disabled (no SLACK_WEBHOOK_URL set)")
	}

	// ── Create Server ───────────────────────────────────────
	srv := server.New(db, slack)

	// ── Start Scheduler ─────────────────────────────────────
	sched := scheduler.New(srv.Orchestrator, db, 1*time.Hour)
	sched.Start(rootCtx)

	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      srv,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second, // scans can take a while
		IdleTimeout:  60 * time.Second,
	}

	// ── Start Server ────────────────────────────────────────
	// Channel to catch OS signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 Server listening on http://localhost:%s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// ── Graceful Shutdown ───────────────────────────────────
	sig := <-quit
	log.Printf("⏹️  Received signal %v, shutting down gracefully...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("👋 Cloud Guard stopped.")
}
