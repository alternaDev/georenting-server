package models

import (
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq" // Required for Postgres Databases
)

// DB provides the ability to access the Database.
var DB, err = initDB(os.Getenv("DATABASE_URL"))

func initDB(url string) (gorm.DB, error) {
	db, err := gorm.Open("postgres", url)

	if err != nil {
		log.Fatalf("Error while connecting to DB: %s", err)
		return db, err
	}

	db.AutoMigrate(&User{})

	return db, nil
}
