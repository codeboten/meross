package api

import "errors"

var (
	errNoLogin                 = errors.New("No login found")
	errMissingUser             = errors.New("Email missing in the request")
	errMissingPassword         = errors.New("Password missing in the request")
	errAuthenticationException = errors.New("Authentication failure")
	errSignCheckFailed         = errors.New("Sign check failed")
)
