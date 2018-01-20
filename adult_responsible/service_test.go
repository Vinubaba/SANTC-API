package adult_responsible_test

import (
	. "arthurgustin.fr/teddycare/adult_responsible"

	"arthurgustin.fr/teddycare/shared/mocks"
	"arthurgustin.fr/teddycare/store"
	"arthurgustin.fr/teddycare/store/mocks"
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
		ctx                         = context.Background()
		adultResponsibleService     Service
		mockStore                   *mocks.MockStore
		mockStringGenerator         *shared.MockStringGenerator
		concreteStore               *store.Store
		concreteDb                  *gorm.DB
		addAdultResponsibleErr, err error
		responsibleId               string
		adultResponsibleRequest     AddAdultResponsibleRequest
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
		mockStore = &mocks.MockStore{}
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
		JustBeforeEach(func() {
			responsibleId, addAdultResponsibleErr = adultResponsibleService.AddAdultResponsible(ctx, adultResponsibleRequest)
		})

		Context("default", func() {
			BeforeEach(func() {
				adultResponsibleRequest = AddAdultResponsibleRequest{
					FirstName: "Arthur",
					LastName:  "Gustin",
					Gender:    "M",
					Email:     "arthur.gustin@gmail.com",
					Password:  "azerty",
				}
			})

			It("should work", func() {
				Expect(addAdultResponsibleErr).To(BeNil())
				Expect(responsibleId).NotTo(BeZero())
			})
		})

		Context("when the email is invalid", func() {
			BeforeEach(func() {
				adultResponsibleRequest = AddAdultResponsibleRequest{
					FirstName: "Arthur",
					LastName:  "Gustin",
					Gender:    "M",
					Email:     "arthur.gustingmail.com",
					Password:  "azerty",
				}
			})

			It("should return an error", func() {
				Expect(addAdultResponsibleErr).NotTo(BeNil())
				Expect(errors.Cause(addAdultResponsibleErr)).To(Equal(ErrInvalidEmail))
				Expect(responsibleId).To(BeZero())
			})
		})

		Context("when the password is too short", func() {
			BeforeEach(func() {
				adultResponsibleRequest = AddAdultResponsibleRequest{
					FirstName: "Arthur",
					LastName:  "Gustin",
					Gender:    "M",
					Email:     "arthur.gustin@gmail.com",
					Password:  "123",
				}
			})

			It("should return an error", func() {
				Expect(addAdultResponsibleErr).NotTo(BeNil())
				Expect(errors.Cause(addAdultResponsibleErr)).To(Equal(ErrInvalidPasswordFormat))
				Expect(responsibleId).To(BeZero())
			})
		})
	})

	Context("ListAdultResponsible", func() {

		var (
			allAdultsResponsible []store.AdultResponsible
			err                  error
		)

		JustBeforeEach(func() {
			allAdultsResponsible, err = adultResponsibleService.ListAdultResponsible(ctx)
		})

		Context("default", func() {
			BeforeEach(func() {
				adultResponsibleService.AddAdultResponsible(ctx, AddAdultResponsibleRequest{
					FirstName: "Arthur",
					LastName:  "Gustin",
					Gender:    "M",
					Email:     "arthur.gustin@gmail.com",
					Password:  "azerty",
				})
				adultResponsibleService.AddAdultResponsible(ctx, AddAdultResponsibleRequest{
					FirstName: "Vinu",
					LastName:  "Singh",
					Gender:    "M",
					Email:     "vinu.singh@gmail.com",
					Password:  "qwerty",
				})
			})

			It("should work", func() {
				Expect(allAdultsResponsible).To(HaveLen(2))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("UpdateAdultResponsible", func() {

		var (
			adult         store.AdultResponsible
			adultToUpdate UpdateAdultResponsibleRequest
			err           error
		)

		BeforeEach(func() {
			adultResponsibleService.AddAdultResponsible(ctx, AddAdultResponsibleRequest{
				FirstName: "Arthur",
				LastName:  "Gustin",
				Gender:    "M",
				Email:     "arthur.gustin@gmail.com",
				Password:  "azerty",
			})
			adultResponsibleService.AddAdultResponsible(ctx, AddAdultResponsibleRequest{
				FirstName: "Vinu",
				LastName:  "Singh",
				Gender:    "M",
				Email:     "vinu.singh@gmail.com",
				Password:  "qwerty",
			})
		})

		JustBeforeEach(func() {
			adult, err = adultResponsibleService.UpdateAdultResponsible(ctx, adultToUpdate)
		})

		Context("default", func() {

			BeforeEach(func() {
				adultToUpdate = UpdateAdultResponsibleRequest{
					Id:        "aaa",
					FirstName: "john",
					LastName:  "doe",
					Gender:    "F",
					Email:     "jonhdoe@gmail.com",
				}
			})

			It("should update all fields", func() {
				Expect(adult).To(Equal(store.AdultResponsible{
					ResponsibleId: "aaa",
					FirstName:     "john",
					LastName:      "doe",
					Gender:        "F",
					Email:         "jonhdoe@gmail.com",
				}))
				Expect(err).To(BeNil())
			})
		})

		Context("when the email is invalid", func() {
			BeforeEach(func() {
				adultToUpdate = UpdateAdultResponsibleRequest{
					Email: "jonhdoegmail.com",
				}
			})

			It("should returns an error", func() {
				Expect(adult).To(BeZero())
				Expect(err).NotTo(BeNil())
				Expect(errors.Cause(err)).To(Equal(ErrInvalidEmail))
			})
		})
	})

	Context("GetAdultResponsible", func() {

		var (
			adult   store.AdultResponsible
			err     error
			request GetOrDeleteAdultResponsibleRequest
		)

		JustBeforeEach(func() {
			adult, err = adultResponsibleService.GetAdultResponsible(ctx, request)
		})

		BeforeEach(func() {
			adultResponsibleService.AddAdultResponsible(ctx, AddAdultResponsibleRequest{
				FirstName: "Arthur",
				LastName:  "Gustin",
				Gender:    "M",
				Email:     "arthur.gustin@gmail.com",
				Password:  "azerty",
			})
			adultResponsibleService.AddAdultResponsible(ctx, AddAdultResponsibleRequest{
				FirstName: "Vinu",
				LastName:  "Singh",
				Gender:    "M",
				Email:     "vinu.singh@gmail.com",
				Password:  "qwerty",
			})
		})

		Context("default", func() {
			BeforeEach(func() {
				request = GetOrDeleteAdultResponsibleRequest{
					Id: "aaa",
				}
			})

			It("should return an user", func() {
				Expect(adult).To(Equal(store.AdultResponsible{
					FirstName:     "Arthur",
					LastName:      "Gustin",
					Gender:        "M",
					Email:         "arthur.gustin@gmail.com",
					ResponsibleId: "aaa",
				}))
				Expect(err).To(BeNil())
			})
		})

		Context("when the user does not exists", func() {
			BeforeEach(func() {
				request = GetOrDeleteAdultResponsibleRequest{
					Id: "rgrgdfgb",
				}
			})

			It("should return an error", func() {
				Expect(adult).To(BeZero())
				Expect(err).NotTo(BeNil())
				Expect(errors.Cause(err)).To(Equal(store.ErrUserNotFound))
			})
		})
	})

	Context("DeleteAdultResponsible", func() {

		var (
			err     error
			request GetOrDeleteAdultResponsibleRequest
		)

		JustBeforeEach(func() {
			err = adultResponsibleService.DeleteAdultResponsible(ctx, request)
		})

		BeforeEach(func() {
			adultResponsibleService.AddAdultResponsible(ctx, AddAdultResponsibleRequest{
				FirstName: "Arthur",
				LastName:  "Gustin",
				Gender:    "M",
				Email:     "arthur.gustin@gmail.com",
				Password:  "azerty",
			})
			adultResponsibleService.AddAdultResponsible(ctx, AddAdultResponsibleRequest{
				FirstName: "Vinu",
				LastName:  "Singh",
				Gender:    "M",
				Email:     "vinu.singh@gmail.com",
				Password:  "qwerty",
			})
		})

		Context("default", func() {
			BeforeEach(func() {
				request = GetOrDeleteAdultResponsibleRequest{
					Id: "aaa",
				}
			})

			It("should not return an error", func() {
				Expect(err).To(BeNil())
			})
		})

		Context("when the user does not exists", func() {
			BeforeEach(func() {
				request = GetOrDeleteAdultResponsibleRequest{
					Id: "rgrgdfgb",
				}
			})

			It("should return an error", func() {
				Expect(err).NotTo(BeNil())
				Expect(errors.Cause(err)).To(Equal(store.ErrUserNotFound))
			})
		})
	})

})
