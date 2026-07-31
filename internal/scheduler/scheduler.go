package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/orchestrator"
	"github.com/singh-anurag-7991/cloud-guard/internal/storage"
)

type Scheduler struct {
	Orchestrator *orchestrator.Orchestrator
	DB           *storage.DB
	Interval     time.Duration
}

func New(orch *orchestrator.Orchestrator, db *storage.DB, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = 1 * time.Hour
	}
	return &Scheduler{
		Orchestrator: orch,
		DB:           db,
		Interval:     interval,
	}
}

// Start launches the background periodic scan loop
func (s *Scheduler) Start(ctx context.Context) {
	log.Printf("⏰ Scheduler started (Interval: %v)", s.Interval)
	ticker := time.NewTicker(s.Interval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("⏰ Scheduler stopping...")
				ticker.Stop()
				return
			case <-ticker.C:
				log.Println("⏰ Triggering scheduled scans for all connected accounts...")
				if err := s.Orchestrator.ScanAll(ctx); err != nil {
					log.Printf("Scheduled scan error: %v", err)
				}
			}
		}
	}()
}
