package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alternaDev/georenting-server/jobs"
	"github.com/bgentry/que-go"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	_, qc := jobs.Setup(dbURL)

	wm := que.WorkMap{
		jobs.FenceExpireJobName: jobs.FenceExpireJob,
	}

	// 2 worker go routines
	workers := que.NewWorkerPool(qc, wm, 2)

	// Catch signal so we can shutdown gracefully
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go workers.Start()

	// Wait for a signal
	sig := <-sigCh
	log.Printf("Signal '%s' received. Shutting down.", sig)

	workers.Shutdown()
}
