package user

import "errors"

var (
	UserNotFound       = errors.New("user not found")
	InvalidCredentials = errors.New("invalid credentials")
)
