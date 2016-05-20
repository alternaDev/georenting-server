package jobs

import (
	"github.com/alternaDev/georenting-server/models"
	"github.com/bgentry/que-go"
	"github.com/jackc/pgx"
)

func Setup(dbURL string) (*pgx.ConnPool, *que.Client) {
	pgxpool := models.DBPool

	qc := que.NewClient(pgxpool)

	return pgxpool, qc
}
