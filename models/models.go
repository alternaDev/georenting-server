package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq" // Required for Postgres Databases
)

type pgStringSlice []string

func (p *pgStringSlice) Scan(src interface{}) error {
	srcString, ok := src.(string)
	if !ok {
		return errors.New("Could not convert to String.")
	}
	srcString = "[" + srcString[1:len(srcString)-1] + "]"
	return json.Unmarshal([]byte(srcString), &p)
}
func (p pgStringSlice) Value() (driver.Value, error) {
	res, err := json.Marshal(&p)
	resString := string(res)
	resString = "{" + resString[1:len(resString)-1] + "}"

	return res, err
}

// DB provides the ability to access the Database.
var DB, err = initDB(os.Getenv("DATABASE_URL"))

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

	db.AutoMigrate(&User{}, &Fence{})

	return db, nil
}
