package jobs

import (
	"encoding/json"
	"log"

	//"github.com/alternaDev/georenting-server/models"

	"github.com/alternaDev/georenting-server/google/gcm"
	"github.com/alternaDev/georenting-server/models"
	"github.com/alternaDev/georenting-server/models/search"
	"github.com/bgentry/que-go"
)

const (
	FenceExpireJobName = "ExpireFence"
)

type FenceExpireRequest struct {
	FenceID uint
}

func FenceExpireJob(j *que.Job) error {
	var fer FenceExpireRequest
	err := json.Unmarshal(j.Args, &fer)
	if err != nil {
		log.Fatal("Unable to unmarshal job arguments into FenceExpireRequest")
		return err
	}

	log.Print("Processing FenceExpireJob")

	var fence models.Fence

	notFound := models.DB.Find(&fence, fer.FenceID).RecordNotFound()

	if notFound {
		return nil
	}

	err = search.DeleteGeoFence(&fence)

	if err != nil {
		return err
	}

	err = models.DB.Delete(fence).Error

	if err != nil {
		return err
	}

	QueueNotifyUsersSyncRequest(fence.Lat, fence.Lon)
	QueueSendGcmRequest(gcm.NewMessage(map[string]interface{}{"type": "onFenceExpired", "fenceId": fence.ID, "fenceName": fence.Name}, fence.User.GCMNotificationID))

	return nil
}

func QueueFenceExpireRequest(fence *models.Fence) error {
	enc, err := json.Marshal(&FenceExpireRequest{FenceID: fence.ID})
	if err != nil {
		return err
	}

	j := que.Job{
		Type:  FenceExpireJobName,
		Args:  enc,
		RunAt: fence.DiesAt,
	}

	return QC.Enqueue(&j)
}
