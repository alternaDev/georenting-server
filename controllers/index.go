package controllers

import (
	"fmt"
	"net/http"

	"github.com/alternaDev/geomodel"
	"github.com/gorilla/mux"
)

// IndexHandler Handles the Index
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	fmt.Fprintf(w, "Hi there, I love %s! %f", vars["name"], geomodel.Distance(0, 0, 1, 1))
}
