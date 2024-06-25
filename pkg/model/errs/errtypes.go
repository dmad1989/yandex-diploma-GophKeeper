package errs

import "fmt"

type TokenError struct {
	Err error
}

func (tknErr TokenError) Error() string {
	return fmt.Sprintf("token error: %v", tknErr.Err)
}
