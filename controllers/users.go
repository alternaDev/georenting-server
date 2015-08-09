package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/alternaDev/georenting-server/google"
	"github.com/alternaDev/georenting-server/models"
)

type authBody struct {
	GoogleToken string `json:"google_token"`
}

// AuthHandler handles POST /users/auth
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var b authBody
	err := decoder.Decode(&b)

	if err != nil {
		http.Error(w, "Invalid Body.", http.StatusBadRequest)
	}

	googleUser, err := google.VerifyToken(b.GoogleToken)

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	var user models.User
	models.DB.FirstOrInit(&user, models.User{GoogleID: googleUser.GoogleID})

	// TODO: Create JWT and return it
}
