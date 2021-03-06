package jobs

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/alternaDev/georenting-server/google/gcm"
	"github.com/bgentry/que-go"
)

const (
	// SendGcmJobName is the name for a SendGcmJob.
	SendGcmJobName = "SendGcm"
)

// SendGcmRequest holds the data for a SendGcmRequest.
type SendGcmRequest struct {
	Message *gcm.Message
}

// SendGcmJob executes a SendGcm Job.
func SendGcmJob(j *que.Job) error {
	var r SendGcmRequest
	err := json.Unmarshal(j.Args, &r)
	if err != nil {
		log.Errorf("Unable to unmarshal job arguments into SendGcmRequest")
		return err
	}

	log.Info("Processing SendGcmRequest")

	err = gcm.SendToGroup(r.Message)

	if err != nil {
		log.Errorf("Could not send GCM message: %s", err)
	}

	return err
}

// QueueSendGcmRequest queues a new SendGcm Job.
func QueueSendGcmRequest(m *gcm.Message) error {
	r := SendGcmRequest{Message: m}
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
