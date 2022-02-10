package utils

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Date struct {
	time.Time
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return errors.New("failed to unmarshal date: " + err.Error())
	}
	if len(s) > 0 {
		parsedTime, err := time.Parse("1/2/2006", s)
		if err != nil {
			return errors.New("failed to parse time: " + err.Error())
		}
		*d = Date{parsedTime}
	}
	return nil
}

func (d *Date) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	t := value.(time.Time)
	*d = Date{t}
	return nil
}

type Time struct {
	sql.NullTime
}

func (t *Time) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return errors.New("failed to unmarshal time: " + err.Error())
	}
	if len(s) > 0 {
		parsedTime, err := time.Parse("3:04 PM", s)
		if err != nil {
			return errors.New("failed to parse time: " + err.Error())
		}
		t.Valid = true
		t.Time = parsedTime
	} else {
		t.Valid = false
	}
	return nil
}

func (t *Time) MarshalJSON() ([]byte, error) {
	if t != nil {
		return json.Marshal(t.Time.Format("3:04 PM"))
	}
	return json.Marshal("")
}

func (t *Time) Scan(value interface{}) error {
	if value == nil {
		t.Valid = false
		return nil
	}
	b := value.([]byte)
	parsedTime, err := time.Parse("15:04:05", string(b))
	if err != nil {
		return errors.New("failed to parse time: " + err.Error())
	}
	t.Valid = true
	t.Time = parsedTime
	return nil
}

func (t *Time) Value() (driver.Value, error) {
	if !t.Valid {
		return nil, nil
	}
	return t.Time.Format("15:04:05"), nil
}
