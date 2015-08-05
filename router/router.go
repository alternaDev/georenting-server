package router

import (
	c "github.com/alternaDev/georenting-server/controllers"
	"github.com/gorilla/mux"
)

// SetupRouter sets up the router.
func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/{name}", c.IndexHandler)

	return r
}
