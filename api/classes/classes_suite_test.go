package classes_test

import (
	"testing"

	"github.com/Vinubaba/SANTC-API/api/shared"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUsers(t *testing.T) {
	RegisterFailHandler(Fail)
	shared.InitDb()
	defer shared.DeleteDb()
	RunSpecs(t, "Classes Suite")
}
