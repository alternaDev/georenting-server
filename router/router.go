package router

import (
	c "github.com/alternaDev/georenting-server/controllers"
	"github.com/gorilla/mux"
)

// SetupRouter sets up the router.
func SetupRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/users/auth", c.AuthHandler).Methods("POST")
	r.HandleFunc("/users/auth", c.LogoutHandler).Methods("DELETE")
	r.HandleFunc("/users/refreshToken", c.RefreshTokenHandler).Methods("POST")
	r.HandleFunc("/users/me/gcm", c.GCMAddHandler).Methods("POST")
	r.HandleFunc("/users/me/gcm", c.GCMRemoveHandler).Methods("DELETE")
	r.HandleFunc("/users/me/history", c.HistoryHandler).Methods("GET")
	r.HandleFunc("/users/me/cash", c.CashStatusHandler).Methods("GET")
	r.HandleFunc("/users/{name}/avatar", c.AvatarHandler).Methods("GET")
	r.HandleFunc("/users/{name}/avatar&sz=250", c.AvatarHandler).Methods("GET") // TEMPORYRY WORKAROUND!!!

	r.HandleFunc("/fences", c.GetFencesHandler).Methods("GET")
	r.HandleFunc("/fences", c.CreateFenceHandler).Methods("POST")
	r.HandleFunc("/fences/{fenceId}", c.GetFenceHandler).Methods("GET")
	r.HandleFunc("/fences/{fenceId}", c.RemoveFenceHandler).Methods("DELETE")
	r.HandleFunc("/fences/{fenceId}/visit", c.VisitFenceHandler).Methods("POST")
	r.HandleFunc("/fences/estimateCost", c.EstimateFenceCostHandler).Methods("POST")

	r.HandleFunc("/application/upgrades", c.UpgradesHandler).Methods("GET")

	r.HandleFunc("/{name}", c.IndexHandler)

	//r.HandleFunc("/fences/{fenceId}", c.GetFenceHandler).Methods("GET")
	//r.HandleFunc("/fences/{fenceId}", c.UpdateFenceHandler).Methods("PUT")
	//r.HandleFunc("/fences/{fenceId}", c.DeleteFenceHandler).Methods("DELETE")

	return r
}
