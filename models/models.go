package models

import (
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq" // Required for Postgres Databases
)

// DB provides the ability to access the Database.
var DB gorm.DB

func init() {
	log.Println("Initializing Models.")
	db, err := initDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	DB = db
}

func initDB(url string) (gorm.DB, error) {
	var dbType = os.Getenv("DATABASE_TYPE")

	if dbType == "" {
		dbType = "postgres"
	}

	db, err := gorm.Open(dbType, url)

	if err != nil {
		log.Fatalf("Error while connecting to DB: %s", err)
		return db, err
	}

	db.AutoMigrate(&User{}, &Fence{}, &Score{})

	return db, nil
}
