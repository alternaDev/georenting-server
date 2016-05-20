package jobs

import (
	"encoding/json"
	"log"

	"github.com/alternaDev/georenting-server/google/gcm"
	"github.com/bgentry/que-go"
)

const (
	SendGcmJobName = "SendGcm"
)

type SendGcmRequest struct {
	GCMNotificationID string
	Data              map[string]interface{}
}

func SendGcmJob(j *que.Job) error {
	var r SendGcmRequest
	err := json.Unmarshal(j.Args, &r)
	if err != nil {
		log.Fatal("Unable to unmarshal job arguments into SendGcmRequest")
		return err
	}

	log.Print("Processing SendGcmRequest")

	err = gcm.SendToGroup(gcm.NewMessage(r.Data, r.GCMNotificationID))

	return err
}

func QueueSendGcmRequest(r SendGcmRequest) error {
	enc, err := json.Marshal(r)
	if err != nil {
		return err
	}

	j := que.Job{
		Type: SendGcmJobName,
		Args: enc,
	}

	return QC.Enqueue(&j)
}
