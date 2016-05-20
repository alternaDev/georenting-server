package jobs

import (
	"encoding/json"
	"log"

	//"github.com/alternaDev/georenting-server/models"
	"github.com/bgentry/que-go"
)

const (
	FenceExpireJobName = "ExpireFence"
)

type FenceExpireRequest struct {
	FenceID int64
}

func FenceExpireJob(j *que.Job) error {
	var fer FenceExpireRequest
	err := json.Unmarshal(j.Args, &fer)
	if err != nil {
		log.Fatal("Unable to unmarshal job arguments into FenceExpireRequest")
		return err
	}

	log.Print("Processing FenceExpireJob")

	//models.DB.Model(&models.Fence{}).Delete(value, where)

	return nil
}

func QueueFenceExpireRequest(r FenceExpireRequest) error {
	enc, err := json.Marshal(r)
	if err != nil {
		return err
	}

	j := que.Job{
		Type: FenceExpireJobName,
		Args: enc,
	}

	return QC.Enqueue(&j)
}
