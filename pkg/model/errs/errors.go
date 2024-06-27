package errs

import "errors"

var (
	ErrUserAlreadyExist = errors.New("user already exist")
	ErrUserNotFound     = errors.New("user not found")
	ErrResNotFound      = errors.New("resource not found")
	ErrResTooBig        = errors.New("resource is too big")

	ErrTokenNotFound = errors.New("unauthorized")
	ErrTokenInvalid  = errors.New("token invalid")
	ErrTokenNoUser   = errors.New("token without userID")
	ErrReadMD        = errors.New("read request metadata")

	ErrNoCtxUser     = errors.New("no userID in context")
	ErrNotIntCtxUser = errors.New("wrong type of userID in context")
)
