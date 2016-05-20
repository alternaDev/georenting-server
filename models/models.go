package models

import (
	"log"
	"os"

	"github.com/bgentry/que-go"
	"github.com/jackc/pgx"
	pgxstd "github.com/jackc/pgx/stdlib"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // For GORM
)

const (
	queTableSQL = `
		CREATE TABLE IF NOT EXISTS que_jobs
		(
			priority    smallint    NOT NULL DEFAULT 100,
			run_at      timestamptz NOT NULL DEFAULT now(),
			job_id      bigserial   NOT NULL,
			job_class   text        NOT NULL,
			args        json        NOT NULL DEFAULT '[]'::json,
			error_count integer     NOT NULL DEFAULT 0,
			last_error  text,
			queue       text        NOT NULL DEFAULT '',
			CONSTRAINT que_jobs_pkey PRIMARY KEY (queue, priority, run_at, job_id)
		);` // QueTableSQL to create table idempotently
)

// DB provides the ability to access the Database.
var (
	DBPool *pgx.ConnPool
	DB     gorm.DB
)

// prepQue ensures that the que table exists and que's prepared statements are
// run. It is meant to be used in a pgx.ConnPool's AfterConnect hook.
func prepQue(conn *pgx.Conn) error {
	_, err := conn.Exec(queTableSQL)
	if err != nil {
		return err
	}

	return que.PrepareStatements(conn)
}

func getPgxPool(dbURL string) (*pgx.ConnPool, error) {
	pgxcfg, err := pgx.ParseURI(dbURL)
	if err != nil {
		return nil, err
	}

	pgxpool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:   pgxcfg,
		AfterConnect: prepQue,
	})

	if err != nil {
		return nil, err
	}

	return pgxpool, nil
}

func init() {
	log.Println("Initializing Models.")

	pool, err := getPgxPool(os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	DBPool = pool

	db, err := initDB()
	if err != nil {
		panic(err)
	}
	DB = db

	defer DBPool.Close()
}

func initDB() (gorm.DB, error) {
	dbC, err := pgxstd.OpenFromConnPool(DBPool)
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open("postgres", dbC)

	if err != nil {
		log.Fatalf("Error while connecting to DB: %s", err)
		return db, err
	}

	db.AutoMigrate(&User{}, &Fence{}, &Score{})

	return db, nil
}
