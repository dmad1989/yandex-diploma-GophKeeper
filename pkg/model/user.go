package model

type User struct {
	ID           int32
	Login        string
	Password     string
	HashPassword []byte
}
