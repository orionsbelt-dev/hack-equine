package rides

import (
	"database/sql"
	"errors"

	"hack/utils"
)

type Ride struct {
	ID      int64      `json:"id"`
	HorseID int64      `json:"horse_id"`
	RiderID int64      `json:"rider_id"`
	Date    utils.Date `json:"date"`
	Notes   string     `json:"notes"`
}

func (r *Ride) Save(db *sql.DB) error {
	query := "insert into rides (horse_id, rider_id, date, notes) values (?, ?, ?, ?)"
	result, err := db.Exec(query, r.HorseID, r.RiderID, r.Date.Format("2006-01-02"), r.Notes)
	if err != nil {
		return errors.New("failed to insert ride into database: " + err.Error())
	}
	r.ID, err = result.LastInsertId()
	if err != nil {
		return errors.New("failed to get last insert ID: " + err.Error())
	}
	return nil
}
