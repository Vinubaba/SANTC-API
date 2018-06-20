package storage_test

import (
	"context"
	b64 "encoding/base64"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Vinubaba/SANTC-API/api/shared"
	. "github.com/Vinubaba/SANTC-API/api/shared/mocks"

	. "github.com/Vinubaba/SANTC-API/common/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gcs", func() {

	var (
		storage             *GoogleStorage
		config              *shared.AppConfig
		mockStringGenerator *MockStringGenerator
		ctx                 = context.Background()
	)

	BeforeEach(func() {
		mockStringGenerator = &MockStringGenerator{}
		bucketSa := os.Getenv("BUCKET_SERVICE_ACCOUNT_PATH")
		if bucketSa == "" {
			bucketSa = `C:\Users\arthur\gocode\src\github.com\Vinubaba\deployment\bucket-sa.json`
		}
		config = &shared.AppConfig{
			BucketServiceAccount: bucketSa,
			BucketName:           "teddycare",
		}
		var err error
		storage, err = New(ctx, Options{
			CredentialsFile: bucketSa,
			BucketName:      "teddycare",
		})
		if err != nil {
			panic(err)
		}
		storage.StringGenerator = mockStringGenerator
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
			fileName, storeError = storage.Store(ctx, "data:image/jpeg;base64,"+encodedImage, "")

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
			b, _ := ioutil.ReadAll(getResponse.Body)
			Expect(b).To(Equal(image))
			Expect(getError).To(BeNil())
			Expect(getResponse.StatusCode).To(Equal(http.StatusOK))

			// Delete
			Expect(deleteError).To(BeNil())
		})

	})

})
