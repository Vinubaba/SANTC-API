package schedules_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Vinubaba/SANTC-API/api/authentication"
	. "github.com/Vinubaba/SANTC-API/api/schedules"
	"github.com/Vinubaba/SANTC-API/api/shared"
	. "github.com/Vinubaba/SANTC-API/api/shared/mocks"
	"github.com/Vinubaba/SANTC-API/api/users"
	. "github.com/Vinubaba/SANTC-API/common/firebase/mocks"
	"github.com/Vinubaba/SANTC-API/common/store"

	"github.com/Vinubaba/SANTC-API/common/log"
	"github.com/Vinubaba/SANTC-API/common/roles"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

		assertReturnedNoPayload = func() {
			It("should respond with 1 users", func() {
				Expect(recorder.Body.String()).To(Equal(""))
			})
		}

		assertReturnedSingleSchedules = func(ageRangeJson string) {
			It("should respond with 1 ageRange", func() {
				Expect(recorder.Body.String()).To(MatchJSON(ageRangeJson))
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
		concreteDb = shared.NewDbInstance(false)
		concreteDb.LogMode(true)

		mockStringGenerator = &MockStringGenerator{}
		mockStringGenerator.On("GenerateUuid").Return("aaa").Once()
		mockStringGenerator.On("GenerateUuid").Return("bbb").Once()

		concreteStore = &store.Store{
			Db:              concreteDb,
			StringGenerator: mockStringGenerator,
		}

		userService := &users.UserService{
			FirebaseClient: mockFirebaseClient,
			Store:          concreteStore,
		}
		logger := log.NewLogger("teddycare")

		authenticator = &authentication.Authenticator{
			UserService: userService,
			Logger:      logger,
		}

		ageRangeService := &ScheduleService{
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

		router.Handle("/children/{childId}/schedules", authenticator.Roles(handlerFactory.Add(opts), roles.ROLE_OFFICE_MANAGER, roles.ROLE_ADMIN)).Methods(http.MethodPost)
		router.Handle("/children/{childId}/schedules/{scheduleId}", authenticator.Roles(handlerFactory.Get(opts), roles.ROLE_OFFICE_MANAGER, roles.ROLE_ADMIN)).Methods(http.MethodGet)
		router.Handle("/children/{childId}/schedules/{scheduleId}", authenticator.Roles(handlerFactory.Update(opts), roles.ROLE_OFFICE_MANAGER, roles.ROLE_ADMIN)).Methods(http.MethodPatch)
		router.Handle("/children/{childId}/schedules/{scheduleId}", authenticator.Roles(handlerFactory.Delete(opts), roles.ROLE_OFFICE_MANAGER, roles.ROLE_ADMIN)).Methods(http.MethodDelete)

		router.Handle("/teachers/{teacherId}/schedules", authenticator.Roles(handlerFactory.Add(opts), roles.ROLE_OFFICE_MANAGER, roles.ROLE_ADMIN)).Methods(http.MethodPost)
		router.Handle("/teachers/{teacherId}/schedules/{scheduleId}", authenticator.Roles(handlerFactory.Get(opts), roles.ROLE_OFFICE_MANAGER, roles.ROLE_ADMIN)).Methods(http.MethodGet)
		router.Handle("/teachers/{teacherId}/schedules/{scheduleId}", authenticator.Roles(handlerFactory.Update(opts), roles.ROLE_OFFICE_MANAGER, roles.ROLE_ADMIN)).Methods(http.MethodPatch)
		router.Handle("/teachers/{teacherId}/schedules/{scheduleId}", authenticator.Roles(handlerFactory.Delete(opts), roles.ROLE_OFFICE_MANAGER, roles.ROLE_ADMIN)).Methods(http.MethodDelete)

		recorder = httptest.NewRecorder()

		shared.SetDbInitialState()
	})

	AfterEach(func() {
		concreteDb.Close()
	})

	BeforeEach(func() {
		claims = map[string]interface{}{
			"userId":                  "",
			"daycareId":               "namek",
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

	Describe("SCHEDULES", func() {

		Describe("GET", func() {

			var (
				expectedJsonSchedule = `{
				  "id": "scheduleid-1",
				  "walkIn": false,
				  "mondayStart": "8:30 AM",
				  "mondayEnd": "6:00 PM",
				  "tuesdayStart": "8:30 AM",
				  "tuesdayEnd": "6:00 PM",
				  "wednesdayStart": "8:30 AM",
				  "wednesdayEnd": "6:00 PM",
				  "thursdayStart": "8:30 AM",
				  "thursdayEnd": "6:00 PM",
				  "fridayStart": "8:30 AM",
				  "fridayEnd": "6:00 PM",
				  "saturdayStart": "",
				  "saturdayEnd": "",
				  "sundayStart": "",
				  "sundayEnd": ""
				}`
			)

			BeforeEach(func() {
				httpMethodToUse = http.MethodGet
				httpEndpointToUse = "/children/childid-1/schedules/scheduleid-1"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedSingleSchedules(expectedJsonSchedule)
				assertHttpCode(http.StatusOK)
			})

			Context("When user is an office manager from peydragon", func() {
				BeforeEach(func() {
					claims[roles.ROLE_OFFICE_MANAGER] = true
					claims["daycareId"] = "peyredragon"
				})
				assertReturnedSingleSchedules(`{"error": "failed to get schedule: child not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

			Context("When user is an office manager from namek", func() {
				BeforeEach(func() {
					claims[roles.ROLE_OFFICE_MANAGER] = true
					claims["daycareId"] = "namek"
				})
				assertReturnedSingleSchedules(expectedJsonSchedule)
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
				assertJsonResponse(`{"error":"failed to get schedule: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the schedule does not exists", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					httpEndpointToUse = "/children/childid-1/schedules/foo"
				})
				assertJsonResponse(`{"error":"failed to get schedule: schedule not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("DELETE", func() {

			BeforeEach(func() {
				httpMethodToUse = http.MethodDelete
				httpEndpointToUse = "/children/childid-1/schedules/scheduleid-1"
			})

			Context("When user is an admin", func() {
				BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
				assertReturnedNoPayload()
				assertHttpCode(http.StatusNoContent)
			})

			Context("When user is an office manager from peyredragon", func() {
				BeforeEach(func() {
					claims[roles.ROLE_OFFICE_MANAGER] = true
					claims["daycareId"] = "peyredragon"
				})
				assertJsonResponse(`{"error": "failed to delete schedule: child not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

			Context("When user is an office manager from namek", func() {
				BeforeEach(func() {
					claims[roles.ROLE_OFFICE_MANAGER] = true
					claims["daycareId"] = "namek"
				})
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
				assertJsonResponse(`{"error":"failed to delete schedule: sql: database is closed"}`)
				assertHttpCode(http.StatusInternalServerError)
			})

			Context("When the schedule does not exists", func() {
				BeforeEach(func() {
					claims[roles.ROLE_ADMIN] = true
					httpEndpointToUse = "/children/childid-1/schedules/foo"
				})
				assertJsonResponse(`{"error":"failed to delete schedule: schedule not found"}`)
				assertHttpCode(http.StatusNotFound)
			})

		})

		Describe("UPDATE", func() {

			var (
				jsonUpdatedSchedules = `{
				  "id": "scheduleid-1",
                  "walkIn": false,
				  "mondayStart": "9:30 AM",
				  "mondayEnd": "7:00 PM",
				  "tuesdayStart": "8:30 AM",
				  "tuesdayEnd": "6:00 PM",
				  "wednesdayStart": "8:30 AM",
				  "wednesdayEnd": "6:00 PM",
				  "thursdayStart": "8:30 AM",
				  "thursdayEnd": "6:00 PM",
				  "fridayStart": "8:30 AM",
				  "fridayEnd": "6:00 PM",
				  "saturdayStart": "",
				  "saturdayEnd": "",
				  "sundayStart": "",
				  "sundayEnd": ""
				}`
			)

			BeforeEach(func() {
				httpMethodToUse = http.MethodPatch
				httpBodyToUse = `{
				  "mondayStart": "9:30 AM",
				  "mondayEnd": "7:00 PM"
				}`
			})

			Context("UPDATE TEACHER SCHEDULE", func() {

				// todo: on success, test that scheduleId is updated in users and children tables

				BeforeEach(func() {
					httpEndpointToUse = "/teachers/id9/schedules/scheduleid-1"
				})

				Context("When user is an admin", func() {
					BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
					assertReturnedSingleSchedules(jsonUpdatedSchedules)
					assertHttpCode(http.StatusOK)
				})

				Context("When user is an office manager of the same daycare", func() {
					BeforeEach(func() {
						claims[roles.ROLE_OFFICE_MANAGER] = true
					})
					assertReturnedSingleSchedules(jsonUpdatedSchedules)
					assertHttpCode(http.StatusOK)
				})

				Context("When user is an office manager of a different daycare", func() {
					BeforeEach(func() {
						claims[roles.ROLE_OFFICE_MANAGER] = true
						claims["daycareId"] = "peyredragon"
					})
					assertJsonResponse(`{"error": "failed to get teacher: user not found"}`)
					assertHttpCode(http.StatusNotFound)
				})

				Context("When the time is not well formatted", func() {
					BeforeEach(func() {
						claims[roles.ROLE_OFFICE_MANAGER] = true
						httpBodyToUse = `{
						  "mondayStart": "8:30"
						}`
					})
					assertJsonResponse(`{"error": "failed to validate request: 8:30 does not match regex: time does not match the following regex: ^\\d{1,2}:\\d{2}\\s(AM|PM)$"}`)
					assertHttpCode(http.StatusBadRequest)
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
					assertJsonResponse(`{"error":"failed to get teacher: sql: database is closed"}`)
					assertHttpCode(http.StatusInternalServerError)
				})

			})

			Context("UPDATE CHILD SCHEDULE", func() {
				BeforeEach(func() {
					httpEndpointToUse = "/children/childid-1/schedules/scheduleid-1"
				})

				Context("When user is an admin", func() {
					BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
					assertReturnedSingleSchedules(jsonUpdatedSchedules)
					assertHttpCode(http.StatusOK)
				})

				Context("When user is an office manager of the same daycare", func() {
					BeforeEach(func() {
						claims[roles.ROLE_OFFICE_MANAGER] = true
					})
					assertReturnedSingleSchedules(jsonUpdatedSchedules)
					assertHttpCode(http.StatusOK)
				})

				Context("When user is an office manager of a different daycare", func() {
					BeforeEach(func() {
						claims[roles.ROLE_OFFICE_MANAGER] = true
						claims["daycareId"] = "peyredragon"
					})
					assertJsonResponse(`{"error": "failed to get child: child not found"}`)
					assertHttpCode(http.StatusNotFound)
				})

				Context("When the time is not well formatted", func() {
					BeforeEach(func() {
						claims[roles.ROLE_OFFICE_MANAGER] = true
						httpBodyToUse = `{
						  "mondayStart": "8:30"
						}`
					})
					assertJsonResponse(`{"error": "failed to validate request: 8:30 does not match regex: time does not match the following regex: ^\\d{1,2}:\\d{2}\\s(AM|PM)$"}`)
					assertHttpCode(http.StatusBadRequest)
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
					assertJsonResponse(`{"error":"failed to get child: sql: database is closed"}`)
					assertHttpCode(http.StatusInternalServerError)
				})

			})

		})

		Describe("CREATE", func() {

			var (
				jsonCreatedSchedules = `{
				  "id": "aaa",
				  "walkIn": false,
				  "mondayStart": "8:30 AM",
				  "mondayEnd": "6:00 PM",
				  "tuesdayStart": "8:30 AM",
				  "tuesdayEnd": "6:00 PM",
				  "wednesdayStart": "8:30 AM",
				  "wednesdayEnd": "6:00 PM",
				  "thursdayStart": "8:30 AM",
				  "thursdayEnd": "6:00 PM",
				  "fridayStart": "8:30 AM",
				  "fridayEnd": "6:00 PM",
				  "saturdayStart": "",
				  "saturdayEnd": "",
				  "sundayStart": "",
				  "sundayEnd": ""
				}`
			)

			BeforeEach(func() {
				httpMethodToUse = http.MethodPost
				httpBodyToUse = `{
				  "walkIn": false,
				  "mondayStart": "8:30 AM",
				  "mondayEnd": "6:00 PM",
				  "tuesdayStart": "8:30 AM",
				  "tuesdayEnd": "6:00 PM",
				  "wednesdayStart": "8:30 AM",
				  "wednesdayEnd": "6:00 PM",
				  "thursdayStart": "8:30 AM",
				  "thursdayEnd": "6:00 PM",
				  "fridayStart": "8:30 AM",
				  "fridayEnd": "6:00 PM"
				}`
			})

			Context("CREATE TEACHER SCHEDULE", func() {

				// todo: on success, test that scheduleId is updated in users and children tables

				BeforeEach(func() {
					httpEndpointToUse = "/teachers/id9/schedules"
				})

				Context("When user is an admin", func() {
					BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
					assertReturnedSingleSchedules(jsonCreatedSchedules)
					assertHttpCode(http.StatusCreated)
				})

				Context("When user is an office manager of the same daycare", func() {
					BeforeEach(func() {
						claims[roles.ROLE_OFFICE_MANAGER] = true
					})
					assertReturnedSingleSchedules(jsonCreatedSchedules)
					assertHttpCode(http.StatusCreated)
				})

				Context("When user is an office manager of a different daycare", func() {
					BeforeEach(func() {
						claims[roles.ROLE_OFFICE_MANAGER] = true
						claims["daycareId"] = "peyredragon"
					})
					assertJsonResponse(`{"error": "failed to get teacher: user not found"}`)
					assertHttpCode(http.StatusNotFound)
				})

				Context("When the time are not well formatted", func() {
					BeforeEach(func() {
						claims[roles.ROLE_OFFICE_MANAGER] = true
						claims["daycareId"] = "peyredragon"
						httpBodyToUse = `{
						  "walkIn": false,
						  "mondayStart": "8:30",
						  "mondayEnd": "06:00",
						  "tuesdayStart": "8:30",
						  "tuesdayEnd": "06:00",
						  "wednesdayStart": "8:30",
						  "wednesdayEnd": "06:00",
						  "thursdayStart": "8:30",
						  "thursdayEnd": "06:00",
						  "fridayStart": "8:30",
						  "fridayEnd": "06:00"
						}`
					})
					assertJsonResponse(`{"error": "failed to validate request: 8:30 does not match regex: time does not match the following regex: ^\\d{1,2}:\\d{2}\\s(AM|PM)$"}`)
					assertHttpCode(http.StatusBadRequest)
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
					assertJsonResponse(`{"error":"failed to get teacher: sql: database is closed"}`)
					assertHttpCode(http.StatusInternalServerError)
				})

			})

			Context("CREATE CHILD SCHEDULE", func() {
				BeforeEach(func() {
					httpEndpointToUse = "/children/childid-1/schedules"
				})

				Context("When user is an admin", func() {
					BeforeEach(func() { claims[roles.ROLE_ADMIN] = true })
					assertReturnedSingleSchedules(jsonCreatedSchedules)
					assertHttpCode(http.StatusCreated)
				})

				Context("When user is an office manager of the same daycare", func() {
					BeforeEach(func() {
						claims[roles.ROLE_OFFICE_MANAGER] = true
					})
					assertReturnedSingleSchedules(jsonCreatedSchedules)
					assertHttpCode(http.StatusCreated)
				})

				Context("When user is an office manager of a different daycare", func() {
					BeforeEach(func() {
						claims[roles.ROLE_OFFICE_MANAGER] = true
						claims["daycareId"] = "peyredragon"
					})
					assertJsonResponse(`{"error": "failed to get child: child not found"}`)
					assertHttpCode(http.StatusNotFound)
				})

				Context("When the time are not well formatted", func() {
					BeforeEach(func() {
						claims[roles.ROLE_OFFICE_MANAGER] = true
						claims["daycareId"] = "peyredragon"
						httpBodyToUse = `{
						  "walkIn": false,
						  "mondayStart": "8:30",
						  "mondayEnd": "06:00",
						  "tuesdayStart": "8:30",
						  "tuesdayEnd": "06:00",
						  "wednesdayStart": "8:30",
						  "wednesdayEnd": "06:00",
						  "thursdayStart": "8:30",
						  "thursdayEnd": "06:00",
						  "fridayStart": "8:30",
						  "fridayEnd": "06:00"
						}`
					})
					assertJsonResponse(`{"error": "failed to validate request: 8:30 does not match regex: time does not match the following regex: ^\\d{1,2}:\\d{2}\\s(AM|PM)$"}`)
					assertHttpCode(http.StatusBadRequest)
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
					assertJsonResponse(`{"error":"failed to get child: sql: database is closed"}`)
					assertHttpCode(http.StatusInternalServerError)
				})

			})

		})

	})

})
