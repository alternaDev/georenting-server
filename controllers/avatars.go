package controllers

import (
	"net/http"

	"github.com/gorilla/mux"

	avatarGen "github.com/alternaDev/go-avatar-gen"
)

// AvatarHandler handles GET /users/{name}/avatar
func AvatarHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	avatar := avatarGen.GenerateAvatar(vars["name"], 64, 32) // => *image.RGBA
	err := avatarGen.WriteImageToHTTP(w, avatar)
	if err != nil {
		InternalServerError(err, w)
	}
}
