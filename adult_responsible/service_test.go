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
		adultResponsibleRequest     AdultResponsibleRequest
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
				adultResponsibleRequest = AdultResponsibleRequest{
					FirstName: "Arthur",
					LastName:  "Gustin",
					Gender:    "M",
				}
			})

			It("should work", func() {
				Expect(addAdultResponsibleErr).To(BeNil())
				Expect(responsibleId).NotTo(BeZero())
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
				adultResponsibleService.AddAdultResponsible(ctx, AdultResponsibleRequest{
					FirstName: "Arthur",
					LastName:  "Gustin",
					Gender:    "M",
				})
				adultResponsibleService.AddAdultResponsible(ctx, AdultResponsibleRequest{
					FirstName: "Vinu",
					LastName:  "Singh",
					Gender:    "M",
				})
			})

			It("should work", func() {
				Expect(allAdultsResponsible).To(HaveLen(2))
				Expect(err).To(BeNil())
			})
		})
	})

})
