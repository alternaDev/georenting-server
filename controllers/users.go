package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/alternaDev/georenting-server/auth"
	"github.com/alternaDev/georenting-server/google"
	"github.com/alternaDev/georenting-server/models"
)

type authBody struct {
	GoogleToken string `json:"google_token"`
}

type authResponseBody struct {
	Token string `json:"token"`
}

// AuthHandler handles POST /users/auth
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var b authBody
	err := decoder.Decode(&b)

	if err != nil {
		http.Error(w, "Invalid Body.", http.StatusBadRequest)
		return
	}

	googleUser, err := google.VerifyToken(b.GoogleToken)

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	var user models.User
	models.DB.Where(models.User{GoogleID: googleUser.GoogleID}).FirstOrInit(&user)

	token, err := auth.GenerateJWTToken(user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	bytes, err := json.Marshal(authResponseBody{Token: token})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}
