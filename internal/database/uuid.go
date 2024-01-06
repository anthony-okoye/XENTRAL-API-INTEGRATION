package database

import (
	"github.com/google/uuid"
)

type UUIDGenerator struct{}

func (u UUIDGenerator) Generate() interface{} {
	return uuid.New().String()
}

func (u UUIDGenerator) String() string {
	return "UUIDGenerator"
}
