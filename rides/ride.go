package rides

import (
	"database/sql"
	"errors"
	"time"

	"hack/utils"
)

type Ride struct {
	ID      int64      `json:"id,omitempty"`
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
	HorseID   int64      `json:"horse_id"`
	RiderID   int64      `json:"rider_id"`
	StartDate utils.Date `json:"start_date"`
	EndDate   utils.Date `json:"end_date"`
	Sunday    bool       `json:"sunday"`
	Monday    bool       `json:"monday"`
	Tuesday   bool       `json:"tuesday"`
	Wednesday bool       `json:"wednesday"`
	Thursday  bool       `json:"thursday"`
	Friday    bool       `json:"friday"`
	Saturday  bool       `json:"saturday"`
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
	query := "insert into schedules (horse_id, rider_id, start_date, day) values (?, ?, ?, ?)"
	for _, day := range days {
		result, err := db.Exec(query, s.HorseID, s.RiderID, s.StartDate.Format("2006-01-02"), day)
		if err != nil {
			return errors.New("failed to insert schedule into database: " + err.Error())
		}
		id, err := result.LastInsertId()
		if err != nil {
			return errors.New("failed to get last insert ID: " + err.Error())
		}
		if s.EndDate.After(s.StartDate.Time) {
			query := "update schedules set end_date = ? where id = ?"
			_, err := db.Exec(query, s.EndDate.Format("2006-01-02"), id)
			if err != nil {
				return errors.New("failed to update schedule end date in database: " + err.Error())
			}
		}
	}
	return nil
}

func GetScheduleByDay(barnID int64, date utils.Date, db *sql.DB) ([]*Ride, error) {
	var rides []*Ride
	ridesQuery := "select horse_id, rider_id, status from rides where date = ? and horse_id in (select id from horses where barn_id = ?) and rider_id in (select id from riders where barn_id = ?)"
	mysqlDate := date.Format("2006-01-02")
	rideRows, err := db.Query(ridesQuery, mysqlDate, barnID, barnID)
	if err != nil {
		return nil, errors.New("failed to select rides from database: " + err.Error())
	}
	defer rideRows.Close()
	for rideRows.Next() {
		var r Ride
		err := rideRows.Scan(&r.HorseID, &r.RiderID, &r.Status)
		if err != nil {
			return nil, errors.New("failed to scan row: " + err.Error())
		}
		r.Date = date
		rides = append(rides, &r)
	}

	schedulesQuery := "select horse_id, rider_id, end_date from schedules where day = ? and start_date <= ? and horse_id in (select id from horses where barn_id = ?) and rider_id in (select id from riders where barn_id = ?)"
	rows, err := db.Query(schedulesQuery, int64(date.Weekday()), mysqlDate, barnID, barnID)
	if err != nil {
		return nil, errors.New("failed to select schedules from database: " + err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var r Ride
		var endDate sql.NullTime
		err := rows.Scan(&r.HorseID, &r.RiderID, &endDate)
		if err != nil {
			return nil, errors.New("failed to scan row: " + err.Error())
		}
		r.Date = date
		r.Status = Scheduled
		if endDate.Valid {
			end := utils.Date{Time: endDate.Time}
			if date.Before(end.Time) {
				found := areHorseAndRiderPresent(&r, rides)
				if !found {
					rides = append(rides, &r)
				}
			}
		} else {
			// if the same horse and rider ID pair is not already in the list of rides, add it
			found := areHorseAndRiderPresent(&r, rides)
			if !found {
				rides = append(rides, &r)
			}
		}
	}

	return rides, nil
}

func areHorseAndRiderPresent(ride *Ride, rides []*Ride) bool {
	for _, r := range rides {
		if r.HorseID == ride.HorseID && r.RiderID == ride.RiderID {
			return true
		}
	}
	return false
}
