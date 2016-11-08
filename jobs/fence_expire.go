package jobs

import (
	"encoding/json"

	//"github.com/alternaDev/georenting-server/models"

	"github.com/alternaDev/georenting-server/activity"
	"github.com/alternaDev/georenting-server/google/gcm"
	"github.com/alternaDev/georenting-server/models"
	"github.com/alternaDev/georenting-server/models/search"
	"github.com/bgentry/que-go"
	"github.com/golang/glog"
)

const (
	// FenceExpireJobName is the Name of FenceExpireJob
	FenceExpireJobName = "ExpireFence"
)

// FenceExpireRequest holds the Data for a FenceExpireJob.
type FenceExpireRequest struct {
	FenceID int
}

// FenceExpireJob executes a FenceExpireJob.
func FenceExpireJob(j *que.Job) error {
	var fer FenceExpireRequest
	err := json.Unmarshal(j.Args, &fer)
	if err != nil {
		glog.Error("Unable to unmarshal job arguments into FenceExpireRequest")
		return err
	}

	glog.Info("Processing FenceExpireJob")

	fence, err, notFound := models.FindFenceByID(fer.FenceID)

	if notFound {
		return nil
	}

	if err != nil {
		glog.Errorf("Fence Finiding error: %v", err)
		return err
	}

	err = activity.AddFenceExpiredActivity(fence.User.ID, fence.ID, fence.Name)
	if err != nil {
		glog.Errorf("Activity creation error: %v", err)
	}

	QueueSendGcmRequest(gcm.NewMessage(map[string]interface{}{"type": "onFenceExpired", "fenceId": fence.ID, "fenceName": fence.Name}, fence.User.GCMNotificationID))

	err = search.DeleteGeoFence(int(fence.ID))

	if err != nil {
		return err
	}

	err = fence.Delete()

	if err != nil {
		return err
	}

	QueueNotifyUsersSyncRequest(fence.Lat, fence.Lon)

	return nil
}

// QueueFenceExpireRequest creates a new FenceExpiry Job for a fence.
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
