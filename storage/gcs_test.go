package storage_test

import (
	b64 "encoding/base64"
	"io/ioutil"

	"github.com/DigitalFrameworksLLC/teddycare/shared"
	. "github.com/DigitalFrameworksLLC/teddycare/shared/mocks"
	. "github.com/DigitalFrameworksLLC/teddycare/storage"

	"context"
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
)

var _ = Describe("Gcs", func() {

	var (
		storage             GoogleStorage
		config              *shared.AppConfig
		mockStringGenerator *MockStringGenerator
		ctx                 = context.Background()
	)

	BeforeEach(func() {
		mockStringGenerator = &MockStringGenerator{}
		config = &shared.AppConfig{
			BucketServiceAccount: `C:\Users\arthur\gocode\src\github.com\DigitalFrameworksLLC\teddycare\adm-bucket-sa.json`,
			BucketImagesName:     "teddycare-images",
		}

		b, err := ioutil.ReadFile(config.BucketServiceAccount)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(b, &config.BucketServiceAccountDetails); err != nil {
			panic(err)
		}

		storage = GoogleStorage{
			StringGenerator: mockStringGenerator,
			Config:          config,
		}
	})

	Context("Store, Get and Delete", func() {

		var (
			image                             []byte
			encodedImage                      string
			storeError, getError, deleteError error
			uri                               string
			fileName                          string
			getResponse                       *http.Response
		)

		BeforeEach(func() {
			mockStringGenerator.On("GenerateUuid").Return("image1")

			// First store
			image, _ = ioutil.ReadFile("test_data/DSCF6458.JPG")
			encodedImage = b64.RawStdEncoding.EncodeToString(image)
			fileName, storeError = storage.Store(ctx, encodedImage, "image/jpeg")

			// Then get
			uri, getError = storage.Get(ctx, fileName)
			getResponse, _ = http.Get(uri)

			// Finally delete
			deleteError = storage.Delete(ctx, fileName)
		})

		// Only 1 test to avoid making too much connexion
		It("should create, get and delete the image", func() {
			// Store
			Expect(storeError).To(BeNil())
			Expect(fileName).To(Equal("image1.jpg"))

			// Get
			Expect(getError).To(BeNil())
			Expect(getResponse.StatusCode).To(Equal(http.StatusOK))
			b, _ := ioutil.ReadAll(getResponse.Body)
			Expect(b).To(Equal(image))

			// Delete
			Expect(deleteError).To(BeNil())
		})

	})

})
