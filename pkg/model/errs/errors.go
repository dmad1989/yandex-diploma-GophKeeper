package errs

import "errors"

var ErrUserAlreadyExist = errors.New("user already exist")
var ErrUserNotFound = errors.New("user not found")
var ErrResNotFound = errors.New("resource not found")
var ErrResTooBig = errors.New("resource is too big")

var ErrTokenNotFound = errors.New("unauthorized")
var ErrTokenInvalid = errors.New("token invalid")
var ErrTokenNoUser = errors.New("token without userID")
var ErrReadMD = errors.New("read request metadata")
