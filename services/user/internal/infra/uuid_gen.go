package infra

import "github.com/google/uuid"

type UUIDGen struct{}

func (UUIDGen) New() string {
	return uuid.NewString()
}

func NewUUIDGen() *UUIDGen {
	return &UUIDGen{}
}
