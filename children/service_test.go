package children_test

import (
	. "arthurgustin.fr/teddycare/children"

	"arthurgustin.fr/teddycare/shared/mocks"
	"arthurgustin.fr/teddycare/store"
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
		mockStringGenerator *shared.MockStringGenerator
		concreteStore       *store.Store
		concreteDb          *gorm.DB
		returnedError       error
		childRef1           ChildTransport
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
		childRef1 = ChildTransport{
			Id:            "aaa",
			FirstName:     "Arthur",
			LastName:      "Gustin",
			BirthDate:     "1992/10/13",
			Relationship:  "father",
			ResponsibleId: "aaa",
			PicturePath:   "gs://foo/bar/picture.jpg",
			Gender:        "M",
			Allergies:     []string{"tomato", "strawberry"},
		}
	})

	BeforeSuite(func() {
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
		childService = &ChildService{
			Store: concreteStore,
		}
		mockStringGenerator.On("GenerateUuid").Return("aaa").Once()
		mockStringGenerator.On("GenerateUuid").Return("bbb").Once()
		mockStringGenerator.On("GenerateUuid").Return("ccc").Once()
		mockStringGenerator.On("GenerateUuid").Return("ddd").Once()
	})

	Context("AddChild", func() {

		var (
			createdChild store.Child
		)

		var (
			assertCreatedTheRightChild = func() {
				It("should create the right child", func() {
					Expect(createdChild).To(Equal(store.Child{
						FirstName:   "Arthur",
						LastName:    "Gustin",
						BirthDate:   time.Date(1992, 10, 13, 0, 0, 0, 0, time.UTC),
						ChildId:     "aaa",
						PicturePath: "gs://foo/bar/picture.jpg",
						Gender:      "M",
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
			concreteStore.Db.Exec(`TRUNCATE TABLE "allergies" CASCADE`)
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

	Context("DeleteChild", func() {

		BeforeEach(func() {
			concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","password") VALUES ('aaa','arthur.gustin@gmail.com','$2a$10$nvGMsswN2Dtwy0iWg590ruMfwZTMaN8tR8/FpiW7ZG..WYEfpjKoS')`)
			concreteStore.Db.Exec(`INSERT INTO "adult_responsibles" ("responsible_id","email","first_name","last_name","gender","phone","addres_1","addres_2","city","state","zip") VALUES ('aaa','arthur.gustin@gmail.com','Arthur','Gustin','M','0633326825','11, rue hergé','app 8','Toulouse','FRANCE','31')`)
			concreteStore.Db.Exec(`INSERT INTO "children" ("child_id","first_name","last_name","birth_date","gender","picture_path") VALUES ('aaa','Arthur','Gustin','1992-10-13T00:00:00Z','M','gs://foo/bar/picture.jpg')`)
			concreteStore.Db.Exec(`INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('aaa','aaa','father')`)
		})

		AfterEach(func() {
			concreteStore.Db.Exec(`TRUNCATE TABLE "users" CASCADE`)
			concreteStore.Db.Exec(`TRUNCATE TABLE "children" CASCADE`)
			concreteStore.Db.Exec(`TRUNCATE TABLE "adult_responsibles" CASCADE`)
			concreteStore.Db.Exec(`TRUNCATE TABLE "responsible_of" CASCADE`)
		})

		JustBeforeEach(func() {
			returnedError = childService.DeleteChild(ctx, childRef1)
		})

		Context("default", func() {
			assertNoError()
		})

		Context("when the child does not exists", func() {
			BeforeEach(func() {
				childRef1.Id = "qdvsdv"
			})
			assertErrorWithCause(store.ErrChildNotFound)
		})

	})

	Context("GetChild", func() {

		var (
			returnedChild store.Child
		)

		BeforeEach(func() {
			concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","password") VALUES ('aaa','arthur.gustin@gmail.com','$2a$10$nvGMsswN2Dtwy0iWg590ruMfwZTMaN8tR8/FpiW7ZG..WYEfpjKoS')`)
			concreteStore.Db.Exec(`INSERT INTO "adult_responsibles" ("responsible_id","email","first_name","last_name","gender","phone","addres_1","addres_2","city","state","zip") VALUES ('aaa','arthur.gustin@gmail.com','Arthur','Gustin','M','0633326825','11, rue hergé','app 8','Toulouse','FRANCE','31')`)
			concreteStore.Db.Exec(`INSERT INTO "children" ("child_id","first_name","last_name","birth_date","gender","picture_path") VALUES ('aaa','Arthur','Gustin','1992-10-13T00:00:00Z','M','gs://foo/bar/picture.jpg')`)
			concreteStore.Db.Exec(`INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('aaa','aaa','father')`)
		})

		AfterEach(func() {
			concreteStore.Db.Exec(`TRUNCATE TABLE "users" CASCADE`)
			concreteStore.Db.Exec(`TRUNCATE TABLE "children" CASCADE`)
			concreteStore.Db.Exec(`TRUNCATE TABLE "adult_responsibles" CASCADE`)
			concreteStore.Db.Exec(`TRUNCATE TABLE "responsible_of" CASCADE`)
		})

		JustBeforeEach(func() {
			returnedChild, returnedError = childService.GetChild(ctx, childRef1)
		})

		Context("default", func() {
			assertNoError()
		})

		Context("when the child does not exists", func() {
			BeforeEach(func() {
				childRef1.Id = "qdvsdv"
			})
			assertErrorWithCause(store.ErrChildNotFound)
		})

	})

	Context("ListChildren", func() {

		var (
			children []store.Child
		)

		var (
			assertRightNumberOfChild = func(count int) {
				It(fmt.Sprintf("should return %d children", count), func() {
					Expect(children).To(HaveLen(2))
				})
			}
		)

		BeforeEach(func() {
			concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","password") VALUES ('aaa','arthur.gustin@gmail.com','$2a$10$nvGMsswN2Dtwy0iWg590ruMfwZTMaN8tR8/FpiW7ZG..WYEfpjKoS')`)
			concreteStore.Db.Exec(`INSERT INTO "adult_responsibles" ("responsible_id","email","first_name","last_name","gender","phone","addres_1","addres_2","city","state","zip") VALUES ('aaa','arthur.gustin@gmail.com','Arthur','Gustin','M','0633326825','11, rue hergé','app 8','Toulouse','FRANCE','31')`)
			concreteStore.Db.Exec(`INSERT INTO "children" ("child_id","first_name","last_name","birth_date","gender","picture_path") VALUES ('aaa','Arthur','Gustin','1992-10-13T00:00:00Z','M','gs://foo/bar/picture.jpg')`)
			concreteStore.Db.Exec(`INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('aaa','aaa','father')`)
			concreteStore.Db.Exec(`INSERT INTO "children" ("child_id","first_name","last_name","birth_date","gender","picture_path") VALUES ('bbb','Arthur','Gustin','1992-10-13T00:00:00Z','M','gs://foo/bar/picture.jpg')`)
			concreteStore.Db.Exec(`INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('aaa','bbb','father')`)
		})

		AfterEach(func() {
			concreteStore.Db.Exec(`TRUNCATE TABLE "users" CASCADE`)
			concreteStore.Db.Exec(`TRUNCATE TABLE "children" CASCADE`)
			concreteStore.Db.Exec(`TRUNCATE TABLE "adult_responsibles" CASCADE`)
			concreteStore.Db.Exec(`TRUNCATE TABLE "responsible_of" CASCADE`)
		})

		JustBeforeEach(func() {
			children, returnedError = childService.ListChild(ctx)
		})

		Context("default", func() {
			assertNoError()
			assertRightNumberOfChild(2)
		})

	})

	Context("Update Child", func() {

		var (
			returnedChild store.Child
			requestChild  ChildTransport
		)

		var (
			assertUpdateRightFields = func() {
				It("should update the right fields", func() {
					Expect(returnedChild.ChildId).To(Equal("aaa"))
					Expect(returnedChild.FirstName).To(Equal("foo"))
					Expect(returnedChild.LastName).To(Equal("foo"))
					Expect(returnedChild.BirthDate.UTC().String()).To(Equal(time.Date(1992, 10, 10, 0, 0, 0, 0, time.UTC).UTC().String()))
					Expect(returnedChild.Gender).To(Equal("F"))
					Expect(returnedChild.PicturePath).To(Equal("foo"))
				})
			}
		)

		BeforeEach(func() {
			requestChild = ChildTransport{
				Id:          "aaa",
				PicturePath: "foo",
				BirthDate:   "1992/10/10",
				Allergies:   []string{"pancake"},
				LastName:    "foo",
				FirstName:   "foo",
				Gender:      "F",
			}
		})

		BeforeEach(func() {
			concreteStore.Db.Exec(`INSERT INTO "users" ("user_id","email","password") VALUES ('aaa','arthur.gustin@gmail.com','$2a$10$nvGMsswN2Dtwy0iWg590ruMfwZTMaN8tR8/FpiW7ZG..WYEfpjKoS')`)
			concreteStore.Db.Exec(`INSERT INTO "adult_responsibles" ("responsible_id","email","first_name","last_name","gender","phone","addres_1","addres_2","city","state","zip") VALUES ('aaa','arthur.gustin@gmail.com','Arthur','Gustin','M','0633326825','11, rue hergé','app 8','Toulouse','FRANCE','31')`)
			concreteStore.Db.Exec(`INSERT INTO "children" ("child_id","first_name","last_name","birth_date","gender","picture_path") VALUES ('aaa','Arthur','Gustin','1992-10-13T00:00:00Z','M','gs://foo/bar/picture.jpg')`)
			concreteStore.Db.Exec(`INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('aaa','aaa','father')`)
			concreteStore.Db.Exec(`INSERT INTO "children" ("child_id","first_name","last_name","birth_date","gender","picture_path") VALUES ('bbb','Arthur','Gustin','1992-10-13T00:00:00Z','M','gs://foo/bar/picture.jpg')`)
			concreteStore.Db.Exec(`INSERT INTO "responsible_of" ("responsible_id","child_id","relationship") VALUES ('aaa','bbb','father')`)
		})

		AfterEach(func() {
			concreteStore.Db.Exec(`TRUNCATE TABLE "users" CASCADE`)
			concreteStore.Db.Exec(`TRUNCATE TABLE "children" CASCADE`)
			concreteStore.Db.Exec(`TRUNCATE TABLE "adult_responsibles" CASCADE`)
			concreteStore.Db.Exec(`TRUNCATE TABLE "responsible_of" CASCADE`)
		})

		JustBeforeEach(func() {
			returnedChild, returnedError = childService.UpdateChild(ctx, requestChild)
		})

		Context("default", func() {
			assertNoError()
			assertUpdateRightFields()
		})

		Context("when the child does not exists", func() {
			BeforeEach(func() { requestChild.Id = "zrgrljhg" })
			assertErrorWithCause(store.ErrChildNotFound)
		})

	})

})
