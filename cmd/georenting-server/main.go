package main

import (
	"net/http"
	"os"

	router "github.com/alternaDev/georenting-server/router"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
)

func listen(port string) {
	r := router.SetupRouter()

	http.Handle("/", handlers.LoggingHandler(os.Stdout, r))
	http.ListenAndServe(":"+port, nil)
}

func init() {
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	log.Infof("Listening on port %s.", port)

	listen(port)
}
