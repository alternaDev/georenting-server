package models

import "time"

// Score is a score for a geohash
type Score struct {
	ID        int       `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	GeoHash   string    `gorm:"unique_index;primary_key" db:"geo_hash"`
	LastVisit int64     `db:"last_visit"`
	Score     float64   `db:"score"`
}

func (s *Score) Save() error {
	var count int64
	err := DB.Get(&count, "SELECT count(*) FROM scores WHERE geo_hash = $1", s.GeoHash)
	if err != nil {
		return err
	}
	if count == 0 {
		s.UpdatedAt = time.Now()
		s.CreatedAt = time.Now()
		var id int
		err := DB.QueryRow(`INSERT INTO scores (
			created_at,
			updated_at,
			geo_hash,
			last_visit,
			score) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			s.CreatedAt,
			s.UpdatedAt,
			s.GeoHash,
			s.LastVisit,
			s.Score).Scan(id)
		s.ID = id
		return err
	} else {
		s.UpdatedAt = time.Now()
		_, err := DB.Exec(`UPDATE scores SET
			created_at=$1,
			updated_at=$2,
			last_visit=$3,
			score=$4 WHERE geo_hash = $5`,
			s.CreatedAt,
			s.UpdatedAt,
			s.LastVisit,
			s.Score,
			s.GeoHash)
		return err
	}
}

func FindScoreByGeoHashOrInit(geoHash string) (Score, error) {
	var score Score

	err := DB.Get(&score, "SELECT * FROM scores WHERE geo_hash = $1 LIMIT 1", geoHash)

	if err != nil {
		score = Score{GeoHash: geoHash}
		score.Save()
	}

	return score, err
}

func FindAllScores() ([]Score, error) {
	var scores []Score
	err := DB.Select(&scores, "SELECT * FROM scores")
	return scores, err
}

func CountScores() (int64, error) {
	var count int64
	err := DB.Get(&count, "SELECT count(*) FROM scores")
	return count, err
}
