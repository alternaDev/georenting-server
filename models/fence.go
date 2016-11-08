package models

import (
	"time"

	"github.com/alternaDev/georenting-server/maths"
	"github.com/lib/pq"
)

var (
	// UpgradeTypesRadius holds the possible Upgrade Types for Radius.
	UpgradeTypesRadius = [...]int{100, 150, 200, 250, 300, 350, 400}
	// UpgradeTypesRent holds the possible rent multipliers.
	UpgradeTypesRent = [...]float64{1, 1.5, 2, 2.5, 3, 3.5, 4}
	// FenceMaxTTL holds the maximum possible TTL of a fence.
	FenceMaxTTL = 60 * 60 * 24 * 7 // 7 days
	// FenceMinTTL holds the minimum possible TTL of a fence.
	FenceMinTTL = 60 * 60 * 1 // 1 hour
	// FenceMinRadius holds the minimum radius of a fence.
	FenceMinRadius = maths.Min(UpgradeTypesRadius[:])
	// FenceMaxRadius holds the maximum radius of a fence.
	FenceMaxRadius = maths.Max(UpgradeTypesRadius[:])
)

// Fence is a fence
type Fence struct {
	ID             int       `db:"id"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
	User           User      `json:"-"`
	UserID         int       `json:"owner_id" db:"user_id"`
	Lat            float64   `json:"center_lat" db:"lat"`
	Lon            float64   `json:"center_lon" db:"lon"`
	Radius         int       `json:"radius" db:"radius"`
	RentMultiplier float64   `json:"rent_multiplier" db:"rent_multiplier"`
	TTL            int       `json:"ttl" db:"ttl"`
	DiesAt         time.Time `json:"diesAt" db:"dies_at"`
	Name           string    `json:"name" db:"name"`
	TotalVisitors  uint      `json:"total_visitors" db:"total_visitors"`
	TotalEarnings  float64   `json:"total_earnings" db:"total_earnings"`
	Cost           float64   `json:"cost" db:"cost"`
}

func (f *Fence) Save() error {
	if f.ID <= 0 {
		f.UpdatedAt = time.Now()
		f.CreatedAt = time.Now()
		var id int
		err := DB.QueryRow(`INSERT INTO fences (
			created_at,
			updated_at,
			user_id,
			lat,
			lon,
			radius,
			rent_multiplier,
			ttl,
			dies_at,
			name,
			total_visitors,
			total_earnings,
			cost) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`,
			f.CreatedAt,
			f.UpdatedAt,
			f.UserID,
			f.Lat,
			f.Lon,
			f.Radius,
			f.RentMultiplier,
			f.TTL,
			f.DiesAt,
			f.Name,
			f.TotalVisitors,
			f.TotalEarnings,
			f.Cost).Scan(&id)

		f.ID = id
		return err
	} else {
		f.UpdatedAt = time.Now()
		_, err := DB.Exec(`UPDATE fences SET
			created_at=$1,
			updated_at=$2,
			user_id=$3,
			lat=$4,
			lon=$5,
			radius=$6,
			rent_multiplier=$7,
			ttl=$8,
			dies_at=$9,
			name=$10,
			total_visitors=$11,
			total_earnings=$12,
			cost=$13 WHERE id = $14`,
			f.CreatedAt,
			f.UpdatedAt,
			f.UserID,
			f.Lat,
			f.Lon,
			f.Radius,
			f.RentMultiplier,
			f.TTL,
			f.DiesAt,
			f.Name,
			f.TotalVisitors,
			f.TotalEarnings,
			f.Cost,
			f.ID)
		return err
	}
}

func (f *Fence) Delete() error {
	_, err := DB.Exec("DELETE FROM fences WHERE id = $1", f.ID)
	return err
}

func FindFencesByIDs(ids []int64) ([]Fence, error) {
	var result []Fence

	if len(ids) == 0 {
		return result, nil
	}

	rows, err := DB.Query("SELECT id, created_at, updated_at, user_id, lat, lon, radius, rent_multiplier, ttl, dies_at, name, total_visitors, total_earnings, cost FROM fences WHERE id = ANY($1);", pq.Array(ids))
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var fence Fence
		var user User

		err = rows.Scan(&fence.ID, &fence.CreatedAt, &fence.UpdatedAt, &fence.UserID,
			&fence.Lat, &fence.Lon, &fence.Radius, &fence.RentMultiplier, &fence.TTL,
			&fence.DiesAt, &fence.Name, &fence.TotalVisitors, &fence.TotalEarnings,
			&fence.Cost)
		if err != nil {
			return nil, err
		}

		user, err = FindUserByID(fence.UserID)
		if err != nil {
			return nil, err
		}

		fence.User = user

		result = append(result, fence)
	}

	return result, err
}

func FindFenceByID(id interface{}) (Fence, error, bool) {
	var fence Fence
	var user User

	err := DB.Get(&fence, "SELECT * FROM fences WHERE id = $1 LIMIT 1", id)
	if err != nil {
		return fence, err, true
	}

	user, err = FindUserByID(fence.UserID)
	if err != nil {
		return fence, err, true
	}

	fence.User = user
	return fence, err, false
}
