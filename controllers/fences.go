package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/alternaDev/geomodel"
	"github.com/alternaDev/georenting-server/activity"
	"github.com/alternaDev/georenting-server/auth"
	"github.com/alternaDev/georenting-server/google/gcm"
	"github.com/alternaDev/georenting-server/jobs"
	"github.com/alternaDev/georenting-server/maths"
	"github.com/alternaDev/georenting-server/models"
	"github.com/alternaDev/georenting-server/models/redis"
	"github.com/alternaDev/georenting-server/models/search"
	"github.com/alternaDev/georenting-server/scores"
	"github.com/alternaDev/georenting-server/util"
	"github.com/gorilla/mux"
)

type fenceResponse struct {
	ID             int            `json:"id"`
	Lat            float64        `json:"center_lat"`
	Lon            float64        `json:"center_lon"`
	Radius         int            `json:"radius"`
	Name           string         `json:"name"`
	Owner          int            `json:"owner"`
	TTL            int            `json:"ttl"`
	RentMultiplier float64        `json:"rent_multiplier"`
	DiesAt         util.Timestamp `json:"dies_at"`
	TotalVisitors  uint           `json:"total_visitors"`
	TotalEarnings  float64        `json:"total_earnings"`
	Cost           float64        `json:"cost"`
	OwnerName      string         `json:"owner_name"`
}

type costEstimateResponse struct {
	Cost      float64 `json:"cost"`
	CanAfford bool    `json:"can_afford"`
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

	fence, err, notFound := models.FindFenceByID(fenceID)

	if err != nil || notFound {
		http.Error(w, "Fence not Found", http.StatusNotFound)
		return
	}

	rent := scores.GetGeoFenceRent(fence)

	if user.Balance < rent {
		rent = user.Balance
	}

	// GCM
	err = jobs.QueueSendGcmRequest(gcm.NewMessage(
		map[string]interface{}{"type": "onForeignFenceEntered", "fenceId": fence.ID, "fenceName": fence.Name, "ownerName": fence.User.Name}, user.GCMNotificationID))

	if err != nil {
		InternalServerError(err, w)
		return
	}

	err = jobs.QueueSendGcmRequest(gcm.NewMessage(
		map[string]interface{}{"type": "onOwnFenceEntered", "fenceId": fence.ID, "fenceName": fence.Name, "visitorName": user.Name}, fence.User.GCMNotificationID))

	if err != nil {
		InternalServerError(err, w)
		return
	}

	// Activity Stream
	err = activity.AddForeignVisitedActivity(user.ID, fence.User.Name, fence.User.ID, fence.Name, fence.ID, rent)

	if err != nil {
		InternalServerError(err, w)
		return
	}

	err = activity.AddOwnFenceVisitedActivity(fence.User.ID, user.Name, user.ID, fence.Name, fence.ID, rent)

	if err != nil {
		InternalServerError(err, w)
		return
	}

	// Balance Recording
	redis.AddBalanceRecord(redis.GetBalanceRecordName(user.ID, redis.BalanceNameExpenseRent), rent)
	redis.AddBalanceRecord(redis.GetBalanceRecordName(fence.User.ID, redis.BalanceNameEarningsRent), rent)

	// Score Calculation
	err = jobs.QueueRecordVisitRequest(fence.Lat, fence.Lon, time.Now()) //scores.RecordVisit(fence.Lat, fence.Lon)

	if err != nil {
		InternalServerError(err, w)
		return
	}

	user.LastKnownGeoHash = geomodel.GeoCell(fence.Lat, fence.Lon, models.LastKnownGeoHashResolution)
	user.Balance = user.Balance - rent

	if user.Balance < 0 { // TODO: Decide what to do here.
		user.Balance = 0
	}

	user.ExpensesRentAllTime = user.ExpensesRentAllTime + rent

	err = user.Save()
	if err != nil {
		InternalServerError(err, w)
		return
	}

	fence.User.Balance = fence.User.Balance + rent
	fence.User.EarningsRentAllTime = user.ExpensesRentAllTime + rent

	err = fence.User.Save()

	if err != nil {
		InternalServerError(err, w)
		return
	}

	fence.TotalEarnings = fence.TotalEarnings + rent
	fence.TotalVisitors = fence.TotalVisitors + 1

	err = fence.Save()

	if err != nil {
		InternalServerError(err, w)
		return
	}

	w.Write([]byte("{}"))
}

