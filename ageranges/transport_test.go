package ageranges_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/Vinubaba/SANTC-API/ageranges"
	"github.com/Vinubaba/SANTC-API/authentication"
	. "github.com/Vinubaba/SANTC-API/firebase/mocks"
	"github.com/Vinubaba/SANTC-API/shared"
	. "github.com/Vinubaba/SANTC-API/shared/mocks"
	"github.com/Vinubaba/SANTC-API/store"
	"github.com/Vinubaba/SANTC-API/users"

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

		assertReturnedAgeRangeWithIds = func(ids ...string) {
			It(fmt.Sprintf("should respond %d ageRanges", len(ids)), func() {
				if len(ids) == 0 {
					panic("cant test with 0 id")
				}
				ageRangesTransport := []AgeRangeTransport{}
				json.Unmarshal([]byte(recorder.Body.String()), &ageRangesTransport)
				Expect(ageRangesTransport).To(HaveLen(len(ids)))

				returnedId := func(id string, response []AgeRangeTransport) bool {
					for _, r := range response {
						if r.Id == id {
							return true
						}
					}
					return false
				}

				for _, id := range ids {
					if !returnedId(id, ageRangesTransport) {
						Fail(fmt.Sprintf("%s was not found in response %s", id, ageRangesTransport))
					}
				}
			})
		}

		assertReturnedNoPayload = func() {
			It("should respond with 1 users", func() {
				Expect(recorder.Body.String()).To(Equal(""))
			})
		}

		assertReturnedSingleAgeRange = func(ageRangeJson string) {
			It("should respond with 1 ageRange", func() {
				Expect(recorder.Body.String()).To(MatchJSON(ageRangeJson))
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

		ageRangeService := &AgeRangeService{
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
			Service: ageRangeService,
		}

		router.Handle("/age-ranges", authenticator.Roles(handlerFactory.Add(opts), shared.ROLE_OFFICE_MANAGER, shared.ROLE_ADMIN)).Methods(http.MethodPost)
		router.Handle("/age-ranges", authenticator.Roles(handlerFactory.List(opts), shared.ROLE_OFFICE_MANAGER, shared.ROLE_ADMIN)).Methods(http.MethodGet)
		router.Handle("/age-ranges/{ageRangeId}", authenticator.Roles(handlerFactory.Get(opts), shared.ROLE_OFFICE_MANAGER, shared.ROLE_ADMIN)).Methods(http.MethodGet)
		router.Handle("/age-ranges/{ageRangeId}", authenticator.Roles(handlerFactory.Update(opts), shared.ROLE_OFFICE_MANAGER, shared.ROLE_ADMIN)).Methods(http.MethodPatch)
		router.Handle("/age-ranges/{ageRangeId}", authenticator.Roles(handlerFactory.Delete(opts), shared.ROLE_OFFICE_MANAGER, shared.ROLE_ADMIN)).Methods(http.MethodDelete)

		recorder = httptest.NewRecorder()

		concreteStore.Db.Exec(`TRUNCATE TABLE "users" CASCADE`)
		concreteStore.Db.Exec(`TRUNCATE TABLE "roles" CASCADE`)
		concreteStore.Db.Exec(`TRUNCATE TABLE "age_ranges" CASCADE`)
		concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id1','arthur@gmail.com','Arthur','Gustin','M','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com')`)
		concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id2','vinu@gmail.com','Vinu','Singh','M','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com')`)
		concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id3','john@gmail.com','John','John','M','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com')`)
		concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id4','estree@gmail.com','Estree','Delacour','F','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com')`)
		concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","first_name","last_name","gender","phone","address_1","address_2","city","state","zip","image_uri") VALUES ('id5','anna@gmail.com','Anna','Melnychuk','F','+3365651','1 RUE TRUC','APP 4','Toulouse','FRANCE','31400','http://image.com')`)
		concreteStore.Db.Exec(`INSERT INTO "roles" ("user_id","role") VALUES ('id1', '` + shared.ROLE_ADMIN + `')`)
		concreteStore.Db.Exec(`INSERT INTO "roles" ("user_id","role") VALUES ('id2', '` + shared.ROLE_OFFICE_MANAGER + `')`)
		concreteStore.Db.Exec(`INSERT INTO "roles" ("user_id","role") VALUES ('id3', '` + shared.ROLE_OFFICE_MANAGER + `')`)
		concreteStore.Db.Exec(`INSERT INTO "roles" ("user_id","role") VALUES ('id4', '` + shared.ROLE_TEACHER + `')`)
		concreteStore.Db.Exec(`INSERT INTO "roles" ("user_id","role") VALUES ('id5', '` + shared.ROLE_ADULT + `')`)

		concreteStore.Db.Exec(`INSERT INTO "age_ranges" ("age_range_id","stage","min","min_unit","max","max_unit") VALUES ('agerangeid-1','infant','3','M','12','M')`)
		concreteStore.Db.Exec(`INSERT INTO "age_ranges" ("age_range_id","stage","min","min_unit","max","max_unit") VALUES ('agerangeid-2','toddlers','12','M','18','M')`)
	})

	AfterEach(func() {
		concreteDb.Close()
	})

	BeforeEach(func() {
		claims = map[string]interface{}{
			"userId":                   "",
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

	Describe("AGE RANGES", func() {

		Describe("LIST", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodGet
				httpEndpointToUse = "/age-ranges"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertReturnedAgeRangeWithIds("agerangeid-1", "agerangeid-2")
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[shared.ROLE_OFFICE_MANAGER] = true })
				assertReturnedAgeRangeWithIds("agerangeid-1", "agerangeid-2")
				assertHttpCode(http.StatusOK)
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
				assertJsonResponse(`{"error":"failed to list age ranges: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

		})

		Describe("GET", func() {

			var (
				jsonClassRef = `{
					"id": "agerangeid-1",
					"stage": "infant",
					"min": 3,
					"minUnit": "M",
					"max": 12,
					"maxUnit": "M"
				}`
			)

			BeforeEach(func() {
				httpMethodToUse = http.MethodGet
				httpEndpointToUse = "/age-ranges/agerangeid-1"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertReturnedSingleAgeRange(jsonClassRef)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[shared.ROLE_OFFICE_MANAGER] = true })
				assertReturnedSingleAgeRange(jsonClassRef)
				assertHttpCode(http.StatusOK)
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
				assertJsonResponse(`{"error":"failed to get age range: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the ageRange does not exists", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					httpEndpointToUse = "/age-ranges/foo"
				})
				assertJsonResponse(`{"error":"failed to get age range: age range not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("DELETE", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodDelete
				httpEndpointToUse = "/age-ranges/agerangeid-1"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusNoContent)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[shared.ROLE_OFFICE_MANAGER] = true })
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
				assertJsonResponse(`{"error":"failed to delete age range: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the age range does not exists", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					httpEndpointToUse = "/age-ranges/foo"
				})
				assertJsonResponse(`{"error":"failed to delete age range: age range not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("UPDATE", func() {

			var (
				jsonUpdatedAgeRange = `{
					"id": "agerangeid-1",
					"stage": "updated infant",
					"min": 2,
					"minUnit": "Y",
					"max": 3,
					"maxUnit": "Y"
				}`
			)

			BeforeEach(func() {
				httpMethodToUse = http.MethodPatch
				httpEndpointToUse = "/age-ranges/agerangeid-1"
				httpBodyToUse = `{"stage": "updated infant","min": 2,"minUnit": "Y","max": 3,"maxUnit": "Y"}`
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertReturnedSingleAgeRange(jsonUpdatedAgeRange)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[shared.ROLE_OFFICE_MANAGER] = true })
				assertReturnedSingleAgeRange(jsonUpdatedAgeRange)
				assertHttpCode(http.StatusOK)
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
				assertJsonResponse(`{"error":"failed to update age range: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the ageRange does not exists", func() {
				BeforeEach(func() {
					claims[shared.ROLE_ADMIN] = true
					httpEndpointToUse = "/age-ranges/foo"
				})
				assertJsonResponse(`{"error":"failed to update age range: age range not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("CREATE", func() {

			var (
				jsonCreatedAgeRange = `{
					"id": "aaa",
					"stage": "created infant",
					"min": 2,
					"minUnit": "Y",
					"max": 3,
					"maxUnit": "Y"
				}`
			)

			BeforeEach(func() {
				httpMethodToUse = http.MethodPost
				httpEndpointToUse = "/age-ranges"
				httpBodyToUse = `{"stage": "created infant","min": 2,"minUnit": "Y","max": 3,"maxUnit": "Y"}`
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[shared.ROLE_ADMIN] = true })
				assertReturnedSingleAgeRange(jsonCreatedAgeRange)
				assertHttpCode(http.StatusCreated)
			})

			Context("When user is an office manager", func() {
				BeforeEach(func() { claims[shared.ROLE_OFFICE_MANAGER] = true })
				assertReturnedSingleAgeRange(jsonCreatedAgeRange)
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
				assertJsonResponse(`{"error":"failed to add age range: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

		})

	})

})
