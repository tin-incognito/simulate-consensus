package utils

import (
	"github.com/segmentio/ksuid"
)

//GenerateHashV1 ...
func GenerateHashV1() string{
	id := ksuid.New()
	return "0x"+ id.String()
}

func GenerateKey() string{
	id := ksuid.New()
	return id.String()
}