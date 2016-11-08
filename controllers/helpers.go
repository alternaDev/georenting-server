package controllers

import (
	"net/http"
	"runtime/debug"

	log "github.com/Sirupsen/logrus"
)

func InternalServerError(err error, w http.ResponseWriter) {
	log.Errorf("Internal Server Error occured: %s", err.Error())
	debug.PrintStack()
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
