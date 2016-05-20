package jobs

import (
	"encoding/json"
	"log"

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

	return nil
}
