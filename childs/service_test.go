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
	"github.com/pkg/errors"
	"time"
)

var _ = Describe("Service", func() {

	var (
		ctx                 = context.Background()
		childService        Service
		mockStore           *mocks.MockStore
		mockStringGenerator *shared.MockStringGenerator
		concreteStore       *store.Store
		concreteDb          *gorm.DB
		returnedError, err  error
		childRef1           ChildRequest
	)

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

	BeforeEach(func() {
		childRef1 = ChildRequest{
			FirstName:     "Arthur",
			LastName:      "Gustin",
			BirthDate:     "1992/10/13",
			Relationship:  "father",
			ResponsibleId: "aaa",
		}
	})

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

		var (
			assertCreatedTheRightChild = func() {
				It("should create the right child", func() {
					Expect(createdChild).To(Equal(store.Child{
						FirstName: "Arthur",
						LastName:  "Gustin",
						BirthDate: time.Date(1992, 10, 13, 0, 0, 0, 0, time.UTC),
						ChildId:   "aaa",
					}))
				})
			}
		)

		BeforeEach(func() {
			concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","password") VALUES ('aaa','arthur.gustin@gmail.com','$2a$10$nvGMsswN2Dtwy0iWg590ruMfwZTMaN8tR8/FpiW7ZG..WYEfpjKoS')`)
			concreteStore.Db.Exec(`INSERT INTO "adult_responsibles" ("responsible_id","email","first_name","last_name","gender","phone","addres_1","addres_2","city","state","zip") VALUES ('aaa','arthur.gustin@gmail.com','Arthur','Gustin','M','0633326825','11, rue hergé','app 8','Toulouse','FRANCE','31')`)
		})

		AfterEach(func() {
			concreteStore.Db.Exec(`TRUNCATE TABLE "users" CASCADE`)
			concreteStore.Db.Exec(`TRUNCATE TABLE "children" CASCADE`)
		})

		JustBeforeEach(func() {
			createdChild, returnedError = childService.AddChild(ctx, childRef1)
		})

		Context("default", func() {
			assertNoError()
			assertCreatedTheRightChild()
		})

		Context("when the responsibleId does not exists", func() {
			BeforeEach(func() {
				childRef1.ResponsibleId = "unknown"
			})
			assertErrorWithCause(ErrSetResponsible)
		})

		Context("when the relationship is invalid", func() {
			BeforeEach(func() {
				childRef1.Relationship = "zefzef"
			})
			assertErrorWithCause(ErrSetResponsible)
		})

		Context("when the responsableId is empty", func() {
			BeforeEach(func() {
				childRef1.ResponsibleId = ""
			})
			assertErrorWithCause(ErrNoParent)
		})

	})

})
