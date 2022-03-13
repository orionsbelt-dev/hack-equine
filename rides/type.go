package rides

import (
	"database/sql"
	"errors"
)

type EventType struct {
	ID   int64
	Name string
}

func (t *EventType) Save(db *sql.DB) error {
	query := "insert into event_types (name) values (?)"
	result, err := db.Exec(query, t.Name)
	if err != nil {
		return errors.New("failed to insert event type into database: " + err.Error())
	}
	id, err := result.LastInsertId()
	if err != nil {
		return errors.New("failed to get last insert id: " + err.Error())
	}
	t.ID = id
	return nil
}

func ListEventTypes(db *sql.DB) ([]EventType, error) {
	query := "select id, name from event_types"
	rows, err := db.Query(query)
	if err != nil {
		return nil, errors.New("failed to query event types: " + err.Error())
	}
	defer rows.Close()
	var types []EventType
	for rows.Next() {
		var t EventType
		err = rows.Scan(&t.ID, &t.Name)
		if err != nil {
			return nil, errors.New("failed to scan event type: " + err.Error())
		}
		types = append(types, t)
	}
	return types, nil
}
