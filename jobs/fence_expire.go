package jobs

import (
	"github.com/bgentry/que-go"
)

const (
	FenceExpireJob = "ExpireFence"
)

type FenceExpireRequest struct {
}

func fenceExpireJob(j *que.Job) error {

}
