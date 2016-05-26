package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/alternaDev/geomodel"
	"github.com/alternaDev/georenting-server/activity"
	"github.com/alternaDev/georenting-server/auth"
	"github.com/alternaDev/georenting-server/google/gcm"
	"github.com/alternaDev/georenting-server/jobs"
	"github.com/alternaDev/georenting-server/maths"
	"github.com/alternaDev/georenting-server/models"
	"github.com/alternaDev/georenting-server/models/search"
	"github.com/alternaDev/georenting-server/scores"
	"github.com/gorilla/mux"
)

type fenceResponse struct {
	ID             uint      `json:"id"`
	Lat            float64   `json:"centerLat"`
	Lon            float64   `json:"centerLon"`
	Radius         int       `json:"radius"`
	Name           string    `json:"name"`
	Owner          uint      `json:"owner"`
	TTL            int       `json:"ttl"`
	RentMultiplier float64   `json:"rentMultiplier"`
	DiesAt         time.Time `json:"diesAt"`
}

type costEstimateResponse struct {
	Cost      float64 `json:"cost"`
	CanAfford bool    `json:"canAfford"`
}

// VisitFenceHandler handles POST /fences/{fenceId}/visit
func VisitFenceHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.ValidateSession(r)

	if err != nil {
		http.Error(w, "Invalid Session token. "+err.Error(), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)

	fenceID, err := strconv.ParseUint(vars["fenceId"], 10, 8)
	if err != nil {
		http.Error(w, "Invalid Fence ID. "+err.Error(), http.StatusUnauthorized)
		return
	}

	var fence models.Fence

	err = models.DB.Preload("User").Find(&fence, fenceID).Error

	if err != nil {
		http.Error(w, "Fence not Found", http.StatusNotFound)
		return
	}

	rent := scores.GetGeoFenceRent(&fence)

	// GCM
	err = jobs.QueueSendGcmRequest(gcm.NewMessage(
		map[string]interface{}{"type": "onForeignFenceEntered", "fenceId": fence.ID, "fenceName": fence.Name, "ownerName": fence.User.Name}, user.GCMNotificationID))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = jobs.QueueSendGcmRequest(gcm.NewMessage(
		map[string]interface{}{"type": "onOwnFenceEntered", "fenceId": fence.ID, "fenceName": fence.Name, "visitorName": user.Name}, fence.User.GCMNotificationID))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Activity Stream
	err = activity.AddForeignVisitedActivity(user.ID, fence.User.Name, fence.User.ID, fence.Name, fence.ID, rent)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = activity.AddOwnFenceVisitedActivity(fence.User.ID, user.Name, user.ID, fence.Name, fence.ID, rent)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Score Calculation
	err = jobs.QueueRecordVisitRequest(fence.Lat, fence.Lon, time.Now()) //scores.RecordVisit(fence.Lat, fence.Lon)

	if err != nil {
		log.Fatalf("Error while calulating score: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user.LastKnownGeoHash = geomodel.GeoCell(fence.Lat, fence.Lon, models.LastKnownGeoHashResolution)
	user.Balance = user.Balance - rent

	if user.Balance < 0 { // TODO: Decide what to do here.
		user.Balance = 0
	}

	models.DB.Save(&user)

	fence.User.Balance = fence.User.Balance + rent
	models.DB.Save(&fence.User)

	w.Write([]byte("{}"))
}

// GetFencesHandler GET /fences
func GetFencesHandler(w http.ResponseWriter, r *http.Request) {
	lat, err1 := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
	lon, err2 := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)
	radius, err3 := strconv.ParseInt(r.URL.Query().Get("radius"), 10, 64)
	userID, err4 := strconv.ParseUint(r.URL.Query().Get("user"), 10, 8)

	if err1 == nil && err2 == nil && err3 == nil {
		ids, err := search.FindGeoFences(lat, lon, radius)
		if err != nil {
			log.Fatal(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result := make([]models.Fence, len(ids))
		models.DB.Where(ids).Find(&result)

		fences := make([]fenceResponse, len(result))
		for i := range result {
			f := result[i]
			fences[i].ID = f.ID
			fences[i].Lat = f.Lat
			fences[i].Lon = f.Lon
			fences[i].Name = f.Name
			fences[i].Radius = f.Radius
			fences[i].Owner = f.UserID
			fences[i].DiesAt = f.DiesAt
			fences[i].RentMultiplier = f.RentMultiplier
		}

		bytes, err := json.Marshal(&fences)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user, err := auth.ValidateSession(r)

		if err == nil {
			user.LastKnownGeoHash = geomodel.GeoCell(lat, lon, models.LastKnownGeoHashResolution)
			models.DB.Save(&user)
		}

		w.Write(bytes)
		return
	}

	if err4 == nil {
		var user models.User
		models.DB.Preload("Fences").First(&user, userID)
		result := user.Fences

		fences := make([]fenceResponse, len(result))
		for i := range result {
			f := result[i]
			fences[i].ID = f.ID
			fences[i].Lat = f.Lat
			fences[i].Lon = f.Lon
			fences[i].Name = f.Name
			fences[i].Radius = f.Radius
			fences[i].Owner = f.UserID
			fences[i].DiesAt = f.DiesAt
			fences[i].RentMultiplier = f.RentMultiplier
		}

		bytes, err := json.Marshal(&fences)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(bytes)
		return
	}

	err := err1
	if err == nil {
		err = err2
	}
	if err == nil {
		err = err3
	}
	if err == nil {
		err = err4
	}
	if err == nil {
		err = errors.New("Please specify valid query options.")
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// CreateFenceHandler POST /fences
func CreateFenceHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.ValidateSession(r)

	if err != nil {
		http.Error(w, "Invalid Session token. "+err.Error(), http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var requestFence fenceResponse
	err = decoder.Decode(&requestFence)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	indexRadius := sort.SearchInts(models.UpgradeTypesRadius[:], requestFence.Radius)
	if !(indexRadius < len(models.UpgradeTypesRadius) && models.UpgradeTypesRadius[indexRadius] == requestFence.Radius) {
		http.Error(w, "Invalid Radius", http.StatusExpectationFailed)
		return
	}

	indexRent := sort.SearchFloat64s(models.UpgradeTypesRent[:], float64(requestFence.RentMultiplier))
	if !(indexRent < len(models.UpgradeTypesRent) && models.UpgradeTypesRent[indexRent] == requestFence.RentMultiplier) {
		http.Error(w, "Invalid RentMultiplier", http.StatusExpectationFailed)
		return
	}

	if requestFence.TTL <= 0 || requestFence.TTL > models.FenceMaxTTL {
		http.Error(w, "Invalid TTL", http.StatusExpectationFailed)
		return
	}

	var f models.Fence

	f.Lat = requestFence.Lat
	f.Lon = requestFence.Lon
	f.Name = requestFence.Name
	f.Radius = requestFence.Radius
	f.RentMultiplier = requestFence.RentMultiplier
	f.TTL = requestFence.TTL
	f.DiesAt = time.Now().Add(time.Duration(f.TTL) * time.Second) // TODO: Add Destruction Task.

	f.Radius = models.FenceMinRadius

	overlap, err := checkFenceOverlap(&f)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if overlap {
		http.Error(w, "Fence does overlap.", http.StatusBadRequest)
		return
	}

	price, err := scores.GetGeoFencePrice(f.Lat, f.Lon, f.TTL, f.RentMultiplier, indexRadius)
	if price > user.Balance {
		http.Error(w, "You do not have enough money for this thing.", http.StatusPaymentRequired)
		return
	}

	user.LastKnownGeoHash = geomodel.GeoCell(requestFence.Lat, requestFence.Lon, models.LastKnownGeoHashResolution)
	user.Balance = user.Balance - price

	err = models.DB.Save(&user).Error
	if err != nil {
		log.Printf("Error while saving user: %v", err)
	}

	f.User = user

	err = models.DB.Save(&f).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = search.IndexGeoFence(&f)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jobs.QueueNotifyUsersSyncRequest(f.Lat, f.Lon)
	jobs.QueueFenceExpireRequest(&f)

	bytes, err := json.Marshal(&f)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}

func checkFenceOverlapWithFenceResponse(f *fenceResponse) (bool, error) {
	return checkFenceOverlap(&models.Fence{Lat: f.Lat, Lon: f.Lon, Radius: f.Radius})
}

func checkFenceOverlap(fence *models.Fence) (bool, error) {
	ids, err := search.FindGeoFences(fence.Lat, fence.Lon, int64(fence.Radius+models.FenceMaxRadius))

	if err != nil {
		return false, err
	}

	result := make([]models.Fence, len(ids))
	err = models.DB.Where(ids).Find(&result).Error

	if err != nil {
		return false, err
	}

	for i := range result {
		fenceB := result[i]
		distance := maths.Distance(fence.Lat, fence.Lon, fenceB.Lat, fenceB.Lon)
		if distance < float64(fence.Radius+fenceB.Radius) {
			return true, nil
		}
	}
	return false, nil
}

// GetFenceHandler GET /fences/{fenceId}
func GetFenceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	fenceID, err := strconv.ParseUint(vars["fenceId"], 10, 8)
	if err != nil {
		http.Error(w, "Invalid Fence ID. "+err.Error(), http.StatusUnauthorized)
		return
	}

	var fence models.Fence

	notFound := models.DB.Find(&fence, fenceID).RecordNotFound()

	if notFound {
		http.Error(w, "GeoFence Not Found.", http.StatusNotFound)
		return
	}

	var f fenceResponse
	f.ID = fence.ID
	f.Lat = fence.Lat
	f.Lon = fence.Lon
	f.Name = fence.Name
	f.Radius = fence.Radius
	f.Owner = fence.UserID
	f.DiesAt = fence.DiesAt
	f.RentMultiplier = fence.RentMultiplier

	bytes, err := json.Marshal(&f)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}

// RemoveFenceHandler DELETE /fences/{fenceId}
func RemoveFenceHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.ValidateSession(r)

	if err != nil {
		http.Error(w, "Invalid Session token. "+err.Error(), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)

	fenceID, err := strconv.ParseUint(vars["fenceId"], 10, 8)
	if err != nil {
		http.Error(w, "Invalid Fence ID. "+err.Error(), http.StatusBadRequest)
		return
	}

	var fence models.Fence

	notFound := models.DB.Find(&fence, fenceID).RecordNotFound()

	if notFound {
		http.Error(w, "GeoFence Not Found.", http.StatusNotFound)
		return
	}

	if fence.UserID != user.ID {
		http.Error(w, "Unauthorized User.", http.StatusUnauthorized)
		return
	}

	err = search.DeleteGeoFence(&fence)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = models.DB.Delete(fence).Error

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jobs.QueueNotifyUsersSyncRequest(fence.Lat, fence.Lon)

	fmt.Fprintf(w, "{}")
}

// EstimateFenceCostHandler POST /fences/estimateCost
func EstimateFenceCostHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.ValidateSession(r)

	if err != nil {
		http.Error(w, "Invalid Session token. "+err.Error(), http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var f fenceResponse
	err = decoder.Decode(&f)

	indexRadius := sort.SearchInts(models.UpgradeTypesRadius[:], f.Radius)
	if !(indexRadius < len(models.UpgradeTypesRadius) && models.UpgradeTypesRadius[indexRadius] == f.Radius) {
		http.Error(w, "Invalid Radius", http.StatusExpectationFailed)
		return
	}

	indexRent := sort.SearchFloat64s(models.UpgradeTypesRent[:], float64(f.RentMultiplier))
	if !(indexRent < len(models.UpgradeTypesRent) && models.UpgradeTypesRent[indexRent] == f.RentMultiplier) {
		http.Error(w, "Invalid RentMultiplier", http.StatusExpectationFailed)
		return
	}

	if f.TTL <= 0 || f.TTL > models.FenceMaxTTL {
		http.Error(w, "Invalid TTL", http.StatusExpectationFailed)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	price, err := scores.GetGeoFencePrice(f.Lat, f.Lon, f.TTL, f.RentMultiplier, indexRadius)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	overlap, err := checkFenceOverlapWithFenceResponse(&f)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if overlap {
		http.Error(w, "Fence does overlap.", http.StatusBadRequest)
		return
	}

	var response = costEstimateResponse{Cost: price, CanAfford: user.Balance >= price}

	bytes, err := json.Marshal(&response)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}
