package model

type Content struct {
	ID     int32
	UserID int32
	Type   int32
	Data   []byte
	Meta   string
}
