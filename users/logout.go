package users

import (
	"errors"

	"github.com/stytchauth/stytch-go/v4/stytch"
	"github.com/stytchauth/stytch-go/v4/stytch/stytchapi"
)

func Logout(sessionToken string, client *stytchapi.API) error {
	params := stytch.SessionsRevokeParams{
		SessionToken: sessionToken,
	}
	_, err := client.Sessions.Revoke(&params)
	if err != nil {
		return errors.New("error revoking session: " + err.Error())
	}
	return nil
}
