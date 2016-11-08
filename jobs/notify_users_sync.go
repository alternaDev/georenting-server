package jobs

import (
	"encoding/json"

	"github.com/alternaDev/geomodel"
	"github.com/alternaDev/georenting-server/google/gcm"
	"github.com/alternaDev/georenting-server/models"
	"github.com/bgentry/que-go"
	"github.com/golang/glog"
)

const (
	// NotifyUsersSyncJobName is the internal name of a NotifyUsersSync job.
	NotifyUsersSyncJobName = "NotifyUsersSync"
)

// NotifyUsersSyncRequest holds the data for a NotifyUsersSyncJob.
type NotifyUsersSyncRequest struct {
	GeoHash string
}

// NotifyUsersSyncJob executes a NotifyUsersSyncJob.
func NotifyUsersSyncJob(j *que.Job) error {
	var r NotifyUsersSyncRequest
	err := json.Unmarshal(j.Args, &r)
	if err != nil {
		glog.Errorf("Unable to unmarshal job arguments into NotifyUsersSyncRequest")
		return err
	}

	glog.Info("Processing NotifyUsersSyncJob")

	users, err := models.FindUsersByLastKnownGeoHash(r.GeoHash)
	if err != nil {
		glog.Errorf("Unable to Find Users")
		return err
	}

	for _, user := range users {
		QueueSendGcmRequest(gcm.NewMessage(map[string]interface{}{"type": "sync"}, user.GCMNotificationID))
	}

	return nil
}

// QueueNotifyUsersSyncRequest queues a new NotifyUsersSyncJob
func QueueNotifyUsersSyncRequest(lat float64, lon float64) error {
	r := NotifyUsersSyncRequest{GeoHash: geomodel.GeoCell(lat, lon, models.LastKnownGeoHashResolution)}
	enc, err := json.Marshal(r)
	if err != nil {
		return err
	}

	j := que.Job{
		Type: NotifyUsersSyncJobName,
		Args: enc,
	}

	return QC.Enqueue(&j)
}
