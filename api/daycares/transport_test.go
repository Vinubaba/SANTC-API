package daycares_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Vinubaba/SANTC-API/api/authentication"
	. "github.com/Vinubaba/SANTC-API/api/daycares"
	. "github.com/Vinubaba/SANTC-API/api/firebase/mocks"
	"github.com/Vinubaba/SANTC-API/api/shared"
	. "github.com/Vinubaba/SANTC-API/api/shared/mocks"
	"github.com/Vinubaba/SANTC-API/api/store"
	"github.com/Vinubaba/SANTC-API/api/users"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Transport", func() {

	var (
		router   *mux.Router
		recorder *httptest.ResponseRecorder

		concreteStore       *store.Store
		concreteDb          *gorm.DB
		mockStringGenerator *MockStringGenerator
		mockFirebaseClient  *MockClient

		authenticator *authentication.Authenticator

		claims                                            map[string]interface{}
		reqToUse                                          *http.Request
		httpMethodToUse, httpEndpointToUse, httpBodyToUse string
	)

	var (
		assertHttpCode = func(code int) {
			It(fmt.Sprintf("should respond with status code %d", code), func() {
				Expect(recorder.Code).To(Equal(code))
			})
		}

		assertReturnedDaycareWithIds = func(ids ...string) {
			It(fmt.Sprintf("should respond %d daycares", len(ids)), func() {
				if len(ids) == 0 {
					panic("cant test with 0 id")
				}
				daycaresTransport := []DaycareTransport{}
				json.Unmarshal([]byte(recorder.Body.String()), &daycaresTransport)
				Expect(daycaresTransport).To(HaveLen(len(ids)))

				returnedId := func(id string, response []DaycareTransport) bool {
					for _, r := range response {
						if r.Id == id {
							return true
						}
					}
					return false
				}

				for _, id := range ids {
					if !returnedId(id, daycaresTransport) {
						Fail(fmt.Sprintf("%s was not found in response %s", id, daycaresTransport))
					}
				}
			})
		}

		assertReturnedNoPayload = func() {
			It("should respond with 1 users", func() {
				Expect(recorder.Body.String()).To(Equal(""))
			})
		}

		assertReturnedSingleDaycare = func(daycareJson string) {
			It("should respond with 1 daycare", func() {
				Expect(recorder.Body.String()).To(MatchJSON(daycareJson))
			})
		}

		assertJsonResponse = func(response string) {
			It("should respond with json response", func() {
				Expect(recorder.Body.String()).To(MatchJSON(response))
			})
		}
	)

	BeforeEach(func() {
		concreteDb = shared.NewDbInstance(false)

		mockStringGenerator = &MockStringGenerator{}
		mockStringGenerator.On("GenerateUuid").Return("aaa").Once()
		mockStringGenerator.On("GenerateUuid").Return("bbb").Once()

		concreteStore = &store.Store{
			Db:              concreteDb,
			StringGenerator: mockStringGenerator,
		}

		mockFirebaseClient = &MockClient{}
		mockFirebaseClient.On("DeleteUser", mock.Anything, mock.Anything).Return(nil)

		userService := &users.UserService{
			FirebaseClient: mockFirebaseClient,
			Store:          concreteStore,
		}
		logger := shared.NewLogger("teddycare")

		authenticator = &authentication.Authenticator{
			UserService: userService,
			Logger:      logger,
		}

		daycareService := &DaycareService{
			Store:  concreteStore,
			Logger: logger,
		}

		httpMethodToUse = ""
		httpEndpointToUse = ""
		httpBodyToUse = ""

		router = mux.NewRouter()
		opts := []kithttp.ServerOption{
			kithttp.ServerErrorLogger(logger),
			kithttp.ServerErrorEncoder(EncodeError),
		}

		handlerFactory := HandlerFactory{
			Service: daycareService,
		}

		router.Handle("/daycares", authenticator.Roles(handlerFactory.Add(opts), shared.ROLE_ADMIN)).Methods(http.MethodPost)
		router.Handle("/daycares", authenticator.Roles(handlerFactory.List(opts), shared.ROLE_ADMIN)).Methods(http.MethodGet)
		router.Handle("/daycares/{daycareId}", authenticator.Roles(handlerFactory.Get(opts), shared.ROLE_ADMIN)).Methods(http.MethodGet)
		router.Handle("/daycares/{daycareId}", authenticator.Roles(handlerFactory.Update(opts), shared.ROLE_ADMIN)).Methods(http.MethodPatch)
		router.Handle("/daycares/{daycareId}", authenticator.Roles(handlerFactory.Delete(opts), shared.ROLE_ADMIN)).Methods(http.MethodDelete)

		recorder = httptest.NewRecorder()

		shared.SetDbInitialState()
	})

	AfterEach(func() {
		concreteDb.Close()
	})

	BeforeEach(func() {
		claims = map[string]interface{}{
			"userId":                   "",
			"daycareId":                "peyredragon",
			shared.ROLE_TEACHER:        false,
			shared.ROLE_OFFICE_MANAGER: false,
			shared.ROLE_ADULT:          false,
			shared.ROLE_ADMIN:          false,
		}
	})

	JustBeforeEach(func() {
		reqToUse, _ = http.NewRequest(httpMethodToUse, httpEndpointToUse, strings.NewReader(httpBodyToUse))
		reqToUse = reqToUse.WithContext(context.WithValue(context.Background(), "claims", claims))
		router.ServeHTTP(recorder, reqToUse)
	})

	Describe("DAYCARES", func() {

		Describe("LIST", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodGet
				httpEndpointToUse = "/daycares"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertReturnedDaycareWithIds("namek", "peyredragon")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() {
					claims[shared.ROLE_OFFICE_MANAGER] = true
					claims["daycareId"] = "namek"
				})
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[shared.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[shared.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to list daycares: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

		})

		Describe("GET", func() {

			var (
				jsonDaycareRef = `{
				  "id": "namek",
				  "name": "namek",
				  "address_1": "namek",
				  "address_2": "namek",
				  "city": "namek",
				  "state": "namek",
				  "zip": "namek"
				}`
			)

			BeforeEach(func() {
				httpMethodToUse = http.MethodGet
				httpEndpointToUse = "/daycares/namek"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertReturnedSingleDaycare(jsonDaycareRef)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[shared.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[shared.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[shared.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to get daycare: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the daycare does not exists", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					httpEndpointToUse = "/daycares/foo"
				})
				assertJsonResponse(`{"error":"failed to get daycare: daycare not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("DELETE", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodDelete
				httpEndpointToUse = "/daycares/namek"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertJsonResponse(`{"error": "not implemented yet"}`)
				assertHttpCode(http.StatusNotImplemented)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[shared.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[shared.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[shared.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error": "not implemented yet"}`)
				assertHttpCode(http.StatusNotImplemented)
			})

			Context("When the daycare does not exists", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					httpEndpointToUse = "/daycares/foo"
				})
				assertJsonResponse(`{"error": "not implemented yet"}`)
				assertHttpCode(http.StatusNotImplemented)
			})

		})

		Describe("UPDATE", func() {

			var (
				jsonUpdatedDaycare = `{
				  "id": "namek",
				  "name": "namek 2",
				  "address_1": "namek 2",
				  "address_2": "namek 2",
				  "city": "namek 2",
				  "state": "namek 2",
				  "zip": "namek 2"
				}`
			)

			BeforeEach(func() {
				httpMethodToUse = http.MethodPatch
				httpEndpointToUse = "/daycares/namek"
				httpBodyToUse = `{
				  "name": "namek 2",
				  "address_1": "namek 2",
				  "address_2": "namek 2",
				  "city": "namek 2",
				  "state": "namek 2",
				  "zip": "namek 2"
				}`
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertReturnedSingleDaycare(jsonUpdatedDaycare)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[shared.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[shared.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[shared.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to update daycare: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the daycare does not exists", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					httpEndpointToUse = "/daycares/foo"
				})
				assertJsonResponse(`{"error":"failed to update daycare: daycare not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("CREATE", func() {

			var (
				jsonCreatedDaycare = `{
				  "id": "aaa",
				  "name": "centurion",
				  "address_1": "centurion",
				  "address_2": "centurion",
				  "city": "centurion",
				  "state": "centurion",
				  "zip": "centurion"
				}`
			)

			BeforeEach(func() {
				httpMethodToUse = http.MethodPost
				httpEndpointToUse = "/daycares"
				httpBodyToUse = `{"name": "centurion", "address_1": "centurion", "address_2": "centurion", "city": "centurion", "state": "centurion", "zip": "centurion"}`
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertReturnedSingleDaycare(jsonCreatedDaycare)
				assertHttpCode(http.StatusCreated)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[shared.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[shared.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[shared.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to add daycare: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

		})

	})

})
