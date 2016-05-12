package main

import (
	"log"
	"net/http"
	"os"

	models "github.com/alternaDev/georenting-server/models"
	router "github.com/alternaDev/georenting-server/router"

	"github.com/gorilla/handlers"
)

func listen(port string) {
	r := router.SetupRouter()

	http.Handle("/", handlers.LoggingHandler(os.Stdout, r))
	http.ListenAndServe(":"+port, nil)
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	models.Init()
	//models.MigrateGeofencesToElasticSearch()

	log.Printf("Listening on port %s.", port)

	listen(port)
}
