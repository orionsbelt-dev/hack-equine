package users

import (
	"errors"

	"github.com/stytchauth/stytch-go/v4/stytch"
	"github.com/stytchauth/stytch-go/v4/stytch/stytchapi"
)

type Session struct {
	Token string `json:"token"`
}

const sessionDurationMinutes = 7 * 24 * 60

func (s *Session) Validate(client *stytchapi.API) error {
	if s.Token == "" {
		return errors.New("session token is required")
	}

	params := &stytch.SessionsAuthenticateParams{
		SessionToken:           s.Token,
		SessionDurationMinutes: sessionDurationMinutes,
	}

	_, err := client.Sessions.Authenticate(params)
	if err != nil {
		return errors.New("error authenticating session: " + err.Error())
	}

	return nil
}
