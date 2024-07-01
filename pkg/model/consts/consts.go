package consts

const (
	//CLI Args
	LoginPassword = "lp"
	File          = "fl"
	BankCard      = "bc"

	// Descs
	LoginPasswordDesc = "Login and Password"
	BankCardDesc      = "Bank Card"
	FileDesc          = ""
)

var UserCtxKey = &contextKey{"userID"}
var LoggerCtxKey = &contextKey{"logger"}

type contextKey struct {
	name string
}
