package storage_test

import (
	"context"
	b64 "encoding/base64"
	"io/ioutil"
	"os"

	"github.com/Vinubaba/SANTC-API/api/shared"
	. "github.com/Vinubaba/SANTC-API/api/shared/mocks"
	. "github.com/Vinubaba/SANTC-API/common/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LocalFilesystem", func() {

	var (
		storage             LocalStorage
		config              *shared.AppConfig
		mockStringGenerator *MockStringGenerator
		ctx                 = context.Background()
	)

	BeforeEach(func() {
		mockStringGenerator = &MockStringGenerator{}
		config = &shared.AppConfig{
			LocalStoragePath: "test_data",
		}

		storage = LocalStorage{
			StringGenerator: mockStringGenerator,
			Config:          config,
		}
	})

	Context("Store", func() {

		var (
			encodedImage  string
			returnedError error
			uri           string
		)

		BeforeEach(func() {
			mockStringGenerator.On("GenerateUuid").Return("aze3215fe-513df")
		})

		AfterEach(func() {
			os.Remove("test_data/aze3215fe-513df.jpg")
		})

		JustBeforeEach(func() {
			image, _ := ioutil.ReadFile("test_data/DSCF6458.JPG")
			encodedImage = b64.RawStdEncoding.EncodeToString(image)
			uri, returnedError = storage.Store(ctx, encodedImage, "image/jpeg", "")
		})

		It("should not return an error", func() {
			Expect(returnedError).To(BeNil())
		})

		It("should return the right uri", func() {
			Expect(uri).To(Equal("test_data/aze3215fe-513df.jpg"))
		})
	})

	Context("Store", func() {

		var (
			encodedFile   string
			returnedError error
		)

		JustBeforeEach(func() {
			encodedFile, returnedError = storage.Get(ctx, "test_data/DSCF6458.JPG")
		})

		It("should not return an error", func() {
			Expect(returnedError).To(BeNil())
		})

		It("should return the right encodedFile", func() {
			Expect(len(encodedFile)).To(Equal(2001288))
		})
	})

})
