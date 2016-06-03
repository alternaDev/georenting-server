package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/alternaDev/georenting-server/activity"
	"github.com/alternaDev/georenting-server/auth"
	"github.com/alternaDev/georenting-server/google"
	"github.com/alternaDev/georenting-server/models"

	nameGen "github.com/alternaDev/go-random-name-gen"
)

type authBody struct {
	GoogleToken string `json:"google_token"`
}

type authResponseBody struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

type refreshTokenBody struct {
	Token string `json:"token"`
}

type cashResponseBody struct {
	EarningsRentSevenDays     float64 `json:"earnings_rent_7d"`
	ExpensesRentSevenDays     float64 `json:"expenses_rent_7d"`
	ExpensesGeoFenceSevenDays float64 `json:"expenses_geofence_7d"`
	EarningsRentAll           float64 `json:"earnings_rent_all"`
	ExpensesRentAll           float64 `json:"expenses_rent_all"`
	ExpensesGeoFenceAll       float64 `json:"expenses_geofence_all"`
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

	if user.Name == "" {
		var id int64
		id, err = strconv.ParseInt(googleUser.GoogleID[:7], 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		name := ""
		i := 0
		for name == "" {
			genName, err2 := nameGen.GenerateNameWithSeed(1, 1, 3, id+int64(i))
			if err2 != nil {
				http.Error(w, err2.Error(), http.StatusForbidden)
				return
			}
			count := 0
			models.DB.Where(models.User{Name: genName}).Count(&count)
			if count == 0 {
				name = genName
			}
			i = i + 1
		}

		user.Name = name
	}

	models.DB.Save(&user)

	token, err := auth.GenerateJWTToken(user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	user.AvatarURL = os.Getenv("BASE_URL") + "/users/" + user.Name + "/avatar"

	bytes, err := json.Marshal(authResponseBody{Token: token, User: user})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}

// RefreshTokenHandler handles POST /users/refreshToken
func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var b refreshTokenBody
	err := decoder.Decode(&b)

	if err != nil {
		http.Error(w, "Invalid Body.", http.StatusBadRequest)
		return
	}

	user, err := auth.ValidateJWTToken(b.Token)

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	token, err := auth.GenerateJWTToken(user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	err = auth.InvalidateToken(b.Token)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user.AvatarURL = os.Getenv("BASE_URL") + "/users/" + user.Name + "/avatar"

	bytes, err := json.Marshal(authResponseBody{Token: token, User: user})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}

// LogoutHandler DELETE /user/auth
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")

	err := auth.InvalidateToken(token)

	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	fmt.Fprintf(w, "{}")
}

// HistoryHandler GET /users/me/history
func HistoryHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.ValidateSession(r)

	if err != nil {
		http.Error(w, "Invalid Session token. "+err.Error(), http.StatusForbidden)
		return
	}

	from, err := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid From Value. "+err.Error(), http.StatusBadRequest)
		return
	}

	to, err := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid To Value. "+err.Error(), http.StatusBadRequest)
		return
	}

	data, err := activity.GetActivities(user.ID, to, from)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var buffer bytes.Buffer

	buffer.WriteString("[")

	for i := 0; i < len(data); i++ {
		buffer.WriteString(data[i])
		buffer.WriteString(",")
	}

	if len(data) > 0 {
		buffer.Truncate(buffer.Len() - 1)
	}

	buffer.WriteString("]")

	w.Write(buffer.Bytes())
}
