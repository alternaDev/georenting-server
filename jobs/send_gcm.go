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
	Message *gcm.Message
}

func SendGcmJob(j *que.Job) error {
	var r SendGcmRequest
	err := json.Unmarshal(j.Args, &r)
	if err != nil {
		log.Fatal("Unable to unmarshal job arguments into SendGcmRequest")
		return err
	}

	log.Print("Processing SendGcmRequest")

	err = gcm.SendToGroup(r.Message)

	if err != nil {
		log.Printf("Could not send GCM message: %s", err)
	}

	return err
}

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
