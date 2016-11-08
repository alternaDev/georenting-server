package controllers

import (
	"net/http"
	"runtime/debug"

	"github.com/golang/glog"
)

func InternalServerError(err error, w http.ResponseWriter) {
	glog.Errorf("Internal Server Error occured: %s", err.Error())
	debug.PrintStack()
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
