package utils

import (
	"encoding/json"
	"errors"
	"fmt"
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

type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return errors.New("failed to unmarshal time: " + err.Error())
	}
	fmt.Println("provided time as string: " + s)
	if len(s) > 0 {
		parsedTime, err := time.Parse("3:04 PM", s)
		if err != nil {
			return errors.New("failed to parse time: " + err.Error())
		}
		*t = Time{parsedTime}
	}
	return nil
}
