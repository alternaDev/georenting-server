package jobs

import (
	"encoding/json"
	"log"

	"github.com/alternaDev/georenting-server/scores"
	"github.com/bgentry/que-go"
)

const (
	RecordVisitJobName = "RecordVisit"
)

type RecordVisitRequest struct {
	Lat float64
	Lon float64
}

func RecordVisitJob(j *que.Job) error {
	var r RecordVisitRequest
	err := json.Unmarshal(j.Args, &r)
	if err != nil {
		log.Fatal("Unable to unmarshal job arguments into RecordVisitRequest")
		return err
	}

	log.Print("Processing RecordVisitRequest")

	err = scores.RecordVisit(r.Lat, r.Lon)

	if err != nil {
		log.Printf("Could not calculate new Score: %s", err)
	}

	return err
}

func QueueRecordVisitRequest(lat float64, lon float64) error {
	r := RecordVisitRequest{Lat: lat, Lon: lon}
	enc, err := json.Marshal(r)
	if err != nil {
		return err
	}

	j := que.Job{
		Type: RecordVisitJobName,
		Args: enc,
	}

	return QC.Enqueue(&j)
}
