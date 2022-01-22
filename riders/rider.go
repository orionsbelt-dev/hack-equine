package riders

import (
	"database/sql"
	"errors"
)

type Rider struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	BarnID int    `json:"barn_id"`
}

func (r *Rider) Save(db *sql.DB) error {
	query := "insert into riders (name, barn_id) values (?, ?)"
	result, err := db.Exec(query, r.Name, r.BarnID)
	if err != nil {
		return errors.New("failed to insert rider into database: " + err.Error())
	}
	r.ID, err = result.LastInsertId()
	if err != nil {
		return errors.New("failed to get last insert ID: " + err.Error())
	}
	return nil
}

func GetRiders(db *sql.DB) ([]*Rider, error) {
	query := "select id, name, barn_id from riders"
	rows, err := db.Query(query)
	if err != nil {
		return nil, errors.New("failed to select riders from database: " + err.Error())
	}
	defer rows.Close()
	var riders []*Rider
	for rows.Next() {
		var r Rider
		err := rows.Scan(&r.ID, &r.Name, &r.BarnID)
		if err != nil {
			return nil, errors.New("failed to scan row: " + err.Error())
		}
		riders = append(riders, &r)
	}
	return riders, nil
}
