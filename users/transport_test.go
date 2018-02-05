package users_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"

	. "github.com/DigitalFrameworksLLC/teddycare/shared/mocks"
	. "github.com/DigitalFrameworksLLC/teddycare/storage/mocks"
	"github.com/DigitalFrameworksLLC/teddycare/store"
	. "github.com/DigitalFrameworksLLC/teddycare/users"

	"encoding/json"
	"github.com/DigitalFrameworksLLC/teddycare/authentication"
	. "github.com/DigitalFrameworksLLC/teddycare/firebase/mocks"
	"github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"strings"
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

		mockImageUriName string
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
				Expect(recorder.Body.String()).To(MatchJSON(response))
			})
		}
	)

	BeforeEach(func() {
		var err error
		connectString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			"localhost",
			"5432",
			"postgres",
			"postgres",
			"test_teddycare")
		concreteDb, err = gorm.Open("postgres", connectString)
		if err != nil {
			panic(err)
		}
		concreteDb.LogMode(false)

		mockStringGenerator = &MockStringGenerator{}
		mockStringGenerator.On("GenerateUuid").Return("aaa").Once()

		mockStorage = &MockGcs{}
		mockImageUriName = "bar.jpg"
		mockStorage.On("Get", mock.Anything, mock.Anything).Return("gs://foo/"+mockImageUriName, nil)
		mockStorage.On("Delete", mock.Anything, mock.Anything).Return(nil)

		concreteStore = &store.Store{
			Db:              concreteDb,
			StringGenerator: mockStringGenerator,
		}

		mockFirebaseClient = &MockClient{}
		mockFirebaseClient.On("DeleteUser", mock.Anything, mock.Anything).Return(nil)

		userService := &UserService{
			FirebaseClient: mockFirebaseClient,
			Store:          concreteStore,
			Storage:        mockStorage,
		}

		authenticator = &authentication.Authenticator{
			UserService: userService,
		}

		httpMethodToUse = ""
		httpEndpointToUse = ""
		httpBodyToUse = ""

		router = mux.NewRouter()
		var logger log.Logger
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		httpLogger := log.With(logger, "component", "http")
		opts := []kithttp.ServerOption{
			kithttp.ServerErrorLogger(httpLogger),
			kithttp.ServerErrorEncoder(EncodeError),
		}

		handlerFactory := HandlerFactory{
			Service: userService,
		}

		router.Handle("/office-managers", authenticator.Roles(handlerFactory.ListOfficeManager(opts), store.ROLE_ADMIN)).Methods(http.MethodGet)
		router.Handle("/office-managers/{id}", authenticator.Roles(handlerFactory.GetOfficeManager(opts), store.ROLE_ADMIN)).Methods(http.MethodGet)
		router.Handle("/office-managers/{id}", authenticator.Roles(handlerFactory.DeleteOfficeManager(opts), store.ROLE_ADMIN)).Methods(http.MethodDelete)
		router.Handle("/office-managers/{id}", authenticator.Roles(handlerFactory.UpdateOfficeManager(opts), store.ROLE_ADMIN)).Methods(http.MethodPatch)

		router.Handle("/teachers", authenticator.Roles(handlerFactory.ListTeacher(opts), store.ROLE_ADMIN, store.ROLE_OFFICE_MANAGER)).Methods(http.MethodGet)
		router.Handle("/teachers/{id}", authenticator.Roles(handlerFactory.GetTeacher(opts), store.ROLE_ADMIN, store.ROLE_OFFICE_MANAGER)).Methods(http.MethodGet)
		router.Handle("/teachers/{id}", authenticator.Roles(handlerFactory.DeleteTeacher(opts), store.ROLE_ADMIN, store.ROLE_OFFICE_MANAGER)).Methods(http.MethodDelete)
		router.Handle("/teachers/{id}", authenticator.Roles(handlerFactory.UpdateTeacher(opts), store.ROLE_ADMIN, store.ROLE_OFFICE_MANAGER)).Methods(http.MethodPatch)

		router.Handle("/adults", authenticator.Roles(handlerFactory.ListAdult(opts), store.ROLE_ADMIN, store.ROLE_OFFICE_MANAGER)).Methods(http.MethodGet)
		router.Handle("/adults/{id}", authenticator.Roles(handlerFactory.GetAdult(opts), store.ROLE_ADMIN, store.ROLE_OFFICE_MANAGER)).Methods(http.MethodGet)
		router.Handle("/adults/{id}", authenticator.Roles(handlerFactory.DeleteAdult(opts), store.ROLE_ADMIN, store.ROLE_OFFICE_MANAGER)).Methods(http.MethodDelete)
		router.Handle("/adults/{id}", authenticator.Roles(handlerFactory.UpdateAdult(opts), store.ROLE_ADMIN, store.ROLE_OFFICE_MANAGER)).Methods(http.MethodPatch)

		recorder = httptest.NewRecorder()

		concreteStore.Db.Exec(`TRUNCATE TABLE "users" CASCADE`)
		concreteStore.Db.Exec(`TRUNCATE TABLE "roles" CASCADE`)
		concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id1','arthur@gmail.com','Arthur','Gustin','M','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com')`)
		concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id2','vinu@gmail.com','Vinu','Singh','M','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com')`)
		concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id3','john@gmail.com','John','John','M','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com')`)
		concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id4','estree@gmail.com','Estree','Delacour','F','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com')`)
		concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id5','anna@gmail.com','Anna','Melnychuk','F','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com')`)
		concreteStore.Db.Exec(`INSERT INTO "roles" ("user_id","role") VALUES ('id1', '` + store.ROLE_ADMIN + `')`)
		concreteStore.Db.Exec(`INSERT INTO "roles" ("user_id","role") VALUES ('id2', '` + store.ROLE_OFFICE_MANAGER + `')`)
		concreteStore.Db.Exec(`INSERT INTO "roles" ("user_id","role") VALUES ('id3', '` + store.ROLE_OFFICE_MANAGER + `')`)
		concreteStore.Db.Exec(`INSERT INTO "roles" ("user_id","role") VALUES ('id4', '` + store.ROLE_TEACHER + `')`)
		concreteStore.Db.Exec(`INSERT INTO "roles" ("user_id","role") VALUES ('id5', '` + store.ROLE_ADULT + `')`)
	})

	AfterEach(func() {
		concreteDb.Close()
	})

	BeforeEach(func() {
		claims = map[string]interface{}{
			"userId":                  "",
			store.ROLE_TEACHER:        false,
			store.ROLE_OFFICE_MANAGER: false,
			store.ROLE_ADULT:          false,
			store.ROLE_ADMIN:          false,
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
				BeforeEach(func() { claims[store.ROLE_ADMIN] = true })
				assertReturnedUsersWithIds("id5")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[store.ROLE_OFFICE_MANAGER] = true })
				assertReturnedUsersWithIds("id5")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[store.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[store.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
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
				BeforeEach(func() { claims[store.ROLE_ADMIN] = true })
				assertReturnedSingleUser(`{"id":"id5","firstName":"Anna","lastName":"Melnychuk","gender":"F","email":"anna@gmail.com","phone":"+3365651","address_1":"1 RUE TRUC","address_2":"APP 4","city":"Toulouse","state":"FRANCE","zip":"31400","imageUri":"http://image.com","roles":["adult"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[store.ROLE_OFFICE_MANAGER] = true })
				assertReturnedSingleUser(`{"id":"id5","firstName":"Anna","lastName":"Melnychuk","gender":"F","email":"anna@gmail.com","phone":"+3365651","address_1":"1 RUE TRUC","address_2":"APP 4","city":"Toulouse","state":"FRANCE","zip":"31400","imageUri":"http://image.com","roles":["adult"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[store.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[store.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to get user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
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
				BeforeEach(func() { claims[store.ROLE_ADMIN] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusNoContent)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[store.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusNoContent)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[store.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[store.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to delete user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
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
				BeforeEach(func() { claims[store.ROLE_ADMIN] = true })
				assertReturnedSingleUser(`{"id":"id5","firstName":"Anna","lastName":"Melnychuk","gender":"F","email":"anna@gmail.com","phone":"+3365651","address_1": "8 RUE PIERRE DELDI", "address_2": "VILLA 13","city":"Toulouse","state":"FRANCE","zip":"31400","imageUri":"gs://foo/bar.jpg","roles":["adult"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[store.ROLE_OFFICE_MANAGER] = true })
				assertReturnedSingleUser(`{"id":"id5","firstName":"Anna","lastName":"Melnychuk","gender":"F","email":"anna@gmail.com","phone":"+3365651","address_1": "8 RUE PIERRE DELDI", "address_2": "VILLA 13","city":"Toulouse","state":"FRANCE","zip":"31400","imageUri":"gs://foo/bar.jpg","roles":["adult"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[store.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[store.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to update user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
					httpEndpointToUse = "/adults/foo"
				})
				assertJsonResponse(`{"error":"failed to update user: user not found"}`)
				assertHttpCode(http.StatusNotFound)
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
				BeforeEach(func() { claims[store.ROLE_ADMIN] = true })
				assertReturnedUsersWithIds("id2", "id3")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[store.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[store.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[store.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
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
				BeforeEach(func() { claims[store.ROLE_ADMIN] = true })
				assertReturnedSingleUser(`{"id":"id2","firstName":"Vinu","lastName":"Singh","gender":"M","email":"vinu@gmail.com","phone":"+3365651","address_1":"1 RUE TRUC","address_2":"APP 4","city":"Toulouse","state":"FRANCE","zip":"31400","imageUri":"http://image.com","roles":["officemanager"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[store.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[store.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[store.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to get user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
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
				BeforeEach(func() { claims[store.ROLE_ADMIN] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusNoContent)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[store.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[store.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[store.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to delete user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
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
				BeforeEach(func() { claims[store.ROLE_ADMIN] = true })
				assertReturnedSingleUser(`{"id": "id2","firstName": "Vinu","lastName": "Singh","gender": "M","email": "vinu@gmail.com","phone": "+3365651","address_1": "8 RUE PIERRE DELDI","address_2": "VILLA 13","city": "Toulouse","state": "FRANCE","zip": "31400","imageUri": "gs://foo/bar.jpg","roles": ["officemanager"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[store.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[store.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[store.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to update user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
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
				BeforeEach(func() { claims[store.ROLE_ADMIN] = true })
				assertReturnedUsersWithIds("id4")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[store.ROLE_OFFICE_MANAGER] = true })
				assertReturnedUsersWithIds("id4")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[store.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[store.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
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
				BeforeEach(func() { claims[store.ROLE_ADMIN] = true })
				assertReturnedSingleUser(`{"id":"id4","firstName":"Estree","lastName":"Delacour","gender":"F","email":"estree@gmail.com","phone":"+3365651","address_1":"1 RUE TRUC","address_2":"APP 4","city":"Toulouse","state":"FRANCE","zip":"31400","imageUri":"http://image.com","roles":["teacher"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[store.ROLE_OFFICE_MANAGER] = true })
				assertReturnedSingleUser(`{"id":"id4","firstName":"Estree","lastName":"Delacour","gender":"F","email":"estree@gmail.com","phone":"+3365651","address_1":"1 RUE TRUC","address_2":"APP 4","city":"Toulouse","state":"FRANCE","zip":"31400","imageUri":"http://image.com","roles":["teacher"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[store.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[store.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to get user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
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
				BeforeEach(func() { claims[store.ROLE_ADMIN] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusNoContent)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[store.ROLE_OFFICE_MANAGER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusNoContent)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[store.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[store.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to delete user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
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
				BeforeEach(func() { claims[store.ROLE_ADMIN] = true })
				assertReturnedSingleUser(`{"id":"id4","firstName":"Estree","lastName":"Delacour","gender":"F","email":"estree@gmail.com","phone":"+3365651","address_1":"8 RUE PIERRE DELDI","address_2":"VILLA 13","city":"Toulouse","state":"FRANCE","zip":"31400","imageUri":"gs://foo/bar.jpg","roles":["teacher"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[store.ROLE_OFFICE_MANAGER] = true })
				assertReturnedSingleUser(`{"id":"id4","firstName":"Estree","lastName":"Delacour","gender":"F","email":"estree@gmail.com","phone":"+3365651","address_1":"8 RUE PIERRE DELDI","address_2":"VILLA 13","city":"Toulouse","state":"FRANCE","zip":"31400","imageUri":"gs://foo/bar.jpg","roles":["teacher"]}`)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is a teacher", func() {
				BeforeEach(func() { claims[store.ROLE_TEACHER] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When user is an adult", func() {
				BeforeEach(func() { claims[store.ROLE_ADULT] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusUnauthorized)
			})

			Context("When database is closed", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
					concreteDb.Close()
				})
				assertJsonResponse(`{"error":"failed to update user: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the user does not exists", func() {
				BeforeEach(func() {
					claims[store.ROLE_ADMIN] = true
					httpEndpointToUse = "/adults/foo"
				})
				assertJsonResponse(`{"error":"failed to update user: user not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

	})

})
