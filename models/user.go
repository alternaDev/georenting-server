package models

import "time"

const (
	// LastKnownGeoHashResolution is the resolution for the geohash of the last known position.
	LastKnownGeoHashResolution = 5
)

// User is a user.
type User struct {
	ID                      uint      `gorm:"primary_key" db:"id"`
	CreatedAt               time.Time `db:"created_at"`
	UpdatedAt               time.Time `db:"updated_at"`
	GoogleID                string    `json:"-" gorm:"index" db:"google_id"`
	Fences                  []Fence   `json:"fences"`
	PrivateKey              string    `sql:"size:4096" json:"-" db:"private_key"`
	GCMNotificationID       string    `json:"-" db:"gcm_notification_id"`
	Name                    string    `json:"name" db:"name"`
	AvatarURL               string    `json:"avatar_url" gorm:"-"`
	Balance                 float64   `json:"balance" db:"balance"`
	LastKnownGeoHash        string    `json:"-" db:"last_known_geo_hash"`
	EarningsRentAllTime     float64   `json:"-" db:"earnings_rent_all_time"`
	ExpensesRentAllTime     float64   `json:"-" db:"expenses_rent_all_time"`
	ExpensesGeoFenceAllTime float64   `json:"-" db:"expenses_geo_fence_all_time"`
}

func (u User) Save() error {
	if u.ID <= 0 {
		u.UpdatedAt = time.Now()
		u.CreatedAt = time.Now()
		_, err := DB.Exec(`INSERT INTO users (
			created_at,
			updated_at,
			google_id,
			private_key,
			name,
			gcm_notification_id,
			balance,
			last_known_geo_hash,
			earnings_rent_all_time,
			expenses_rent_all_time,
			expenses_geo_fence_all_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			u.CreatedAt,
			u.UpdatedAt,
			u.GoogleID,
			u.PrivateKey,
			u.Name,
			u.GCMNotificationID,
			u.Balance,
			u.LastKnownGeoHash,
			u.EarningsRentAllTime,
			u.ExpensesRentAllTime,
			u.ExpensesGeoFenceAllTime)
		return err
	} else {
		u.UpdatedAt = time.Now()
		_, err := DB.Exec(`UPDATE users SET
			created_at=?,
			updated_at=?,
			google_id=?,
			private_key=?,
			name=?,
			gcm_notification_id=?,
			balance=?,
			last_known_geo_hash=?,
			earnings_rent_all_time=?,
			expenses_rent_all_time=?,
			expenses_geo_fence_all_time=? WHERE id = ?`,
			u.CreatedAt,
			u.UpdatedAt,
			u.GoogleID,
			u.PrivateKey,
			u.Name,
			u.GCMNotificationID,
			u.Balance,
			u.LastKnownGeoHash,
			u.EarningsRentAllTime,
			u.ExpensesRentAllTime,
			u.ExpensesGeoFenceAllTime,
			u.ID)
		return err
	}
}

func (u User) GetFences() ([]Fence, error) {
	fences := []Fence{}

	// this will pull places with telcode > 50 into the slice pp
	err := DB.Select(&fences, "SELECT * FROM fences WHERE user_id = $1;", u.ID)

	return fences, err
}

func FindUserByID(id uint) (User, error) {
	var result User
	err := DB.Get(&result, "SELECT * FROM users WHERE id = $1 LIMIT 1;", id)
	return result, err
}

func FindUsersByLastKnownGeoHash(hash string) ([]User, error) {
	var users []User
	err := DB.Select(&users, "SELECT * FROM users WHERE last_known_geo_hash = $1;", hash)
	return users, err
}

func FindUserByGoogleIDOrInit(id string) (User, error) {
	var user User

	err := DB.Get(&user, "SELECT * FROM users WHERE google_id = $1 LIMIT 1;", id)

	if err != nil {
		user = User{GoogleID: id}
		user.Save()
	}

	return user, err
}

func CountUsersByName(name string) (int64, error) {
	if name == "" {
		return 0, nil
	}
	var count int64
	err := DB.Get(&count, "SELECT count(*) FROM users WHERE name = $1;", name)
	return count, err
}
