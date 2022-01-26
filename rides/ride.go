package rides

import (
	"database/sql"
	"errors"
	"time"

	"hack/utils"
)

type Ride struct {
	ID      int64      `json:"id"`
	HorseID int64      `json:"horse_id"`
	RiderID int64      `json:"rider_id"`
	Date    utils.Date `json:"date"`
	Notes   string     `json:"notes"`
	Status  Status     `json:"status"`
}

type Status string

const (
	Scheduled Status = "scheduled"
	Cancelled Status = "cancelled"
	Completed Status = "completed"
)

func (r *Ride) Save(db *sql.DB) error {
	query := "insert into rides (horse_id, rider_id, date, notes, status) values (?, ?, ?, ?, ?)"
	result, err := db.Exec(query, r.HorseID, r.RiderID, r.Date.Format("2006-01-02"), r.Notes, r.Status)
	if err != nil {
		return errors.New("failed to insert ride into database: " + err.Error())
	}
	r.ID, err = result.LastInsertId()
	if err != nil {
		return errors.New("failed to get last insert ID: " + err.Error())
	}
	return nil
}

type Schedule struct {
	HorseID   int64 `json:"horse_id"`
	RiderID   int64 `json:"rider_id"`
	Sunday    bool  `json:"sunday"`
	Monday    bool  `json:"monday"`
	Tuesday   bool  `json:"tuesday"`
	Wednesday bool  `json:"wednesday"`
	Thursday  bool  `json:"thursday"`
	Friday    bool  `json:"friday"`
	Saturday  bool  `json:"saturday"`
}

func (s *Schedule) Save(db *sql.DB) error {
	var days []time.Weekday
	if s.Sunday {
		days = append(days, time.Sunday)
	}
	if s.Monday {
		days = append(days, time.Monday)
	}
	if s.Tuesday {
		days = append(days, time.Tuesday)
	}
	if s.Wednesday {
		days = append(days, time.Wednesday)
	}
	if s.Thursday {
		days = append(days, time.Thursday)
	}
	if s.Friday {
		days = append(days, time.Friday)
	}
	if s.Saturday {
		days = append(days, time.Saturday)
	}
	query := "insert into schedules (horse_id, rider_id, day) values (?, ?, ?)"
	for _, day := range days {
		_, err := db.Exec(query, s.HorseID, s.RiderID, day)
		if err != nil {
			return errors.New("failed to insert schedule into database: " + err.Error())
		}
	}
	return nil
}
