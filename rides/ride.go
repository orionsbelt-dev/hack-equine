package rides

import (
	"database/sql"
	"errors"
	"reflect"
	"sort"
	"time"

	"hack/utils"
)

type Ride struct {
	ID          int64       `json:"id,omitempty"`
	HorseID     int64       `json:"horse_id"`
	RiderID     int64       `json:"rider_id"`
	EventTypeID int64       `json:"event_type_id"`
	Date        utils.Date  `json:"date"`
	Time        *utils.Time `json:"time,omitempty"`
	Notes       string      `json:"notes"`
	Status      Status      `json:"status"`
}

type Status string

const (
	Scheduled Status = "scheduled"
	Cancelled Status = "cancelled"
	Completed Status = "completed"
)

func (r *Ride) Save(db *sql.DB) error {
	// set default status
	if r.Status == "" {
		r.Status = Scheduled
	}

	if r.EventTypeID == 0 {
		r.EventTypeID = 10 // Generic event type
	}

	if r.ID == 0 {
		query := "insert into rides (horse_id, rider_id, event_type_id, date, time, notes, status) values (?, ?, ?, ?, ?, ?, ?)"
		result, err := db.Exec(query, r.HorseID, r.RiderID, r.EventTypeID, r.Date.Format("2006-01-02"), r.Time, r.Notes, r.Status)
		if err != nil {
			return errors.New("failed to insert ride into database: " + err.Error())
		}
		r.ID, err = result.LastInsertId()
		if err != nil {
			return errors.New("failed to get last insert ID: " + err.Error())
		}
	} else {
		query := "update rides set horse_id = ?, rider_id = ?, event_type_id = ?, date = ?, time = ?, notes = ?, status = ? where id = ?"
		_, err := db.Exec(query, r.HorseID, r.RiderID, r.EventTypeID, r.Date.Format("2006-01-02"), r.Time, r.Notes, r.Status, r.ID)
		if err != nil {
			return errors.New("failed to update ride in database: " + err.Error())
		}
	}
	return nil
}

func (r *Ride) Cancel(db *sql.DB) error {
	r.Status = Cancelled
	if r.ID > 0 {
		query := "update rides set status = ? where id = ?"
		_, err := db.Exec(query, r.Status, r.ID)
		if err != nil {
			return errors.New("failed to set ride to cancelled status: " + err.Error())
		}
		return nil
	}
	query := "insert into rides (horse_id, rider_id, event_type_id, date, time, status) values (?, ?, ?, ?, ?, ?)"
	result, err := db.Exec(query, r.HorseID, r.RiderID, r.EventTypeID, r.Date.Format("2006-01-02"), r.Time, r.Status)
	if err != nil {
		return errors.New("failed to insert cancelled ride into database: " + err.Error())
	}
	r.ID, err = result.LastInsertId()
	if err != nil {
		return errors.New("failed to get last insert ID: " + err.Error())
	}
	return nil
}

type Schedule struct {
	ID        int64       `json:"id,omitempty"`
	HorseID   int64       `json:"horse_id"`
	HorseName string      `json:"horse_name,omitempty"`
	RiderID   int64       `json:"rider_id"`
	RiderName string      `json:"rider_name,omitempty"`
	EventType EventType   `json:"event_type,omitempty"`
	StartDate utils.Date  `json:"start_date"`
	EndDate   *utils.Date `json:"end_date,omitempty"`
	Time      *utils.Time `json:"time,omitempty"`
	Sunday    bool        `json:"sunday"`
	Monday    bool        `json:"monday"`
	Tuesday   bool        `json:"tuesday"`
	Wednesday bool        `json:"wednesday"`
	Thursday  bool        `json:"thursday"`
	Friday    bool        `json:"friday"`
	Saturday  bool        `json:"saturday"`
}

func (s *Schedule) Save(db *sql.DB) error {
	if s.ID == 0 {
		query := "insert into schedules (horse_id, rider_id, event_type_id, start_date, time, sunday, monday, tuesday, wednesday, thursday, friday, saturday) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		result, err := db.Exec(query, s.HorseID, s.RiderID, s.EventType.ID, s.StartDate.Format("2006-01-02"), s.Time, s.Sunday, s.Monday, s.Tuesday, s.Wednesday, s.Thursday, s.Friday, s.Saturday)
		if err != nil {
			return errors.New("failed to insert schedule into database: " + err.Error())
		}
		s.ID, err = result.LastInsertId()
		if err != nil {
			return errors.New("failed to get last insert ID: " + err.Error())
		}
	} else {
		query := "update schedules set horse_id = ?, rider_id = ?, event_type_id = ?, start_date = ?, time = ?, sunday = ?, monday = ?, tuesday = ?, wednesday = ?, thursday = ?, friday = ?, saturday = ? where id = ?"
		_, err := db.Exec(query, s.HorseID, s.RiderID, s.EventType.ID, s.StartDate.Format("2006-01-02"), s.Time, s.Sunday, s.Monday, s.Tuesday, s.Wednesday, s.Thursday, s.Friday, s.Saturday, s.ID)
		if err != nil {
			return errors.New("failed to update schedule in database: " + err.Error())
		}
	}
	// could improve this with custom value method for an end date type
	if s.EndDate != nil && s.EndDate.After(s.StartDate.Time) {
		query := "update schedules set end_date = ? where id = ?"
		_, err := db.Exec(query, s.EndDate.Format("2006-01-02"), s.ID)
		if err != nil {
			return errors.New("failed to update schedule end date in database: " + err.Error())
		}
	}
	return nil
}

