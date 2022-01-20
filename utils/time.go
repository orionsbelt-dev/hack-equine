package utils

import (
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
	parsedTime, err := time.Parse("1/2/2006", s)
	if err != nil {
		return errors.New("failed to parse time: " + err.Error())
	}
	*d = Date{parsedTime}
	return nil
}
