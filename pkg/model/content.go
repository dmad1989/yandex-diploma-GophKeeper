package model

import "github.com/dmad1989/gophKeeper/pkg/model/enum"

type Content struct {
	ID     int32
	UserID int32
	Type   enum.ContentType
	Data   []byte
	Desc   []byte
	Meta   string
}