func ListSchedules(barnID int64, db *sql.DB) ([]*Schedule, error) {
	query := "select id, horse_id, (select name from horses where id = horse_id) horse_name, rider_id, (select name from riders where id = rider_id) rider_name, event_type_id, (select name from event_types where id = event_type_id) event_type_name, start_date, end_date, time, sunday, monday, tuesday, wednesday, thursday, friday, saturday from schedules where horse_id in (select id from horses where barn_id = ?) and rider_id in (select id from riders where barn_id = ?) order by start_date, time"
	rows, err := db.Query(query, barnID, barnID)
	if err != nil {
		return nil, errors.New("failed to query schedules from database: " + err.Error())
	}
	defer rows.Close()

	var schedules []*Schedule
	for rows.Next() {
		var s Schedule
		err := rows.Scan(&s.ID, &s.HorseID, &s.HorseName, &s.RiderID, &s.RiderName, &s.EventType.ID, &s.EventType.Name, &s.StartDate, &s.EndDate, &s.Time, &s.Sunday, &s.Monday, &s.Tuesday, &s.Wednesday, &s.Thursday, &s.Friday, &s.Saturday)
		if err != nil {
			return nil, errors.New("failed to scan schedule from database: " + err.Error())
		}
		schedules = append(schedules, &s)
	}
	return schedules, nil
}

func DeleteSchedule(id int64, db *sql.DB) error {
	query := "delete from schedules where id = ?"
	_, err := db.Exec(query, id)
	if err != nil {
		return errors.New("failed to delete schedule from database: " + err.Error())
	}
	return nil
}

type RideDetail struct {
	Ride
	HorseName     string `json:"horse_name"`
	RiderName     string `json:"rider_name"`
	EventTypeName string `json:"event_type_name"`
}

func GetHorseScheduleByDay(horseID int64, date utils.Date, db *sql.DB) ([]*RideDetail, error) {
	query := "select id, horse_id, (select name from horses where id = horse_id) horse_name, rider_id, (select name from riders where id = rider_id) rider_name, event_type_id, (select name from event_types where id = event_type_id) event_type_name, date, time, status from rides where horse_id = ? and date = ? order by time"
	rows, err := db.Query(query, horseID, date.Format("2006-01-02"))
	if err != nil {
		return nil, errors.New("failed to query schedules from database: " + err.Error())
	}
	defer rows.Close()

	var rides []*RideDetail
	for rows.Next() {
		var r RideDetail
		err := rows.Scan(&r.ID, &r.HorseID, &r.HorseName, &r.RiderID, &r.RiderName, &r.EventTypeID, &r.EventTypeName, &r.Date, &r.Time, &r.Status)
		if err != nil {
			return nil, errors.New("failed to scan schedule from database: " + err.Error())
		}
		rides = append(rides, &r)
	}
	schedulesQuery := "select horse_id, (select name from horses where id = horse_id) horse_name, rider_id, (select name from riders where id = rider_id) rider_name, event_type_id, (select name from event_types where id = event_type_id) event_type_name, time, sunday, monday, tuesday, wednesday, thursday, friday, saturday from schedules where horse_id = ? and start_date <= ? and (end_date is null or end_date >= ?) order by start_date, time"
	scheduleRows, err := db.Query(schedulesQuery, horseID, date.Format("2006-01-02"), date.Format("2006-01-02"))
	if err != nil {
		return nil, errors.New("failed to query schedules from database: " + err.Error())
	}
	defer scheduleRows.Close()
	for scheduleRows.Next() {
		var s Schedule
		err := scheduleRows.Scan(&s.HorseID, &s.HorseName, &s.RiderID, &s.RiderName, &s.EventType.ID, &s.EventType.Name, &s.Time, &s.Sunday, &s.Monday, &s.Tuesday, &s.Wednesday, &s.Thursday, &s.Friday, &s.Saturday)
		if err != nil {
			return nil, errors.New("failed to scan schedule row: " + err.Error())
		}

		day := date.Weekday().String()
		scheduleValue := reflect.ValueOf(&s)
		field := scheduleValue.Elem().FieldByName(day)

		if field.Bool() {
			rides = appendScheduledRide(&s, date, rides)
		}
	}

	return rides, nil
}

