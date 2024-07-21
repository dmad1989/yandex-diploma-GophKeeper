package enum

import "github.com/dmad1989/gophKeeper/pkg/model/consts"

type ContentType int8

const (
	Nan ContentType = iota
	LoginPassword
	BankCard
	File
)

var (
	ArgToType = map[string]ContentType{
		consts.BankCard:      BankCard,
		consts.LoginPassword: LoginPassword,
		consts.File:          File,
	}

	TypeToArg = map[ContentType]string{
		BankCard:      consts.BankCard,
		LoginPassword: consts.LoginPassword,
		File:          consts.File,
	}

	DescToType = map[string]ContentType{
		consts.BankCardDesc:      BankCard,
		consts.LoginPasswordDesc: LoginPassword,
		consts.FileDesc:          File,
	}

	TypeToDesc = map[ContentType]string{
		BankCard:      consts.BankCardDesc,
		LoginPassword: consts.LoginPasswordDesc,
		File:          consts.FileDesc,
	}
)
