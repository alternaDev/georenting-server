package jobs

import (
	"encoding/json"
	"log"

	"github.com/alternaDev/geomodel"
	"github.com/alternaDev/georenting-server/google/gcm"
	"github.com/alternaDev/georenting-server/models"
	"github.com/bgentry/que-go"
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
		log.Fatal("Unable to unmarshal job arguments into NotifyUsersSyncRequest")
		return err
	}

	log.Print("Processing NotifyUsersSyncJob")

	var users []models.User
	models.DB.Where(&models.User{LastKnownGeoHash: r.GeoHash}).Find(&users)

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
