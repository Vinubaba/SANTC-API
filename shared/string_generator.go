package shared

import (
	"strings"

	"github.com/Pallinder/go-randomdata"
	"github.com/satori/go.uuid"
	"log"
)

type StringGenerator struct {
}

func (n *StringGenerator) GenerateRandomName() string {
	return strings.ToLower(randomdata.SillyName())
}

func (n *StringGenerator) GenerateUuid() string {
	id, err := uuid.NewV4()
	if err != nil {
		log.Fatal(err)
	}
	return id.String()
}
