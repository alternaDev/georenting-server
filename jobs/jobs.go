package jobs

import (
	"os"

	"github.com/alternaDev/georenting-server/models"
	"github.com/bgentry/que-go"
)

var (
	QC *que.Client
)

func init() {
	QC = setup(os.Getenv("DATABASE_URL"))
}

func setup(dbURL string) *que.Client {
	pgxpool := models.DBPool

	qc := que.NewClient(pgxpool)

	return qc
}