func GetScheduleByDay(barnID int64, date utils.Date, db *sql.DB) ([]*RideDetail, error) {
	var rides []*RideDetail
	ridesQuery := "select id, horse_id, (select name from horses where id = horse_id) horse_name, rider_id, (select name from riders where id = rider_id) rider_name, event_type_id, (select name from event_types where id = event_type_id) event_type_name, time, notes, status from rides where date = ? and horse_id in (select id from horses where barn_id = ?) and rider_id in (select id from riders where barn_id = ?) order by time"
	mysqlDate := date.Format("2006-01-02")
	rideRows, err := db.Query(ridesQuery, mysqlDate, barnID, barnID)
	if err != nil {
		return nil, errors.New("failed to select rides from database: " + err.Error())
	}
	defer rideRows.Close()
	for rideRows.Next() {
		var r RideDetail
		var notes sql.NullString
		err := rideRows.Scan(&r.ID, &r.HorseID, &r.HorseName, &r.RiderID, &r.RiderName, &r.EventTypeID, &r.EventTypeName, &r.Time, &notes, &r.Status)
		if err != nil {
			return nil, errors.New("failed to scan ride row: " + err.Error())
		}
		if notes.Valid {
			r.Notes = notes.String
		}
		r.Date = date
		rides = append(rides, &r)
	}

	schedulesQuery := "select horse_id, (select name from horses where id = horse_id) horse_name, rider_id, (select name from riders where id = rider_id) rider_name, event_type_id, (select name from event_types where id = event_type_id) event_type_name, end_date, time, sunday, monday, tuesday, wednesday, thursday, friday, saturday from schedules where start_date <= ? and horse_id in (select id from horses where barn_id = ?) and rider_id in (select id from riders where barn_id = ?) order by time"
	rows, err := db.Query(schedulesQuery, mysqlDate, barnID, barnID)
	if err != nil {
		return nil, errors.New("failed to select schedules from database: " + err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var s Schedule
		var endDate *time.Time
		err := rows.Scan(&s.HorseID, &s.HorseName, &s.RiderID, &s.RiderName, &s.EventType.ID, &s.EventType.Name, &endDate, &s.Time, &s.Sunday, &s.Monday, &s.Tuesday, &s.Wednesday, &s.Thursday, &s.Friday, &s.Saturday)
		if err != nil {
			return nil, errors.New("failed to scan schedule row: " + err.Error())
		}
		if endDate != nil {
			s.EndDate = &utils.Date{Time: *endDate}
		}

		day := date.Weekday().String()
		scheduleValue := reflect.ValueOf(&s)
		field := scheduleValue.Elem().FieldByName(day)

		if s.EndDate != nil {
			if date.Before(s.EndDate.Time) && field.Bool() {
				rides = appendScheduledRide(&s, date, rides)
			}
		} else if field.Bool() {
			rides = appendScheduledRide(&s, date, rides)
		}
	}
	sortRides(rides)
	return rides, nil
}

func appendScheduledRide(s *Schedule, date utils.Date, rides []*RideDetail) []*RideDetail {
	var r RideDetail
	r.Date = date
	r.Status = Scheduled
	r.HorseID = s.HorseID
	r.HorseName = s.HorseName
	r.RiderID = s.RiderID
	r.RiderName = s.RiderName
	r.EventTypeID = s.EventType.ID
	r.EventTypeName = s.EventType.Name
	r.Time = s.Time
	found := areHorseAndRiderPresent(&r, rides)
	if !found {
		rides = append(rides, &r)
	}
	return rides
}

func areHorseAndRiderPresent(ride *RideDetail, rides []*RideDetail) bool {
	for _, r := range rides {
		if r.HorseID == ride.HorseID && r.RiderID == ride.RiderID && r.Time == ride.Time {
			return true
		}
	}
	return false
}

// I hate this function, but I don't know how to do it better right now
func sortRides(rides []*RideDetail) {
	sort.SliceStable(rides, func(i, j int) bool {
		var zeroTime time.Time
		if rides[i].Time == nil && rides[j].Time != nil {
			return zeroTime.After(rides[j].Time.Time)
		}
		if rides[j].Time == nil && rides[i].Time != nil {
			return rides[i].Time.Time.After(zeroTime)
		}
		if rides[i].Time == nil && rides[j].Time == nil {
			return true
		}
		return rides[i].Time.Time.Before(rides[j].Time.Time)
	})
}
