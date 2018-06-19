package consumers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestConsumers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Consumers Suite")
}
