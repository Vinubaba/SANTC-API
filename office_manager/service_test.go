package office_manager_test

import (
	. "arthurgustin.fr/teddycare/office_manager"

	"arthurgustin.fr/teddycare/shared/mocks"
	"arthurgustin.fr/teddycare/store"
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"log"
	"os/exec"
)

var _ = Describe("Service", func() {

	var (
		ctx                                  = context.Background()
		officeManagerResponsibleService      Service
		mockStringGenerator                  *shared.MockStringGenerator
		concreteStore                        *store.Store
		concreteDb                           *gorm.DB
		returnedError, err                   error
		createdAdult                         store.OfficeManager
		officeManagerRef1, officeManagerRef2 OfficeManagerTransport
	)

	BeforeEach(func() {
		officeManagerRef1 = OfficeManagerTransport{
			Id:       "aaa",
			Email:    "arthur.gustin@gmail.com",
			Password: "azerty",
		}
		officeManagerRef2 = OfficeManagerTransport{
			Id:       "bbb",
			Email:    "patrick.gustin@gmail.com",
			Password: "azerty",
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
		officeManagerResponsibleService = &OfficeManagerService{
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

	Context("AddOfficeManager", func() {

		var (
			assertCreatedTheRightUser = func() {
				It("should create the right user in database", func() {
					Expect(createdAdult).NotTo(BeZero())
					Expect(createdAdult.OfficeManagerId).To(Equal("aaa"))
				})
			}
			assertUserNotCreated = func() {
				It("should not create the user in database", func() {
					Expect(createdAdult).To(BeZero())
				})
			}
		)

		JustBeforeEach(func() {
			createdAdult, returnedError = officeManagerResponsibleService.AddOfficeManager(ctx, officeManagerRef1)
		})

		Context("default", func() {
			BeforeEach(func() { officeManagerRef1.Id = "" })
			assertCreatedTheRightUser()
			assertNoError()
		})

		Context("when the email is invalid", func() {
			BeforeEach(func() { officeManagerRef1.Email = "arthur.gustingmail.com" })
			assertErrorWithCause(ErrInvalidEmail)
			assertUserNotCreated()
		})

		Context("when the password is too short", func() {
			BeforeEach(func() { officeManagerRef1.Password = "123" })
			assertErrorWithCause(ErrInvalidPasswordFormat)
			assertUserNotCreated()
		})
	})

	Context("ListOfficeManager", func() {

		var (
			allAdultsResponsible []store.OfficeManager
		)

		var (
			assertReturnedTheRightNumberOfUsers = func(num int) {
				It("should return the right number of users", func() {
					Expect(allAdultsResponsible).To(HaveLen(num))
				})
			}
		)

		JustBeforeEach(func() {
			allAdultsResponsible, returnedError = officeManagerResponsibleService.ListOfficeManager(ctx)
		})

		Context("default", func() {
			BeforeEach(func() {
				officeManagerResponsibleService.AddOfficeManager(ctx, officeManagerRef1)
				officeManagerResponsibleService.AddOfficeManager(ctx, officeManagerRef2)
			})
			assertNoError()
			assertReturnedTheRightNumberOfUsers(2)
		})
	})

	Context("UpdateOfficeManager", func() {

		var (
			updatedAdult          store.OfficeManager
			officeManagerToUpdate OfficeManagerTransport
		)

		BeforeEach(func() {
			officeManagerResponsibleService.AddOfficeManager(ctx, officeManagerRef1)
			officeManagerResponsibleService.AddOfficeManager(ctx, officeManagerRef2)
		})

		JustBeforeEach(func() {
			updatedAdult, returnedError = officeManagerResponsibleService.UpdateOfficeManager(ctx, officeManagerToUpdate)
		})

		Context("default", func() {

			BeforeEach(func() {
				officeManagerToUpdate = OfficeManagerTransport{
					Id:       "aaa",
					Email:    "jonhdoe@gmail.com",
					Password: "123465789",
				}
			})

			It("should update all fields but the password and the id", func() {
				Expect(updatedAdult).To(Equal(store.OfficeManager{
					OfficeManagerId: "aaa",
					Email:           "jonhdoe@gmail.com",
				}))
			})
			assertNoError()
		})

		Context("when the email is invalid", func() {
			BeforeEach(func() {
				officeManagerToUpdate = OfficeManagerTransport{
					Id:    "aaa",
					Email: "jonhdoegmail.com",
				}
			})
			assertErrorWithCause(ErrInvalidEmail)
		})
	})

	Context("GetOfficeManager", func() {

		var (
			returnedAdult store.OfficeManager
			request       OfficeManagerTransport
		)

		JustBeforeEach(func() {
			returnedAdult, returnedError = officeManagerResponsibleService.GetOfficeManager(ctx, request)
		})

		BeforeEach(func() {
			_, e1 := officeManagerResponsibleService.AddOfficeManager(ctx, officeManagerRef1)
			if e1 != nil {
				panic(e1)
			}
			_, e2 := officeManagerResponsibleService.AddOfficeManager(ctx, officeManagerRef2)
			if e2 != nil {
				panic(e2)
			}
		})

		Context("default", func() {
			BeforeEach(func() { request = OfficeManagerTransport{Id: "aaa"} })

			It("should return an user", func() {
				Expect(returnedAdult).To(Equal(store.OfficeManager{
					OfficeManagerId: "aaa",
					Email:           "arthur.gustin@gmail.com",
				}))
			})
			assertNoError()
		})

		Context("when the user does not exists", func() {
			BeforeEach(func() { request = OfficeManagerTransport{Id: "rgrgdfgb"} })
			assertErrorWithCause(store.ErrUserNotFound)
		})
	})

	Context("DeleteOfficeManager", func() {

		var (
			request OfficeManagerTransport
		)

		JustBeforeEach(func() {
			returnedError = officeManagerResponsibleService.DeleteOfficeManager(ctx, request)
		})

		BeforeEach(func() {
			officeManagerResponsibleService.AddOfficeManager(ctx, officeManagerRef1)
			officeManagerResponsibleService.AddOfficeManager(ctx, officeManagerRef2)

			Context("default", func() {
				BeforeEach(func() { request = OfficeManagerTransport{Id: "aaa"} })
				assertNoError()
			})

			Context("when the user does not exists", func() {
				BeforeEach(func() { request = OfficeManagerTransport{Id: "rgrgdfgb"} })
				assertErrorWithCause(store.ErrUserNotFound)
			})
		})
	})
})
