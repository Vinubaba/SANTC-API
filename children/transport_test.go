package children_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Vinubaba/SANTC-API/authentication"
	. "github.com/Vinubaba/SANTC-API/children"
	. "github.com/Vinubaba/SANTC-API/firebase/mocks"
	. "github.com/Vinubaba/SANTC-API/shared/mocks"
	. "github.com/Vinubaba/SANTC-API/storage/mocks"
	"github.com/Vinubaba/SANTC-API/store"
	"github.com/Vinubaba/SANTC-API/users"

	"github.com/Vinubaba/SANTC-API/shared"
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
		mockStorage         *MockGcs
		mockFirebaseClient  *MockClient

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

		assertReturnedChildrenWithIds = func(ids ...string) {
			It(fmt.Sprintf("should respond %d children", len(ids)), func() {
				if len(ids) == 0 {
					panic("cant test with 0 id")
				}
				childrenTransport := []ChildTransport{}
				json.Unmarshal([]byte(recorder.Body.String()), &childrenTransport)
				Expect(childrenTransport).To(HaveLen(len(ids)))

				returnedId := func(id string, response []ChildTransport) bool {
					for _, r := range response {
						if r.Id == id {
							return true
						}
					}
					return false
				}

				for _, id := range ids {
					if !returnedId(id, childrenTransport) {
						Fail(fmt.Sprintf("%s was not found in response %s", id, childrenTransport))
					}
				}
			})
		}

		assertReturnedNoPayload = func() {
			It("should respond with 1 users", func() {
				Expect(recorder.Body.String()).To(Equal(""))
			})
		}

		assertReturnedSingleChild = func(childJson string) {
			It("should respond with 1 child", func() {
				Expect(recorder.Body.String()).To(MatchJSON(childJson))
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
		mockStringGenerator.On("GenerateUuid").Return("ccc").Once()
		mockStringGenerator.On("GenerateUuid").Return("ddd").Once()

		mockStorage = &MockGcs{}
		mockStorage.On("Get", mock.Anything, mock.Anything).Return("gs://foo/"+mockImageUriName, nil)
		mockStorage.On("Delete", mock.Anything, mock.Anything).Return(nil)

		concreteStore = &store.Store{
			Db:              concreteDb,
			StringGenerator: mockStringGenerator,
		}

		mockFirebaseClient = &MockClient{}
		mockFirebaseClient.On("DeleteUser", mock.Anything, mock.Anything).Return(nil)

		userService := &users.UserService{
			FirebaseClient: mockFirebaseClient,
			Store:          concreteStore,
			Storage:        mockStorage,
		}
		logger := shared.NewLogger("teddycare")

		authenticator = &authentication.Authenticator{
			UserService: userService,
			Logger:      logger,
		}

		childService := &ChildService{
			Storage: mockStorage,
			Store:   concreteStore,
			Logger:  logger,
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
			Service: childService,
		}

		router.Handle("/children", authenticator.Roles(handlerFactory.Add(opts), shared.ROLE_OFFICE_MANAGER, shared.ROLE_ADMIN)).Methods(http.MethodPost)
		router.Handle("/children", authenticator.Roles(handlerFactory.List(opts), shared.ROLE_OFFICE_MANAGER, shared.ROLE_ADULT, shared.ROLE_ADMIN, shared.ROLE_TEACHER)).Methods(http.MethodGet)
		router.Handle("/children/{childId}", authenticator.Roles(handlerFactory.Get(opts), shared.ROLE_OFFICE_MANAGER, shared.ROLE_ADULT, shared.ROLE_ADMIN, shared.ROLE_TEACHER)).Methods(http.MethodGet)
		router.Handle("/children/{childId}", authenticator.Roles(handlerFactory.Update(opts), shared.ROLE_OFFICE_MANAGER, shared.ROLE_ADULT, shared.ROLE_ADMIN)).Methods(http.MethodPatch)
		router.Handle("/children/{childId}", authenticator.Roles(handlerFactory.Delete(opts), shared.ROLE_OFFICE_MANAGER, shared.ROLE_ADMIN)).Methods(http.MethodDelete)

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

	Describe("CHILDREN", func() {

		Describe("LIST", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodGet
				httpEndpointToUse = "/children"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertReturnedChildrenWithIds("childid-1", "childid-2", "childid-3", "childid-4")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[shared.ROLE_OFFICE_MANAGER] = true })
				assertReturnedChildrenWithIds("childid-3", "childid-4")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[shared.ROLE_TEACHER] = true })
				assertReturnedChildrenWithIds("childid-3", "childid-4")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADULT] = true
					claims["userId"] = "id4"
				})
				assertReturnedChildrenWithIds("childid-3")
				assertHttpCode(http.StatusOK)
			})

			Context("When there are no children", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADULT] = true
					claims["userId"] = "foo"
				})
				assertJsonResponse(`[]`)
				assertHttpCode(http.StatusOK)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to list children: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

		})

		Describe("GET", func() {

			var (
				jsonChildRef = `{
				  "id": "childid-1",
				  "daycareId": "namek",
				  "firstName": "Goten",
				  "lastName": "Goten",
				  "birthDate": "1992-10-13 00:00:00 +0000 UTC",
				  "gender": "M",
				  "imageUri": "gs://foo/bar.jpg",
				  "startDate": "2018-03-28 00:00:00 +0000 UTC",
				  "notes": "some special notes",
				  "allergies": [
					{
					  "id": "allergyid-1",
					  "allergy": "tomato",
					  "instruction": "call the doctor"
					}
				  ],
				  "responsibleId": "id6",
				  "relationship": "father",
				  "specialInstructions": [
					{
					  "id": "specialinstruction-1",
					  "childId": "childid-1",
					  "instruction": "this boy always sleeps please keep him awaken"
					}
				  ]
				}`
			)

			BeforeEach(func() {
				httpMethodToUse = http.MethodGet
				httpEndpointToUse = "/children/childid-1"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertReturnedSingleChild(jsonChildRef)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager from the same daycare of the child", func() {
				BeforeEach(func() {
					claims[shared.ROLE_OFFICE_MANAGER] = true
					claims["daycareId"] = "namek"
				})
				assertReturnedSingleChild(jsonChildRef)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager from another daycare", func() {
				BeforeEach(func() {
					claims[shared.ROLE_OFFICE_MANAGER] = true
					claims["daycareId"] = "peyredragon"
				})
				assertJsonResponse(`{"error": "failed to get child: child not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

			// TODO: When https://github.com/Vinubaba/SANTC-API/issues/19 is done
			/*Context("When user is a teacher", func() {
				BeforeEach(func() { claims[shared.ROLE_TEACHER] = true })
				assertReturnedSingleChild(jsonChildRef)
				assertHttpCode(http.StatusOK)
			})*/

			Context("When user is an adult responsible", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADULT] = true
					claims["userId"] = "id6"
					claims["daycareId"] = "namek"
				})
				assertReturnedSingleChild(jsonChildRef)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a random adult", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADULT] = true
					claims["userId"] = "foo"
				})
				assertJsonResponse(`{"error": "failed to get child: child not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to get child: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the child does not exists", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					httpEndpointToUse = "/children/foo"
				})
				assertJsonResponse(`{"error":"failed to get child: child not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("DELETE", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodDelete
				httpEndpointToUse = "/children/childid-1"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusNoContent)
			})

			Context("When user is an office manager from a different daycare", func() {
				BeforeEach(func() {
					claims[shared.ROLE_OFFICE_MANAGER] = true
					claims["daycareId"] = "peyredragon"
				})
				assertJsonResponse(`{"error": "failed to delete child: child not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

			Context("When user is an office manager from the same daycare", func() {
				BeforeEach(func() {
					claims[shared.ROLE_OFFICE_MANAGER] = true
					claims["daycareId"] = "namek"
				})
				assertReturnedNoPayload()
				assertHttpCode(http.StatusNoContent)
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
				assertJsonResponse(`{"error":"failed to delete child: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the child does not exists", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					httpEndpointToUse = "/children/foo"
				})
				assertJsonResponse(`{"error":"failed to delete child: child not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("UPDATE", func() {

			var (
				jsonUpdatedChild = `
				{
				  "id": "childid-1",
				  "daycareId": "namek",
				  "firstName": "Rickon",
				  "lastName": "Stark",
				  "birthDate": "1992-10-14 00:00:00 +0000 UTC",
				  "gender": "M",
				  "imageUri": "gs://foo/bar.jpg",
				  "startDate": "2018-03-28 00:00:00 +0000 UTC",
				  "notes": "updated notes",
				  "allergies": [
					{
					  "id": "bbb",
					  "allergy": "tomato",
					  "instruction": "take him to the doctor"
					}
				  ],
				  "responsibleId": "id6",
				  "relationship": "mother",
				  "specialInstructions": [
					{
					  "id": "aaa",
					  "childId": "childid-1",
					  "instruction": "another special instruction"
					}
				  ]
				}`
			)

			BeforeEach(func() {
				mockStorage.On("Store", mock.Anything, mock.Anything, mock.Anything).Return(mockImageUriName, nil)
				httpMethodToUse = http.MethodPatch
				httpEndpointToUse = "/children/childid-1"
				httpBodyToUse = `{"notes": "updated notes", "specialInstructions": [{"instruction": "another special instruction"}], "relationship": "mother", "allergies": [{"allergy": "tomato", "instruction": "take him to the doctor"}], "responsibleId": "id6", "firstName": "Rickon", "lastName": "Stark", "gender": "M", "birthDate": "1992/10/14", "startDate":"2018/03/28", "imageUri": "data:image/jpeg;base64,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"}`

			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertReturnedSingleChild(jsonUpdatedChild)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an admin and tries to set responsible from another daycare", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					httpBodyToUse = `{"relationship": "mother", "responsibleId": "id2"}`
				})
				assertJsonResponse(`{"error": "failed to update child: cannot set responsible from a different daycare: failed to set responsible"}`)
				assertHttpCode(http.StatusBadRequest)
			})

			Context("When user is an office manager from another daycare", func() {
				BeforeEach(func() {
					claims[shared.ROLE_OFFICE_MANAGER] = true
					claims["daycareId"] = "peyredragon"
				})

				assertJsonResponse(`{"error": "failed to update child: child not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

			Context("When user is an office manager from the same daycare", func() {
				BeforeEach(func() {
					claims[shared.ROLE_OFFICE_MANAGER] = true
					claims["daycareId"] = "namek"
				})

				assertReturnedSingleChild(jsonUpdatedChild)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[shared.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult responsible", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADULT] = true
					claims["daycareId"] = "namek"
					claims["userId"] = "id6"
				})
				assertReturnedSingleChild(jsonUpdatedChild)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a random adult", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADULT] = true
					claims["daycareId"] = "namek"
					claims["userId"] = "foo"
				})
				assertJsonResponse(`{"error":"failed to update child: child not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to update child: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the child does not exists", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					httpEndpointToUse = "/children/foo"
				})
				assertJsonResponse(`{"error":"failed to update child: child not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("CREATE", func() {

			var (
				jsonCreatedChild = `{
				  "id": "aaa",
				  "daycareId": "peyredragon",
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
				}`
			)

			BeforeEach(func() {
				mockStorage.On("Store", mock.Anything, mock.Anything, mock.Anything).Return(mockImageUriName, nil)
				httpMethodToUse = http.MethodPost
				httpEndpointToUse = "/children"
				httpBodyToUse = `{"daycareId": "peyredragon", "notes": "he hates yogurt", "specialInstructions": [{"instruction": "vegetarian"}], "relationship": "mother", "allergies": [{"allergy": "tomato", "instruction": "take him to the doctor"}], "responsibleId": "id4", "firstName": "Rickon", "lastName": "Stark", "gender": "M", "birthDate": "1992/10/14", "startDate":"2018/03/28", "imageUri": "data:image/jpeg;base64,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"}`
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertReturnedSingleChild(jsonCreatedChild)
				assertHttpCode(http.StatusCreated)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[shared.ROLE_OFFICE_MANAGER] = true })
				assertReturnedSingleChild(jsonCreatedChild)
				assertHttpCode(http.StatusCreated)
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
				assertJsonResponse(`{"error":"failed to add child: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When responsibleId does not exists", func() {
				BeforeEach(func() {
					claims[shared.ROLE_OFFICE_MANAGER] = true
					httpBodyToUse = `{"relationship": "father", "responsibleId": "foo", "firstName": "Arthur", "lastName": "Gustin", "gender": "M", "birthDate": "1992/10/14", "startDate":"2018/03/28", "imageUri": "data:image/jpeg;base64,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"}`
				})
				assertJsonResponse(`{"error": "failed to add child: cannot set responsible from a different daycare: failed to set responsible"}`)
				assertHttpCode(http.StatusBadRequest)
			})

		})

	})

})
