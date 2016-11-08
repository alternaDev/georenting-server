package jobs

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/alternaDev/georenting-server/models"
	"github.com/alternaDev/georenting-server/scores"
	"github.com/bgentry/que-go"
)

const (
	// DeployGCAidJobName is the Name of DeployGCAidJob
	DeployGCAidJobName = "DeployGCAid"
)

var (
	// MinimumGCAmount, users who have less than this will get GC for free.
	MinimumGCAmount = scores.InitialBalance / 2 // TBD
)

// DeployGCAidJob executes a DeployGCAidJob.
func DeployGCAidJob(j *que.Job) error {
	log.Info("Processing DeployGCAidJob")
	users, err := models.FindUsersWithGCLessThan(MinimumGCAmount)
	if err != nil {
		log.Error("Unable to unmarshal job arguments into FenceExpireRequest")
		return err
	}

	for _, user := range users {
		user.Balance = user.Balance + MinimumGCAmount/4 // TBD
		user.Save()
	}

	return nil
}

// QueueDeployGCAidRequest creates a new DeployGCAidJob.
func QueueDeployGCAidRequest() error {
	j := que.Job{
		Type:  DeployGCAidJobName,
		RunAt: time.Now(),
	}

	return QC.Enqueue(&j)
}
