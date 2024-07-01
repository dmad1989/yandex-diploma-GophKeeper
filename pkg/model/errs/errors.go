package errs

import "errors"

var (
	ErrUserAlreadyExist = errors.New("user already exist")
	ErrUserNotFound     = errors.New("user not found")
	ErrContNotFound     = errors.New("content not found")
	ErrContTooBig       = errors.New("content is too big")

	ErrTokenNotFound = errors.New("unauthorized")
	ErrTokenInvalid  = errors.New("token invalid")
	ErrTokenNoUser   = errors.New("token without userID")
	ErrReadMD        = errors.New("read request metadata")

	ErrNoCtxUser     = errors.New("no userID in context")
	ErrNotIntCtxUser = errors.New("wrong type of userID in context")

	//Cli errors
	ErrInputLogin    = errors.New("empty username not supported")
	ErrInputPassword = errors.New("empty password not supported")
	ErrInputFilePath = errors.New("empty path not supported")
	ErrInputDesc     = errors.New("empty description not supported")

	ErrInputBCNumber     = errors.New("empty numbeer not supported")
	ErrInputBCExpireDate = errors.New("empty data not supported")
	ErrInputBCName       = errors.New("empty name not supported")
	ErrInputBCSurname    = errors.New("empty surname not supported")

	ErrEmptyArgID   = errors.New("arg '[id]' is empty, type 'help' to display available commands format")
	ErrEmptyArgType = errors.New("arg '[type]' is empty, type 'help' to display available commands format")

	ErrFileUpdate  = errors.New("file update is not implemented, create a new")
	ErrFileMaxSize = errors.New("file size is too big")
)
