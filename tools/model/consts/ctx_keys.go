package consts

var UserCtxKey = &contextKey{"userID"}
var LoggerCtxKey = &contextKey{"logger"}

type contextKey struct {
	name string
}
