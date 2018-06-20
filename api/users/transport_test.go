package users_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Vinubaba/SANTC-API/api/authentication"
	"github.com/Vinubaba/SANTC-API/api/shared"
	. "github.com/Vinubaba/SANTC-API/api/shared/mocks"
	. "github.com/Vinubaba/SANTC-API/api/users"
	. "github.com/Vinubaba/SANTC-API/common/firebase/mocks"
	"github.com/Vinubaba/SANTC-API/common/log"
	"github.com/Vinubaba/SANTC-API/common/roles"
	"github.com/Vinubaba/SANTC-API/common/storage/mocks"
	"github.com/Vinubaba/SANTC-API/common/store"

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
		mockStorage         *mocks.MockGcs
		mockFirebaseClient  *MockClient
		config              *shared.AppConfig

		authenticator *authentication.Authenticator

		claims                                            map[string]interface{}
		reqToUse                                          *http.Request
		httpMethodToUse, httpEndpointToUse, httpBodyToUse string

		mockImageUriName = "bar.jpg"
	)

	var (
		assertHttpCode = func(code int) {
			It(fmt.Sprintf("should respond with status code %d", code), func() {
				Expect(recorder.Code).To(Equal(code))
			})
		}

		assertReturnedUsersWithIds = func(ids ...string) {
			It(fmt.Sprintf("should respond %d users", len(ids)), func() {
				if len(ids) == 0 {
					panic("cant test with 0 id")
				}
				usersTransport := []UserTransport{}
				json.Unmarshal([]byte(recorder.Body.String()), &usersTransport)
				Expect(usersTransport).To(HaveLen(len(ids)))

				returnedId := func(id string, response []UserTransport) bool {
					for _, r := range response {
						if r.Id == id {
							return true
						}
					}
					return false
				}

				for _, id := range ids {
					if !returnedId(id, usersTransport) {
						Fail(fmt.Sprintf("%s was not found in response %s", id, usersTransport))
					}
				}
			})
		}

		assertReturnedNoPayload = func() {
			It("should respond with 1 users", func() {
				Expect(recorder.Body.String()).To(Equal(""))
			})
		}

		assertReturnedSingleUser = func(userJson string) {
			It("should respond with 1 users", func() {
				Expect(recorder.Body.String()).To(MatchJSON(userJson))
			})
		}

		assertJsonResponse = func(response string) {
			It("should respond with json response", func() {
				Expect(recorder.Header().Get("Content-Type")).To(ContainSubstring("application/json"))
				Expect(recorder.Body.String()).To(MatchJSON(response))
			})
		}
	)

	BeforeEach(func() {
		logger := log.NewLogger("teddycare")

		concreteDb = shared.NewDbInstance(true, logger)

		mockStringGenerator = &MockStringGenerator{}
		mockStringGenerator.On("GenerateUuid").Return("aaa").Once()

		mockStorage = &mocks.MockGcs{}
		mockStorage.On("Get", mock.Anything, mock.Anything).Return("gs://foo/"+mockImageUriName, nil)
		mockStorage.On("Delete", mock.Anything, mock.Anything).Return(nil)

		mockFirebaseClient = &MockClient{}
		mockFirebaseClient.On("DeleteUserByEmail", mock.Anything, mock.Anything).Return(nil)

		recorder = httptest.NewRecorder()

		concreteStore = &store.Store{
			Db:              concreteDb,
			StringGenerator: mockStringGenerator,
		}

		config = &shared.AppConfig{
			PublicDaycareId: "PUBLIC",
		}

		userService := &UserService{
			FirebaseClient: mockFirebaseClient,
			Store:          concreteStore,
			Storage:        mockStorage,
			Logger:         logger,
			Config:         config,
		}

		authenticator = &authentication.Authenticator{
			UserService: userService,
			Logger:      logger,
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
			Service: userService,
		}

		router.Handle("/office-managers", authenticator.Roles(handlerFactory.ListOfficeManager(opts), roles.ROLE_ADMIN)).Methods(http.MethodGet)
		router.Handle("/office-managers/{id}", authenticator.Roles(handlerFactory.GetOfficeManager(opts), roles.ROLE_ADMIN)).Methods(http.MethodGet)
		router.Handle("/office-managers/{id}", authenticator.Roles(handlerFactory.DeleteOfficeManager(opts), roles.ROLE_ADMIN)).Methods(http.MethodDelete)
		router.Handle("/office-managers/{id}", authenticator.Roles(handlerFactory.UpdateOfficeManager(opts), roles.ROLE_ADMIN)).Methods(http.MethodPatch)

		router.Handle("/teachers", authenticator.Roles(handlerFactory.CreateTeacher(opts), roles.ROLE_ADMIN, roles.ROLE_OFFICE_MANAGER)).Methods(http.MethodPost)
		router.Handle("/teachers", authenticator.Roles(handlerFactory.ListTeacher(opts), roles.ROLE_ADMIN, roles.ROLE_OFFICE_MANAGER, roles.ROLE_ADULT)).Methods(http.MethodGet)
		router.Handle("/teachers/{id}", authenticator.Roles(handlerFactory.GetTeacher(opts), roles.ROLE_ADMIN, roles.ROLE_OFFICE_MANAGER)).Methods(http.MethodGet)
		router.Handle("/teachers/{id}", authenticator.Roles(handlerFactory.DeleteTeacher(opts), roles.ROLE_ADMIN, roles.ROLE_OFFICE_MANAGER)).Methods(http.MethodDelete)
		router.Handle("/teachers/{id}", authenticator.Roles(handlerFactory.UpdateTeacher(opts), roles.ROLE_ADMIN, roles.ROLE_OFFICE_MANAGER)).Methods(http.MethodPatch)
		router.Handle("/teachers/{id}/classes", authenticator.Roles(handlerFactory.SetTeacherClass(opts), roles.ROLE_ADMIN, roles.ROLE_OFFICE_MANAGER)).Methods(http.MethodPost)

		router.Handle("/adults", authenticator.Roles(handlerFactory.CreateAdult(opts), roles.ROLE_ADMIN, roles.ROLE_OFFICE_MANAGER)).Methods(http.MethodPost)
		router.Handle("/adults", authenticator.Roles(handlerFactory.ListAdult(opts), roles.ROLE_ADMIN, roles.ROLE_OFFICE_MANAGER)).Methods(http.MethodGet)
		router.Handle("/adults/{id}", authenticator.Roles(handlerFactory.GetAdult(opts), roles.ROLE_ADMIN, roles.ROLE_OFFICE_MANAGER)).Methods(http.MethodGet)
		router.Handle("/adults/{id}", authenticator.Roles(handlerFactory.DeleteAdult(opts), roles.ROLE_ADMIN, roles.ROLE_OFFICE_MANAGER)).Methods(http.MethodDelete)
		router.Handle("/adults/{id}", authenticator.Roles(handlerFactory.UpdateAdult(opts), roles.ROLE_ADMIN, roles.ROLE_OFFICE_MANAGER)).Methods(http.MethodPatch)

		shared.SetDbInitialState()
	})

	AfterEach(func() {
		concreteDb.Close()
	})

	BeforeEach(func() {
		claims = map[string]interface{}{
			"userId":                  "",
			"daycareId":               "peyredragon",
			roles.ROLE_TEACHER:        false,
			roles.ROLE_OFFICE_MANAGER: false,
			roles.ROLE_ADULT:          false,
			roles.ROLE_ADMIN:          false,
		}
	})

	JustBeforeEach(func() {
		reqToUse, _ = http.NewRequest(httpMethodToUse, httpEndpointToUse, strings.NewReader(httpBodyToUse))
		reqToUse = reqToUse.WithContext(context.WithValue(context.Background(), "claims", claims))
		router.ServeHTTP(recorder, reqToUse)
	})

	Describe("ADULTS", func() {

		Describe("LIST", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodGet
				httpEndpointToUse = "/adults"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedUsersWithIds("id5", "id10")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[roles.ROLE_OFFICE_MANAGER] = true })
				assertReturnedUsersWithIds("id5")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[roles.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

		})

		Describe("GET", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodGet
				httpEndpointToUse = "/adults/id5"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedSingleUser(`{"id":"id5","firstName":"Sansa","lastName":"Stark","gender":"F","email":"sansa.stark@got.com","phone":"+3365651","address_1":"address","address_2":"floor","city":"Peyredragon","state":"WESTEROS","zip":"31400","imageUri":"http://image.com","roles":["adult"],"daycareId":"peyredragon"}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[roles.ROLE_OFFICE_MANAGER] = true })
				assertReturnedSingleUser(`{"id":"id5","firstName":"Sansa","lastName":"Stark","gender":"F","email":"sansa.stark@got.com","phone":"+3365651","address_1":"address","address_2":"floor","city":"Peyredragon","state":"WESTEROS","zip":"31400","imageUri":"http://image.com","roles":["adult"],"daycareId":"peyredragon"}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[roles.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to get user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					httpEndpointToUse = "/adults/foo"
				})
				assertJsonResponse(`{"error":"failed to get user: user not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("DELETE", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodDelete
				httpEndpointToUse = "/adults/id5"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusNoContent)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[roles.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusNoContent)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[roles.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to delete user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					httpEndpointToUse = "/adults/foo"
				})
				assertJsonResponse(`{"error":"failed to delete user: user not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("UPDATE", func() {

			BeforeEach(func() {
				mockStorage.On("Store", mock.Anything, mock.Anything, mock.Anything).Return(mockImageUriName, nil)
				httpMethodToUse = http.MethodPatch
				httpEndpointToUse = "/adults/id5"
				httpBodyToUse = `{"address_1": "8 RUE PIERRE DELDI", "address_2": "VILLA 13", "imageUri": "data:image/jpeg;base64,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"}`
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedSingleUser(`{"daycareId":"peyredragon","id":"id5","firstName":"Sansa","lastName":"Stark","gender":"F","email":"sansa.stark@got.com","phone":"+3365651","address_1": "8 RUE PIERRE DELDI", "address_2": "VILLA 13","city":"Peyredragon","state":"WESTEROS","zip":"31400","imageUri":"gs://foo/bar.jpg","roles":["adult"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[roles.ROLE_OFFICE_MANAGER] = true })
				assertReturnedSingleUser(`{"daycareId":"peyredragon","id":"id5","firstName":"Sansa","lastName":"Stark","gender":"F","email":"sansa.stark@got.com","phone":"+3365651","address_1": "8 RUE PIERRE DELDI", "address_2": "VILLA 13","city":"Peyredragon","state":"WESTEROS","zip":"31400","imageUri":"gs://foo/bar.jpg","roles":["adult"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[roles.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to update user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					httpEndpointToUse = "/adults/foo"
				})
				assertJsonResponse(`{"error":"failed to update user: user not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("CREATE", func() {

			BeforeEach(func() {
				mockStorage.On("Store", mock.Anything, mock.Anything, mock.Anything).Return(mockImageUriName, nil)
				httpMethodToUse = http.MethodPost
				httpEndpointToUse = "/adults"
				httpBodyToUse = fmt.Sprintf(`
					{
						"firstName": "Elaria",
						"lastName": "Sand",
						"gender": "M",
						"email": "saint.sulp.la.pointe@gmail.com",
						"phone": "0633326825",
						"address_1": "8 RUE PIERRE DELDI",
						"address_2": "VILLA 13",
						"city": "TOULOUSE",
						"state": "France",
						"zip": "31100",
						"imageUri": "%s",
						"daycareId": "peyredragon"
					}`, b64imageTest)
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedSingleUser(`{"daycareId":"peyredragon","id": "aaa","firstName": "Elaria","lastName": "Sand","gender": "M","email": "saint.sulp.la.pointe@gmail.com","phone": "0633326825","address_1": "8 RUE PIERRE DELDI","address_2": "VILLA 13","city": "TOULOUSE","state": "France","zip": "31100","imageUri": "gs://foo/bar.jpg","roles": ["adult"], "daycareId": "peyredragon"}`)
				assertHttpCode(http.StatusCreated)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[roles.ROLE_OFFICE_MANAGER] = true })
				assertReturnedSingleUser(`{"daycareId":"peyredragon","id": "aaa","firstName": "Elaria","lastName": "Sand","gender": "M","email": "saint.sulp.la.pointe@gmail.com","phone": "0633326825","address_1": "8 RUE PIERRE DELDI","address_2": "VILLA 13","city": "TOULOUSE","state": "France","zip": "31100","imageUri": "gs://foo/bar.jpg","roles": ["adult"], "daycareId": "peyredragon"}`)
				assertHttpCode(http.StatusCreated)
			})

			Context("When user is an office manager and tries to create adult for another daycare", func() {
				BeforeEach(func() {
					claims[roles.ROLE_OFFICE_MANAGER] = true
					httpBodyToUse = fmt.Sprintf(`
					{
						"firstName": "Elaria",
						"lastName": "Sand",
						"gender": "M",
						"email": "saint.sulp.la.pointe@gmail.com",
						"phone": "0633326825",
						"address_1": "8 RUE PIERRE DELDI",
						"address_2": "VILLA 13",
						"city": "TOULOUSE",
						"state": "France",
						"zip": "31100",
						"imageUri": "%s",
						"daycareId": "foobar"
					}`, b64imageTest)
				})
				assertJsonResponse(`{"error":"cannot create user for another daycare"}`)
				assertHttpCode(http.StatusForbidden)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[roles.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to create user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

		})

	})

	Describe("OFFICE MANAGER", func() {

		Describe("LIST", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodGet
				httpEndpointToUse = "/office-managers"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedUsersWithIds("id2", "id3", "id7", "id8")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[roles.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[roles.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

		})

		Describe("GET", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodGet
				httpEndpointToUse = "/office-managers/id2"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedSingleUser(`{"daycareId":"peyredragon","id":"id2","firstName":"John","lastName":"Snow","gender":"M","email":"john.snow@got.com","phone":"+3365651","address_1":"address","address_2":"floor","city":"Peyredragon","state":"WESTEROS","zip":"31400","imageUri":"http://image.com","roles":["officemanager"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[roles.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[roles.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to get user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					httpEndpointToUse = "/office-managers/foo"
				})
				assertJsonResponse(`{"error":"failed to get user: user not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("DELETE", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodDelete
				httpEndpointToUse = "/office-managers/id2"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusNoContent)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[roles.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[roles.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to delete user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					httpEndpointToUse = "/office-managers/foo"
				})
				assertJsonResponse(`{"error":"failed to delete user: user not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("UPDATE", func() {

			BeforeEach(func() {
				mockStorage.On("Store", mock.Anything, mock.Anything, mock.Anything).Return(mockImageUriName, nil)
				httpMethodToUse = http.MethodPatch
				httpEndpointToUse = "/office-managers/id2"
				httpBodyToUse = `{"address_1": "8 RUE PIERRE DELDI", "address_2": "VILLA 13", "imageUri": "data:image/jpeg;base64,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"}`
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedSingleUser(`{"daycareId": "peyredragon", "id": "id2","firstName": "John","lastName": "Snow","gender": "M","email": "john.snow@got.com","phone": "+3365651","address_1": "8 RUE PIERRE DELDI","address_2": "VILLA 13","city": "Peyredragon","state": "WESTEROS","zip": "31400","imageUri": "gs://foo/bar.jpg","roles": ["officemanager"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[roles.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[roles.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to update user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					httpEndpointToUse = "/adults/foo"
				})
				assertJsonResponse(`{"error":"failed to update user: user not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

	})

	Describe("TEACHER", func() {

		Describe("LIST", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodGet
				httpEndpointToUse = "/teachers"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedUsersWithIds("id4", "id9")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[roles.ROLE_OFFICE_MANAGER] = true })
				assertReturnedUsersWithIds("id4")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADULT] = true
					claims["userId"] = "id4"
				})
				// Should list teacher of childs
				assertReturnedUsersWithIds("id4")
				assertHttpCode(http.StatusOK)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

		})

		Describe("GET", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodGet
				httpEndpointToUse = "/teachers/id4"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedSingleUser(`{"daycareId": "peyredragon","id":"id4","firstName":"Caitlyn","lastName":"Stark","gender":"F","email":"caitlyn.stark@got.com","phone":"+3365651","address_1":"address","address_2":"floor","city":"Peyredragon","state":"WESTEROS","zip":"31400","imageUri":"http://image.com","roles":["teacher"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[roles.ROLE_OFFICE_MANAGER] = true })
				assertReturnedSingleUser(`{"daycareId": "peyredragon","id":"id4","firstName":"Caitlyn","lastName":"Stark","gender":"F","email":"caitlyn.stark@got.com","phone":"+3365651","address_1":"address","address_2":"floor","city":"Peyredragon","state":"WESTEROS","zip":"31400","imageUri":"http://image.com","roles":["teacher"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[roles.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to get user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					httpEndpointToUse = "/teachers/foo"
				})
				assertJsonResponse(`{"error":"failed to get user: user not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("DELETE", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodDelete
				httpEndpointToUse = "/teachers/id4"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusNoContent)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[roles.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusNoContent)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[roles.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to delete user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					httpEndpointToUse = "/adults/foo"
				})
				assertJsonResponse(`{"error":"failed to delete user: user not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("UPDATE", func() {

			BeforeEach(func() {
				mockStorage.On("Store", mock.Anything, mock.Anything, mock.Anything).Return(mockImageUriName, nil)
				httpMethodToUse = http.MethodPatch
				httpEndpointToUse = "/teachers/id4"
				httpBodyToUse = `{"address_1": "8 RUE PIERRE DELDI", "address_2": "VILLA 13", "imageUri": "data:image/jpeg;base64,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"}`
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedSingleUser(`{"daycareId": "peyredragon","id":"id4","firstName":"Caitlyn","lastName":"Stark","gender":"F","email":"caitlyn.stark@got.com","phone":"+3365651","address_1":"8 RUE PIERRE DELDI","address_2":"VILLA 13","city":"Peyredragon","state":"WESTEROS","zip":"31400","imageUri":"gs://foo/bar.jpg","roles":["teacher"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[roles.ROLE_OFFICE_MANAGER] = true })
				assertReturnedSingleUser(`{"daycareId": "peyredragon","id":"id4","firstName":"Caitlyn","lastName":"Stark","gender":"F","email":"caitlyn.stark@got.com","phone":"+3365651","address_1":"8 RUE PIERRE DELDI","address_2":"VILLA 13","city":"Peyredragon","state":"WESTEROS","zip":"31400","imageUri":"gs://foo/bar.jpg","roles":["teacher"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[roles.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to update user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					httpEndpointToUse = "/adults/foo"
				})
				assertJsonResponse(`{"error":"failed to update user: user not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("CREATE", func() {

			BeforeEach(func() {
				mockStorage.On("Store", mock.Anything, mock.Anything, mock.Anything).Return(mockImageUriName, nil)
				httpMethodToUse = http.MethodPost
				httpEndpointToUse = "/teachers"
				httpBodyToUse = fmt.Sprintf(`
					{
						"firstName": "Elaria",
						"lastName": "Sand",
						"gender": "M",
						"email": "saint.sulp.la.pointe@gmail.com",
						"phone": "0633326825",
						"address_1": "8 RUE PIERRE DELDI",
						"address_2": "VILLA 13",
						"city": "TOULOUSE",
						"state": "France",
						"zip": "31100",
						"imageUri": "%s",
						"daycareId": "peyredragon"
					}`, b64imageTest)
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedSingleUser(`{"id": "aaa","firstName": "Elaria","lastName": "Sand","gender": "M","email": "saint.sulp.la.pointe@gmail.com","phone": "0633326825","address_1": "8 RUE PIERRE DELDI","address_2": "VILLA 13","city": "TOULOUSE","state": "France","zip": "31100","imageUri": "gs://foo/bar.jpg","roles": ["teacher"], "daycareId": "peyredragon"}`)
				assertHttpCode(http.StatusCreated)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[roles.ROLE_OFFICE_MANAGER] = true })
				assertReturnedSingleUser(`{"id": "aaa","firstName": "Elaria","lastName": "Sand","gender": "M","email": "saint.sulp.la.pointe@gmail.com","phone": "0633326825","address_1": "8 RUE PIERRE DELDI","address_2": "VILLA 13","city": "TOULOUSE","state": "France","zip": "31100","imageUri": "gs://foo/bar.jpg","roles": ["teacher"], "daycareId": "peyredragon"}`)
				assertHttpCode(http.StatusCreated)
			})

			Context("When user is an office manager and tries to create teacher for another daycare", func() {
				BeforeEach(func() {
					claims[roles.ROLE_OFFICE_MANAGER] = true
					httpBodyToUse = fmt.Sprintf(`
					{
						"firstName": "Elaria",
						"lastName": "Sand",
						"gender": "M",
						"email": "saint.sulp.la.pointe@gmail.com",
						"phone": "0633326825",
						"address_1": "8 RUE PIERRE DELDI",
						"address_2": "VILLA 13",
						"city": "TOULOUSE",
						"state": "France",
						"zip": "31100",
						"imageUri": "%s",
						"daycareId": "foobar"
					}`, b64imageTest)
				})
				assertJsonResponse(`{"error":"cannot create user for another daycare"}`)
				assertHttpCode(http.StatusForbidden)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[roles.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to create user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

		})

		Describe("SET TEACHER CLASS", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodPost
				httpEndpointToUse = "/teachers/id4/classes"
				httpBodyToUse = `{
						"classId": "classid-2"
					}`
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager from the same daycare than teacher and class", func() {
				BeforeEach(func() {
					claims[roles.ROLE_OFFICE_MANAGER] = true
					claims["daycareId"] = "peyredragon"
				})
				assertReturnedNoPayload()
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager from a different daycare than teacher and class", func() {
				BeforeEach(func() {
					claims[roles.ROLE_OFFICE_MANAGER] = true
					claims["daycareId"] = "namek"
				})
				assertJsonResponse(`{"error": "class not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[roles.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[roles.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

		})

	})

})
