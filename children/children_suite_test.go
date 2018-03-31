package children_test

import (
	"testing"

	"github.com/Vinubaba/SANTC-API/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestChildren(t *testing.T) {
	RegisterFailHandler(Fail)
	shared.InitDb()
	defer shared.DeleteDb()
	RunSpecs(t, "Children Suite")
}
