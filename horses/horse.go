package horses

import (
	"database/sql"
	"errors"
	"time"

	"hack/utils"
)

type Horse struct {
	ID     int64      `json:"id"`
	Name   string     `json:"name"`
	DOB    utils.Date `json:"dob"`
	Gender gender     `json:"gender"`
	BarnID int64      `json:"barn_id"`
}

func (h *Horse) Save(db *sql.DB) error {
	query := "insert into horses (name, dob, gender, barn_id) values (?, ?, ?, ?)"
	result, err := db.Exec(query, h.Name, h.DOB.Time.Format("2006-01-02"), h.Gender, h.BarnID)
	if err != nil {
		return errors.New("failed to insert horse into database: " + err.Error())
	}
	h.ID, err = result.LastInsertId()
	if err != nil {
		return errors.New("failed to get last insert ID: " + err.Error())
	}
	return nil
}

type gender string

const (
	Mare     gender = "mare"
	Gelding  gender = "gelding"
	Stallion gender = "stallion"
)

func GetHorsesByBarnID(barnID int64, db *sql.DB) ([]*Horse, error) {
	query := "select id, name, dob, gender from horses where barn_id = ?"
	rows, err := db.Query(query, barnID)
	if err != nil {
		return nil, errors.New("failed to select horses from database: " + err.Error())
	}
	defer rows.Close()
	var horses []*Horse
	for rows.Next() {
		var h Horse
		var dob time.Time
		err := rows.Scan(&h.ID, &h.Name, &dob, &h.Gender)
		if err != nil {
			return nil, errors.New("failed to scan row: " + err.Error())
		}
		h.DOB = utils.Date{Time: dob}
		h.BarnID = barnID
		horses = append(horses, &h)
	}
	return horses, nil
}

func GetHorses(db *sql.DB) ([]*Horse, error) {
	query := "select id, name, dob, gender, barn_id from horses"
	rows, err := db.Query(query)
	if err != nil {
		return nil, errors.New("failed to select horses from database: " + err.Error())
	}
	defer rows.Close()
	var horses []*Horse
	for rows.Next() {
		var h Horse
		var dob time.Time
		err := rows.Scan(&h.ID, &h.Name, &dob, &h.Gender, &h.BarnID)
		if err != nil {
			return nil, errors.New("failed to scan row: " + err.Error())
		}
		h.DOB = utils.Date{Time: dob}
		horses = append(horses, &h)
	}
	return horses, nil
}
