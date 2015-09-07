package router

import (
	c "github.com/alternaDev/georenting-server/controllers"
	"github.com/gorilla/mux"
)

// SetupRouter sets up the router.
func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/{name}", c.IndexHandler)
	r.HandleFunc("/users/auth", c.AuthHandler).Methods("POST")
	r.HandleFunc("/users/auth", c.LogoutHandler).Methods("DELETE")
	r.HandleFunc("/users/me/gcm", c.GCMAddFunc).Methods("POST")

	return r
}
