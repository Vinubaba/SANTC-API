package childs_test

import (
	. "arthurgustin.fr/teddycare/childs"

	"arthurgustin.fr/teddycare/shared/mocks"
	"arthurgustin.fr/teddycare/store"
	"arthurgustin.fr/teddycare/store/mocks"
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service", func() {

	var (
		ctx                    = context.Background()
		childService           Service
		mockStore              *mocks.MockStore
		mockStringGenerator    *shared.MockStringGenerator
		concreteStore          *store.Store
		concreteDb             *gorm.DB
		addChildErr, err       error
		responsibleId, childId string
		childRequest           ChildRequest
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
		childService = &ChildService{
			Store: concreteStore,
		}
	})

	JustBeforeEach(func() {
		childId, addChildErr = childService.AddChild(ctx, childRequest)
	})

	Context("default", func() {
		BeforeEach(func() {
			var err error
			responsibleId, err = concreteStore.AddAdultResponsible(ctx, store.AdultResponsible{
				FirstName: "Patrick",
				LastName:  "Gustin",
				Gender:    "M",
			})
			if err != nil {
				panic(err)
			}
			childRequest = ChildRequest{
				FirstName:     "Arthur",
				LastName:      "Gustin",
				BirthDate:     "1992/10/13",
				Relationship:  "father",
				ResponsibleId: responsibleId,
			}
		})

		It("should work", func() {
			Expect(addChildErr).To(BeNil())
			Expect(childId).NotTo(BeZero())
		})
	})

	Context("when the responsibleId does not exists", func() {
		BeforeEach(func() {
			childRequest = ChildRequest{
				FirstName:     "Arthur",
				LastName:      "Gustin",
				BirthDate:     "1992/10/13",
				Relationship:  "father",
				ResponsibleId: "unknown",
			}
		})

		It("should fail", func() {
			Expect(addChildErr).NotTo(BeNil())
			Expect(childId).To(BeZero())
		})
	})

	Context("when the relationship is invalid", func() {
		BeforeEach(func() {
			var err error
			responsibleId, err = concreteStore.AddAdultResponsible(ctx, store.AdultResponsible{
				FirstName: "Patrick",
				LastName:  "Gustin",
				Gender:    "M",
			})
			if err != nil {
				panic(err)
			}
			childRequest = ChildRequest{
				FirstName:     "Arthur",
				LastName:      "Gustin",
				BirthDate:     "1992/10/13",
				Relationship:  "zefzef",
				ResponsibleId: responsibleId,
			}
		})

		It("should work", func() {
			Expect(addChildErr).NotTo(BeNil())
			Expect(addChildErr.Error()).To(Equal("failed to set responsible: relationship is not valid, it should be one of [father mother grandfather grandmother guardian]"))
			Expect(childId).To(BeZero())
		})
	})

})
