package enum

type ContentType int8

const (
	Nan ContentType = iota
	LoginPassword
	BankCard
	File
)
