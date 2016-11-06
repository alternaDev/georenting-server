package models

import "time"

// Score is a score for a geohash
type Score struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	GeoHash   string `gorm:"unique_index;primary_key"`
	LastVisit int64
	Score     float64
}

func (s Score) Save() error {
	var count int64
	err := DB.Get(&count, "SELECT count(*) FROM scores WHERE geo_hash = ?", s.GeoHash)
	if err != nil {
		return err
	}
	if count == 0 {
		s.UpdatedAt = time.Now()
		s.CreatedAt = time.Now()
		_, err := DB.Exec(`INSERT INTO scores (
			created_at,
			updated_at,
			geo_hash,
			last_visit,
			score) VALUES (?, ?, ?, ?, ?)`,
			s.CreatedAt,
			s.UpdatedAt,
			s.GeoHash,
			s.LastVisit,
			s.Score)
		return err
	} else {
		s.UpdatedAt = time.Now()
		_, err := DB.Exec(`UPDATE scores SET
			created_at=?,
			updated_at=?,
			last_visit=?,
			score=? WHERE geo_hash = ?`,
			s.CreatedAt,
			s.UpdatedAt,
			s.GeoHash,
			s.LastVisit,
			s.Score,
			s.GeoHash)
		return err
	}
}

func FindScoreByGeoHashOrInit(geoHash string) (Score, error) {
	var score Score

	err := DB.Get(&score, "SELECT * FROM scores WHERE geo_hash = ? LIMIT 1", geoHash)

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
