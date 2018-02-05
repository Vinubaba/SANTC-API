package adult_responsible_test

import (
	"context"
	"fmt"
	"log"
	"os/exec"

	. "github.com/DigitalFrameworksLLC/teddycare/adult_responsible"
	"github.com/DigitalFrameworksLLC/teddycare/shared/mocks"
	"github.com/DigitalFrameworksLLC/teddycare/store"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("Service", func() {

	var (
		ctx                     = context.Background()
		adultResponsibleService Service
		mockStringGenerator     *shared.MockStringGenerator
		concreteStore           *store.Store
		concreteDb              *gorm.DB
		returnedError, err      error
		createdAdult            store.AdultResponsible
		adultRef1, adultRef2    AdultResponsibleTransport
	)

	BeforeEach(func() {
		adultRef1 = AdultResponsibleTransport{
			Id:        "aaa",
			FirstName: "Arthur",
			LastName:  "Gustin",
			Gender:    "M",
			Email:     "arthur.gustin@gmail.com",
			Password:  "azerty",
			Zip:       "31",
			State:     "FRANCE",
			Phone:     "0633326825",
			City:      "Toulouse",
			Addres_1:  "11, rue hergé",
			Addres_2:  "app 8",
		}
		adultRef2 = AdultResponsibleTransport{
			Id:        "bbb",
			FirstName: "Patrick",
			LastName:  "Gustin",
			Gender:    "M",
			Email:     "patrick.gustin@gmail.com",
			Password:  "azerty",
			Zip:       "66",
			State:     "FRANCE",
			Phone:     "65412357",
			City:      "Le Perthus",
			Addres_1:  "84 avenue de france",
			Addres_2:  "",
		}
	})

	var (
		assertNoError = func() {
			It("should not return an error", func() {
				Expect(returnedError).To(BeNil())
			})
		}
		assertErrorWithCause = func(cause error) {
			It("should return an error", func() {
				Expect(returnedError).NotTo(BeNil())
				Expect(errors.Cause(returnedError)).To(Equal(cause))
			})
		}
	)

	BeforeSuite(func() {
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
		concreteDb.LogMode(true)
	})

	AfterSuite(func() {
		concreteDb.Close()
	})

	BeforeEach(func() {
		mockStringGenerator = &shared.MockStringGenerator{}
		concreteStore = &store.Store{
			Db:              concreteDb,
			StringGenerator: mockStringGenerator,
		}
		adultResponsibleService = &AdultResponsibleService{
			Store: concreteStore,
		}

		mockStringGenerator.On("GenerateUuid").Return("aaa").Once()
		mockStringGenerator.On("GenerateUuid").Return("bbb").Once()
		mockStringGenerator.On("GenerateUuid").Return("ccc").Once()
		mockStringGenerator.On("GenerateUuid").Return("ddd").Once()
	})

	AfterEach(func() {
		if err := exec.Command("psql", "-U", "postgres", "-h", "localhost", "-d", "test_teddycare", "-c", "truncate table users cascade").Run(); err != nil {
			log.Fatal("failed to truncate table:" + err.Error())
		}
	})

	Context("AddAdultResponsible", func() {

		var (
			assertCreatedTheRightUser = func() {
				It("should create the right user in database", func() {
					Expect(createdAdult).NotTo(BeZero())
					Expect(createdAdult.ResponsibleId).To(Equal("aaa"))
				})
			}
			assertUserNotCreated = func() {
				It("should not create the user in database", func() {
					Expect(createdAdult).To(BeZero())
				})
			}
		)

		JustBeforeEach(func() {
			createdAdult, returnedError = adultResponsibleService.AddAdultResponsible(ctx, adultRef1)
		})

		Context("default", func() {
			BeforeEach(func() { adultRef1.Id = "" })
			assertCreatedTheRightUser()
			assertNoError()
		})

		Context("when the email is invalid", func() {
			BeforeEach(func() { adultRef1.Email = "arthur.gustingmail.com" })
			assertErrorWithCause(ErrInvalidEmail)
			assertUserNotCreated()
		})

		Context("when the password is too short", func() {
			BeforeEach(func() { adultRef1.Password = "123" })
			assertErrorWithCause(ErrInvalidPasswordFormat)
			assertUserNotCreated()
		})
	})

	Context("ListAdultResponsible", func() {

		var (
			allAdultsResponsible []store.AdultResponsible
		)

		var (
			assertReturnedTheRightNumberOfUsers = func(num int) {
				It("should return the right number of users", func() {
					Expect(allAdultsResponsible).To(HaveLen(num))
				})
			}
		)

		JustBeforeEach(func() {
			allAdultsResponsible, returnedError = adultResponsibleService.ListAdultResponsible(ctx)
		})

		Context("default", func() {
			BeforeEach(func() {
				adultResponsibleService.AddAdultResponsible(ctx, adultRef1)
				adultResponsibleService.AddAdultResponsible(ctx, adultRef2)
			})
			assertNoError()
			assertReturnedTheRightNumberOfUsers(2)
		})
	})

	Context("UpdateAdultResponsible", func() {

		var (
			updatedAdult  store.AdultResponsible
			adultToUpdate AdultResponsibleTransport
		)

		BeforeEach(func() {
			adultResponsibleService.AddAdultResponsible(ctx, adultRef1)
			adultResponsibleService.AddAdultResponsible(ctx, adultRef2)
		})

		JustBeforeEach(func() {
			updatedAdult, returnedError = adultResponsibleService.UpdateAdultResponsible(ctx, adultToUpdate)
		})

		Context("default", func() {

			BeforeEach(func() {
				adultToUpdate = AdultResponsibleTransport{
					Id:        "aaa",
					FirstName: "john",
					LastName:  "doe",
					Gender:    "F",
					Email:     "jonhdoe@gmail.com",
					Addres_1:  "ad1",
					Addres_2:  "ad2",
					City:      "Ibiza",
					Phone:     "6465",
					State:     "Iowa",
					Zip:       "6598",
					Password:  "123465789",
				}
			})

			It("should update all fields but the password and the id", func() {
				Expect(updatedAdult).To(Equal(store.AdultResponsible{
					ResponsibleId: "aaa",
					FirstName:     "john",
					LastName:      "doe",
					Gender:        "F",
					Email:         "jonhdoe@gmail.com",
					Addres_1:      "ad1",
					Addres_2:      "ad2",
					City:          "Ibiza",
					Phone:         "6465",
					State:         "Iowa",
					Zip:           "6598",
				}))
			})
			assertNoError()
		})

		Context("when the email is invalid", func() {
			BeforeEach(func() {
				adultToUpdate = AdultResponsibleTransport{
					Id:    "aaa",
					Email: "jonhdoegmail.com",
				}
			})
			assertErrorWithCause(ErrInvalidEmail)
		})
	})

	Context("GetAdultResponsible", func() {

		var (
			returnedAdult store.AdultResponsible
			request       AdultResponsibleTransport
		)

		JustBeforeEach(func() {
			returnedAdult, returnedError = adultResponsibleService.GetAdultResponsible(ctx, request)
		})

		BeforeEach(func() {
			_, e1 := adultResponsibleService.AddAdultResponsible(ctx, adultRef1)
			if e1 != nil {
				panic(e1.Error())
			}
			_, e2 := adultResponsibleService.AddAdultResponsible(ctx, adultRef2)
			if e2 != nil {
				panic(e2.Error())
			}
		})

		Context("default", func() {
			BeforeEach(func() { request = AdultResponsibleTransport{Id: "aaa"} })

			It("should return an user", func() {
				Expect(returnedAdult).To(Equal(store.AdultResponsible{
					ResponsibleId: "aaa",
					FirstName:     "Arthur",
					LastName:      "Gustin",
					Gender:        "M",
					Email:         "arthur.gustin@gmail.com",
					Zip:           "31",
					State:         "FRANCE",
					Phone:         "0633326825",
					City:          "Toulouse",
					Addres_1:      "11, rue hergé",
					Addres_2:      "app 8",
				}))
			})
			assertNoError()
		})

		Context("when the user does not exists", func() {
			BeforeEach(func() { request = AdultResponsibleTransport{Id: "rgrgdfgb"} })
			assertErrorWithCause(store.ErrUserNotFound)
		})
	})

	Context("DeleteAdultResponsible", func() {

		var (
			request AdultResponsibleTransport
		)

		JustBeforeEach(func() {
			returnedError = adultResponsibleService.DeleteAdultResponsible(ctx, request)
		})

		BeforeEach(func() {
			adultResponsibleService.AddAdultResponsible(ctx, adultRef1)
			adultResponsibleService.AddAdultResponsible(ctx, adultRef2)

			Context("default", func() {
				BeforeEach(func() { request = AdultResponsibleTransport{Id: "aaa"} })
				assertNoError()
			})

			Context("when the user does not exists", func() {
				BeforeEach(func() { request = AdultResponsibleTransport{Id: "rgrgdfgb"} })
				assertErrorWithCause(store.ErrUserNotFound)
			})
		})
	})
})
