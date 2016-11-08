package main

import (
	"flag"
	"net/http"
	"os"

	router "github.com/alternaDev/georenting-server/router"

	"github.com/golang/glog"
	"github.com/gorilla/handlers"
)

func listen(port string) {
	r := router.SetupRouter()

	http.Handle("/", handlers.LoggingHandler(os.Stdout, r))
	http.ListenAndServe(":"+port, nil)
}

func main() {
	flag.Parse() // Glog needs this
	port := os.Getenv("PORT")

	if port == "" {
		glog.Fatal("$PORT must be set")
	}

	glog.Infof("Listening on port %s.", port)

	listen(port)
}
