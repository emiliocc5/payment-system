package uidgen

import "github.com/google/uuid"

var _uuidGen = func() string {
	return uuid.New().String()
}

func NewUUID() string {
	return _uuidGen()
}

func UseUUID(fn func() string) {
	_uuidGen = fn
}
