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
		ctx                 = context.Background()
		childService        Service
		mockStore           *mocks.MockStore
		mockStringGenerator *shared.MockStringGenerator
		concreteStore       *store.Store
		concreteDb          *gorm.DB
		addChildErr, err    error
		childRequest        ChildRequest
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
		mockStringGenerator.On("GenerateUuid").Return("aaa")
		mockStringGenerator.On("GenerateUuid").Return("bbb")
		mockStringGenerator.On("GenerateUuid").Return("ccc")
		mockStringGenerator.On("GenerateUuid").Return("ddd")
	})

	Context("AddChild", func() {

		var (
			createdChild store.Child
		)

		BeforeEach(func() {
			concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","password") VALUES ('aaa','arthur.gustin@gmail.com','$2a$10$nvGMsswN2Dtwy0iWg590ruMfwZTMaN8tR8/FpiW7ZG..WYEfpjKoS')`)
			concreteStore.Db.Exec(`INSERT INTO "adult_responsibles" ("responsible_id","email","first_name","last_name","gender") VALUES ('aaa','arthur.gustin@gmail.com','Patrick','Gustin','M')`)
		})

		AfterEach(func() {
			concreteStore.Db.Exec(`TRUNCATE TABLE "users" CASCADE`)
			concreteStore.Db.Exec(`TRUNCATE TABLE "children" CASCADE`)
		})

		JustBeforeEach(func() {
			createdChild, addChildErr = childService.AddChild(ctx, childRequest)
		})

		Context("default", func() {
			BeforeEach(func() {
				childRequest = ChildRequest{
					FirstName:     "Arthur",
					LastName:      "Gustin",
					BirthDate:     "1992/10/13",
					Relationship:  "father",
					ResponsibleId: "aaa",
				}
			})

			It("should work", func() {
				Expect(addChildErr).To(BeNil())
				Expect(createdChild.ChildId).NotTo(BeZero())
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
				Expect(createdChild).To(BeZero())
			})
		})

		Context("when the relationship is invalid", func() {
			BeforeEach(func() {
				childRequest = ChildRequest{
					FirstName:     "Arthur",
					LastName:      "Gustin",
					BirthDate:     "1992/10/13",
					Relationship:  "zefzef",
					ResponsibleId: "aaa",
				}
			})

			It("should return an error", func() {
				Expect(addChildErr).NotTo(BeNil())

				Expect(addChildErr.Error()).To(Equal("failed to set responsible: relationship is not valid, it should be one of [father mother grandfather grandmother guardian]"))
				Expect(createdChild).To(BeZero())
			})
		})

	})

})
