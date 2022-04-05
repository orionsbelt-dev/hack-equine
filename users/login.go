package users

import (
	"database/sql"
	"errors"

	"github.com/stytchauth/stytch-go/v4/stytch"
	"github.com/stytchauth/stytch-go/v4/stytch/stytchapi"
)

type User struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	StytchUserID   string `json:"stytch_user_id"`
	SessionToken   string `json:"session_token"`
	stytchMethodID sql.NullString
}

func (u *User) Signup(client *stytchapi.API, db *sql.DB) error {
	params := &stytch.OTPsSMSLoginOrCreateParams{
		PhoneNumber: u.Phone,
	}
	resp, err := client.OTPs.SMS.LoginOrCreate(params)
	if err != nil {
		return errors.New("error sending SMS: " + err.Error())
	}
	u.StytchUserID = resp.UserID
	u.stytchMethodID.String = resp.PhoneID
	u.stytchMethodID.Valid = true

	// insert user into database
	result, err := db.Exec("INSERT INTO users (name, email, phone, stytch_user_id, stytch_method_id) VALUES (?, ?, ?, ?, ?)", u.Name, u.Email, u.Phone, u.StytchUserID, u.stytchMethodID.String)
	if err != nil {
		return errors.New("error inserting user into database: " + err.Error())
	}
	u.ID, err = result.LastInsertId()
	if err != nil {
		return errors.New("error getting last insert id: " + err.Error())
	}

	return nil
}

func (u *User) Login(client *stytchapi.API, db *sql.DB) error {
	params := &stytch.OTPsSMSLoginOrCreateParams{
		PhoneNumber: u.Phone,
	}
	resp, err := client.OTPs.SMS.LoginOrCreate(params)
	if err != nil {
		return errors.New("error sending SMS: " + err.Error())
	}
	u.StytchUserID = resp.UserID
	u.stytchMethodID.String = resp.PhoneID
	u.stytchMethodID.Valid = true

	err = db.QueryRow("select id, name, email from users where stytch_user_id = ?", u.StytchUserID).Scan(&u.ID, &u.Name, &u.Email)
	if err == sql.ErrNoRows {
		return errors.New("user not found")
	}
	if err != nil {
		return errors.New("error getting user from database: " + err.Error())
	}
	// update method id
	_, err = db.Exec("UPDATE users SET stytch_method_id = ? WHERE id = ?", u.stytchMethodID.String, u.ID)
	if err != nil {
		return errors.New("error updating user in database: " + err.Error())
	}

	return nil
}

type UserAuth struct {
	Phone    string `json:"phone"`
	Passcode string `json:"passcode"`
}

func (a *UserAuth) AuthenticatePasscode(client *stytchapi.API, db *sql.DB) (*User, error) {
	var u User
	u.Phone = a.Phone
	query := "select id, stytch_method_id from users where phone = ?"
	err := db.QueryRow(query, a.Phone).Scan(&u.ID, &u.stytchMethodID)
	if err != nil {
		return nil, errors.New("error getting user from database: " + err.Error())
	}
	if !u.stytchMethodID.Valid {
		return nil, errors.New("user has no Stytch method id")
	}
	params := &stytch.OTPsAuthenticateParams{
		MethodID:               u.stytchMethodID.String,
		Code:                   a.Passcode,
		SessionDurationMinutes: 7 * 24 * 60,
	}
	resp, err := client.OTPs.Authenticate(params)
	if err != nil {
		return nil, errors.New("error authenticating passcode: " + err.Error())
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("passcode authentication failed")
	}
	u.SessionToken = resp.SessionToken
	return &u, nil
}
