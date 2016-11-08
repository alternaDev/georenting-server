package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/alternaDev/georenting-server/jobs"
	"github.com/bgentry/que-go"
	"github.com/golang/glog"
)

func init() {
	flag.Parse() // Glog needs this
}

func main() {

	qc := jobs.QC

	wm := que.WorkMap{
		jobs.FenceExpireJobName:     jobs.FenceExpireJob,
		jobs.SendGcmJobName:         jobs.SendGcmJob,
		jobs.RecordVisitJobName:     jobs.RecordVisitJob,
		jobs.NotifyUsersSyncJobName: jobs.NotifyUsersSyncJob,
	}

	// 2 worker go routines
	workers := que.NewWorkerPool(qc, wm, 2)

	// Catch signal so we can shutdown gracefully
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go workers.Start()

	// Wait for a signal
	sig := <-sigCh
	glog.Infof("Signal '%s' received. Shutting down.", sig)

	workers.Shutdown()
}
