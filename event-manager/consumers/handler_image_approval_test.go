package consumers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Vinubaba/SANTC-API/common/api"
	"github.com/Vinubaba/SANTC-API/common/log"
	. "github.com/Vinubaba/SANTC-API/common/storage/mocks"
	. "github.com/Vinubaba/SANTC-API/event-manager/consumers"
	"github.com/Vinubaba/SANTC-API/event-manager/shared"

	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("HandlerImageApproval", func() {

	var (
		imageApprovalHandler *ImageApprovalHandler
		ctx                  context.Context
		event                Event
		mockConfig           *shared.AppConfig
		mockStorage          = &MockGcs{}
		logger               *log.Logger
		apiClient            api.Client
		mockHttpServer       *httptest.Server
		router               *mux.Router
		returnedError        error
	)

	BeforeEach(func() {
		returnedError = nil
		mockConfig = &shared.AppConfig{}
		logger = log.NewLogger("HandlerImageApprovalTest")

		router = mux.NewRouter()
		router.HandleFunc("/api/v1/children/{childId}/photos", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		})
		router.HandleFunc("/api/v1/children/{childId}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				  "id": "aaa",
				  "daycareId": "peyredragon",
				  "classId": "classid-2",
				  "firstName": "Rickon",
				  "lastName": "Stark",
				  "birthDate": "1992-10-14 00:00:00 +0000 UTC",
				  "gender": "M",
				  "imageUri": "gs://foo/bar.jpg",
				  "startDate": "2018-03-28 00:00:00 +0000 UTC",
				  "notes": "he hates yogurt",
				  "allergies": [
					{
					  "id": "ccc",
					  "allergy": "tomato",
					  "instruction": "take him to the doctor"
					}
				  ],
				  "responsibleId": "id4",
				  "relationship": "mother",
				  "specialInstructions": [
					{
					  "id": "bbb",
					  "childId": "aaa",
					  "instruction": "vegetarian"
					}
				  ]
				}`))
		})

		mockHttpServer = httptest.NewServer(router)

		var err error
		apiClient, err = api.NewDefaultClient("http", strings.TrimPrefix(mockHttpServer.URL, "http://"))
		if err != nil {
			panic(err)
		}
		imageApprovalHandler = &ImageApprovalHandler{
			Config:    mockConfig,
			Storage:   mockStorage,
			Logger:    logger,
			ApiClient: apiClient,
		}

		ctx = context.Background()
		event = Event{
			Type:     "imageApproval",
			SenderId: "dd3a81f0-6432-4ddf-842a-b82a3911dadb",
			ImageApproval: &ImageApproval{
				Image:   b64ImageExample,
				ChildId: "4d8b9d3f-1478-4215-a240-85974b940c97",
			},
		}
	})

	AfterEach(func() {
		mockStorage.Reset()
	})

	JustBeforeEach(func() {
		returnedError = imageApprovalHandler.Handle(ctx, event)
	})

	Context("default", func() {
		BeforeEach(func() {
			mockStorage.On("Store", mock.Anything, mock.Anything, mock.Anything).Return("toto.jpg", nil)
		})
		mockStorage.AssertStoredImage("daycares/peyredragon/children/aaa")
		It("should not return an error", func() {
			Expect(returnedError).To(BeNil())
		})
	})
})
