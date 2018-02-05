package shared

import (
	"strings"

	"github.com/Pallinder/go-randomdata"
	"github.com/satori/go.uuid"
)

type StringGenerator struct {
}

func (n *StringGenerator) GenerateRandomName() string {
	return strings.ToLower(randomdata.SillyName())
}

func (n *StringGenerator) GenerateUuid() string {
	id := uuid.NewV4()
	return id.String()
}