// GetFencesHandler GET /fences
func GetFencesHandler(w http.ResponseWriter, r *http.Request) {
	lat, err1 := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
	lon, err2 := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)
	radius, err3 := strconv.ParseInt(r.URL.Query().Get("radius"), 10, 64)
	userID, err4 := strconv.ParseInt(r.URL.Query().Get("user"), 10, 8)
	excludeOwn, _ := strconv.ParseBool(r.URL.Query().Get("excludeOwn"))

	if err1 == nil && err2 == nil && err3 == nil {
		user, err := auth.ValidateSession(r)

		var result []models.Fence

		if err == nil && excludeOwn {
			result, err = search.FindGeoFencesExceptByUser(lat, lon, radius, user.ID)
		} else {
			result, err = search.FindGeoFences(lat, lon, radius)
		}

		if err != nil {
			InternalServerError(err, w)
			return
		}

		fences := make([]fenceResponse, len(result))
		for i := range result {
			f := result[i]
			fences[i].ID = f.ID
			fences[i].Lat = f.Lat
			fences[i].Lon = f.Lon
			fences[i].Name = f.Name
			fences[i].Radius = f.Radius
			fences[i].Owner = f.UserID
			fences[i].DiesAt = util.Timestamp(f.DiesAt)
			fences[i].RentMultiplier = f.RentMultiplier
			fences[i].OwnerName = f.User.Name
			if user != nil && f.UserID == user.ID {
				fences[i].Cost = f.Cost
				fences[i].TotalEarnings = f.TotalEarnings
				fences[i].TotalVisitors = f.TotalVisitors
			}
		}

		bytes, err := json.Marshal(&fences)

		if err != nil {
			InternalServerError(err, w)
			return
		}

		if user != nil {
			user.LastKnownGeoHash = geomodel.GeoCell(lat, lon, models.LastKnownGeoHashResolution)
			err = user.Save()
			if err != nil {
				InternalServerError(err, w)
				return
			}
		}

		w.Write(bytes)
		return
	}

	if err4 == nil {
		user, _ := auth.ValidateSession(r)

		fenceUser, errA := models.FindUserByID(int(userID))

		if errA != nil {
			InternalServerError(errA, w)
			return
		}

		result, errA := fenceUser.GetFences()
		if errA != nil {
			InternalServerError(errA, w)
			return
		}

		fences := make([]fenceResponse, len(result))
		for i := range result {
			f := result[i]
			fences[i].ID = f.ID
			fences[i].Lat = f.Lat
			fences[i].Lon = f.Lon
			fences[i].Name = f.Name
			fences[i].Radius = f.Radius
			fences[i].Owner = f.UserID
			fences[i].DiesAt = util.Timestamp(f.DiesAt)
			fences[i].RentMultiplier = f.RentMultiplier
			fences[i].OwnerName = fenceUser.Name
			if user != nil && f.UserID == user.ID {
				fences[i].Cost = f.Cost
				fences[i].TotalEarnings = f.TotalEarnings
				fences[i].TotalVisitors = f.TotalVisitors
			}
		}

		bytes, err := json.Marshal(&fences)

		if err != nil {
			InternalServerError(err, w)
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

	InternalServerError(err, w)
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
		InternalServerError(err, w)
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
	f.DiesAt = time.Now().Add(time.Duration(f.TTL) * time.Second)
	f.UserID = user.ID

	overlap, err := checkFenceOverlap(&f)

	if err != nil {
		InternalServerError(err, w)
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

	f.Cost = price

	user.LastKnownGeoHash = geomodel.GeoCell(requestFence.Lat, requestFence.Lon, models.LastKnownGeoHashResolution)
	user.Balance = user.Balance - price
	user.ExpensesGeoFenceAllTime = user.ExpensesGeoFenceAllTime + price

	err = user.Save()
	if err != nil {
		log.Errorf("Error while saving user: %v", err)
	}

	redis.AddBalanceRecord(redis.GetBalanceRecordName(user.ID, redis.BalanceNameExpenseGeoFence), price)

	f.User = *user

	err = f.Save()

	if err != nil {
		InternalServerError(err, w)
		return
	}

	err = search.IndexGeoFence(f)

	if err != nil {
		InternalServerError(err, w)
		return
	}

	jobs.QueueNotifyUsersSyncRequest(f.Lat, f.Lon)
	jobs.QueueFenceExpireRequest(&f)

	bytes, err := json.Marshal(&f)

	if err != nil {
		InternalServerError(err, w)
		return
	}

	w.Write(bytes)
}

func checkFenceOverlapWithFenceResponse(f *fenceResponse) (bool, error) {
	return checkFenceOverlap(&models.Fence{Lat: f.Lat, Lon: f.Lon, Radius: f.Radius})
}

func checkFenceOverlap(fence *models.Fence) (bool, error) {
	result, err := search.FindGeoFences(fence.Lat, fence.Lon, int64(fence.Radius+models.FenceMaxRadius))

	if err != nil {
		return false, err
	}

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
	user, _ := auth.ValidateSession(r)

	vars := mux.Vars(r)

	fenceID, err := strconv.ParseInt(vars["fenceId"], 10, 8)
	if err != nil {
		http.Error(w, "Invalid Fence ID. "+err.Error(), http.StatusUnauthorized)
		return
	}

	fence, err, notFound := models.FindFenceByID(fenceID)

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
	f.DiesAt = util.Timestamp(fence.DiesAt)
	f.RentMultiplier = fence.RentMultiplier
	f.OwnerName = fence.User.Name
	if user != nil && fence.UserID == user.ID {
		f.Cost = fence.Cost
		f.TotalEarnings = fence.TotalEarnings
		f.TotalVisitors = fence.TotalVisitors
	}

	bytes, err := json.Marshal(&f)

	if err != nil {
		InternalServerError(err, w)
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

	fence, err, notFound := models.FindFenceByID(fenceID)

	if notFound {
		http.Error(w, "GeoFence Not Found.", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if fence.UserID != user.ID {
		http.Error(w, "Unauthorized User.", http.StatusUnauthorized)
		return
	}

	err = search.DeleteGeoFence(int(fence.ID))

	if err != nil {
		log.Errorf("Error while deleting Fence from Search: %s", err.Error())
	}

	err = fence.Delete()

	if err != nil {
		InternalServerError(err, w)
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
		InternalServerError(err, w)
		return
	}

	overlap, err := checkFenceOverlapWithFenceResponse(&f)

	if err != nil {
		InternalServerError(err, w)
		return
	}

	if overlap {
		http.Error(w, "Fence does overlap.", http.StatusBadRequest)
		return
	}

	var response = costEstimateResponse{Cost: price, CanAfford: user.Balance >= price}

	bytes, err := json.Marshal(&response)

	if err != nil {
		InternalServerError(err, w)
		return
	}

	w.Write(bytes)
}
