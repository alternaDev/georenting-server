package jobs

import (
	"encoding/json"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/alternaDev/georenting-server/scores"
	"github.com/bgentry/que-go"
)

const (
	// RecordVisitJobName is the name of a RecordVisit Job.
	RecordVisitJobName = "RecordVisit"
)

// RecordVisitRequest holds the data for a RecordVisit Job.
type RecordVisitRequest struct {
	Lat  float64
	Lon  float64
	Time int64
}

// RecordVisitJob executes a RecordVisit Job.
func RecordVisitJob(j *que.Job) error {
	var r RecordVisitRequest
	err := json.Unmarshal(j.Args, &r)
	if err != nil {
		log.Errorf("Unable to unmarshal job arguments into RecordVisitRequest")
		return err
	}

	log.Info("Processing RecordVisitRequest")

	err = scores.RecordVisit(r.Lat, r.Lon, r.Time)

	if err != nil {
		log.Errorf("Could not calculate new Score: %s", err)
	}

	return err
}

// QueueRecordVisitRequest queues a new RecordVisit Job.
func QueueRecordVisitRequest(lat float64, lon float64, date time.Time) error {
	r := RecordVisitRequest{Lat: lat, Lon: lon, Time: date.Unix()}
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
