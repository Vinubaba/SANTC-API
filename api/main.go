package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Vinubaba/SANTC-API/api/ageranges"
	"github.com/Vinubaba/SANTC-API/api/authentication"
	"github.com/Vinubaba/SANTC-API/api/children"
	"github.com/Vinubaba/SANTC-API/api/classes"
	"github.com/Vinubaba/SANTC-API/api/daycares"
	. "github.com/Vinubaba/SANTC-API/api/shared"
	"github.com/Vinubaba/SANTC-API/api/users"
	teddyFirebase "github.com/Vinubaba/SANTC-API/common/firebase"
	"github.com/Vinubaba/SANTC-API/common/log"
	. "github.com/Vinubaba/SANTC-API/common/roles"
	"github.com/Vinubaba/SANTC-API/common/storage"
	. "github.com/Vinubaba/SANTC-API/common/store"
	"github.com/Vinubaba/SANTC-API/common/store/migrations"

	"firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/Vinubaba/SANTC-API/api/schedules"
	"github.com/Vinubaba/SANTC-API/common/generator"
	"github.com/facebookgo/inject"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

var (
	ctx             = context.Background()
	swagger         []byte
	logger          = log.NewLogger("teddycare")
	config          *AppConfig
	db              *gorm.DB
	stringGenerator = &generator.StringGenerator{}

	daycareService  = &daycares.DaycareService{}
	childService    = &children.ChildService{}
	userService     = &users.UserService{}
	classService    = &classes.ClassService{}
	ageRangeService = &ageranges.AgeRangeService{}
	scheduleService = &schedules.ScheduleService{}

	daycareHandlerFactory   = &daycares.HandlerFactory{}
	userHandlerFactory      = &users.HandlerFactory{}
	childrenHandlerFactory  = &children.HandlerFactory{}
	classesHandlerFactory   = &classes.HandlerFactory{}
	ageRangesHandlerFactory = &ageranges.HandlerFactory{}
	schedulesHandlerFactory = &schedules.HandlerFactory{}

	teddyFirebaseClient = &teddyFirebase.Client{}

	dbStore    = &Store{}
	gcsStorage *storage.GoogleStorage

	firebaseClient *auth.Client
	authenticator  = &authentication.Authenticator{}
)

func init() {
	checkErrAndExit(initAppConfiguration())
	checkErrAndExit(initStorage())
	checkErrAndExit(initPostgresConnection())
	checkErrAndExit(initFirebase())
	checkErrAndExit(initApplicationGraph())
	checkErrAndExit(initSwagger())
}

func initAppConfiguration() (err error) {
	config, err = InitAppConfiguration()
	return
}

func initStorage() (err error) {
	gcsStorage, err = storage.New(ctx, storage.Options{
		BucketName:      config.BucketName,
		CredentialsFile: config.BucketServiceAccount,
	})
	return
}

func initPostgresConnection() (err error) {
	connectString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.PgContactPoint,
		config.PgContactPort,
		config.PgUsername,
		config.PgPassword,
		config.PgDbName)
	db, err = gorm.Open("postgres", connectString)
	if err != nil {
		return
	}

	db.LogMode(true)
	db.SetLogger(logger)
	return
}

func initFirebase() error {
	opt := option.WithCredentialsFile(config.FirebaseServiceAccount)
	config := &firebase.Config{ProjectID: config.GcpProjectID}

	firebaseApp, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		return err
	}

	firebaseClient, err = firebaseApp.Auth(context.Background())
	if err != nil {
		return errors.Wrap(err, "error getting Auth client")
	}

	return nil
}

func initApplicationGraph() error {
	g := inject.Graph{}
	g.Provide(
		&inject.Object{Value: config},
		&inject.Object{Value: daycareService},
		&inject.Object{Value: childService},
		&inject.Object{Value: userService},
		&inject.Object{Value: classService},
		&inject.Object{Value: ageRangeService},
		&inject.Object{Value: scheduleService},
		&inject.Object{Value: userHandlerFactory},
		&inject.Object{Value: daycareHandlerFactory},
		&inject.Object{Value: childrenHandlerFactory},
		&inject.Object{Value: classesHandlerFactory},
		&inject.Object{Value: ageRangesHandlerFactory},
		&inject.Object{Value: schedulesHandlerFactory},
		&inject.Object{Value: db},
		&inject.Object{Value: stringGenerator},
		&inject.Object{Value: dbStore},
		&inject.Object{Value: gcsStorage},
		&inject.Object{Value: teddyFirebaseClient, Name: "teddyFirebaseClient"},
		&inject.Object{Value: firebaseClient},
		&inject.Object{Value: authenticator},
		&inject.Object{Value: logger},
	)
	if err := g.Populate(); err != nil {
		return errors.Wrap(err, "failed to populate")
	}
	return nil
}

func initSwagger() error {
	var err error
	swagger, err = ioutil.ReadFile(config.SwaggerFilePath)
	return err
}

func main() {
	if config.StartupMigration {
		applySqlSchemaMigrations(ctx)
	}
	startHttpServer(ctx)
}

func applySqlSchemaMigrations(ctx context.Context) {
	logger.Info(ctx, "applying sql schema migrations")
	migrationResult := migrations.Up(migrations.ApplyOptions{
		SourceURL: fmt.Sprintf("file://%s", config.SqlMigrationsSourceDir),
		DatabaseURL: fmt.Sprintf("postgres://%v:%v/%v?sslmode=disable&user=%s&password=%s",
			config.PgContactPoint, config.PgContactPort, config.PgDbName, config.PgUsername, config.PgPassword),
	})
	checkErrAndExit(migrationResult.Err)
	if !migrationResult.Changes {
		logger.Info(ctx, "no new migrations applied")
	}
}

func startHttpServer(ctx context.Context) {
	daycareOpts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(daycares.EncodeError),
	}

	userOpts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(users.EncodeError),
	}

	childrenOpts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(children.EncodeError),
	}

	classesOpts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(classes.EncodeError),
	}

	ageRangesOpts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(ageranges.EncodeError),
	}

	schedulesOpts := []kithttp.ServerOption{
		kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorEncoder(schedules.EncodeError),
	}

	router := mux.NewRouter()

	router.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	router.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	router.HandleFunc("/swagger.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(swagger)
	})

	if config.TestAuthMode {
		router.HandleFunc("/auth/login", authentication.ServeTestAuth).Methods(http.MethodGet)
		router.HandleFunc("/auth/success", authentication.ServeTestAuthOnSuccess)
	}

	apiRouterV1 := router.PathPrefix("/api/v1").Subrouter()

	apiRouterV1.Handle("/me", authenticator.Roles(userHandlerFactory.Me(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER, ROLE_ADULT, ROLE_TEACHER)).Methods(http.MethodGet)

	apiRouterV1.Handle("/daycares", authenticator.Roles(daycareHandlerFactory.Add(daycareOpts), ROLE_ADMIN)).Methods(http.MethodPost)
	apiRouterV1.Handle("/daycares", authenticator.Roles(daycareHandlerFactory.List(daycareOpts), ROLE_ADMIN)).Methods(http.MethodGet)
	apiRouterV1.Handle("/daycares/{daycareId}", authenticator.Roles(daycareHandlerFactory.Get(daycareOpts), ROLE_ADMIN)).Methods(http.MethodGet)
	apiRouterV1.Handle("/daycares/{daycareId}", authenticator.Roles(daycareHandlerFactory.Update(daycareOpts), ROLE_ADMIN)).Methods(http.MethodPatch)
	apiRouterV1.Handle("/daycares/{daycareId}", authenticator.Roles(daycareHandlerFactory.Delete(daycareOpts), ROLE_ADMIN)).Methods(http.MethodDelete)

	apiRouterV1.Handle("/office-managers", authenticator.Roles(userHandlerFactory.ListOfficeManager(userOpts), ROLE_ADMIN)).Methods(http.MethodGet)
	apiRouterV1.Handle("/office-managers/{id}", authenticator.Roles(userHandlerFactory.GetOfficeManager(userOpts), ROLE_ADMIN)).Methods(http.MethodGet)
	apiRouterV1.Handle("/office-managers/{id}", authenticator.Roles(userHandlerFactory.DeleteOfficeManager(userOpts), ROLE_ADMIN)).Methods(http.MethodDelete)
	apiRouterV1.Handle("/office-managers/{id}", authenticator.Roles(userHandlerFactory.UpdateOfficeManager(userOpts), ROLE_ADMIN)).Methods(http.MethodPatch)

	apiRouterV1.Handle("/teachers", authenticator.Roles(userHandlerFactory.CreateTeacher(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodPost)
	apiRouterV1.Handle("/teachers", authenticator.Roles(userHandlerFactory.ListTeacher(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER, ROLE_ADULT)).Methods(http.MethodGet)
	apiRouterV1.Handle("/teachers/{id}", authenticator.Roles(userHandlerFactory.GetTeacher(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodGet)
	apiRouterV1.Handle("/teachers/{id}", authenticator.Roles(userHandlerFactory.DeleteTeacher(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodDelete)
	apiRouterV1.Handle("/teachers/{id}", authenticator.Roles(userHandlerFactory.UpdateTeacher(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodPatch)
	apiRouterV1.Handle("/teachers/{id}/classes", authenticator.Roles(userHandlerFactory.SetTeacherClass(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodPost)
	apiRouterV1.Handle("/teachers/{teacherId}/schedules", authenticator.Roles(schedulesHandlerFactory.Add(schedulesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodPost)
	apiRouterV1.Handle("/teachers/{teacherId}/schedules/{scheduleId}", authenticator.Roles(schedulesHandlerFactory.Get(schedulesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodGet)
	apiRouterV1.Handle("/teachers/{teacherId}/schedules/{scheduleId}", authenticator.Roles(schedulesHandlerFactory.Update(schedulesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodPatch)
	apiRouterV1.Handle("/teachers/{teacherId}/schedules/{scheduleId}", authenticator.Roles(schedulesHandlerFactory.Delete(schedulesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodDelete)

	apiRouterV1.Handle("/adults", authenticator.Roles(userHandlerFactory.CreateAdult(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodPost)
	apiRouterV1.Handle("/adults", authenticator.Roles(userHandlerFactory.ListAdult(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodGet)
	apiRouterV1.Handle("/adults/{id}", authenticator.Roles(userHandlerFactory.GetAdult(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodGet)
	apiRouterV1.Handle("/adults/{id}", authenticator.Roles(userHandlerFactory.DeleteAdult(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodDelete)
	apiRouterV1.Handle("/adults/{id}", authenticator.Roles(userHandlerFactory.UpdateAdult(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodPatch)

	apiRouterV1.Handle("/children", authenticator.Roles(childrenHandlerFactory.Add(childrenOpts), ROLE_OFFICE_MANAGER, ROLE_ADULT, ROLE_ADMIN)).Methods(http.MethodPost)
	apiRouterV1.Handle("/children", authenticator.Roles(childrenHandlerFactory.List(childrenOpts), ROLE_OFFICE_MANAGER, ROLE_ADULT, ROLE_ADMIN, ROLE_TEACHER)).Methods(http.MethodGet)
	apiRouterV1.Handle("/children/{childId}", authenticator.Roles(childrenHandlerFactory.Get(childrenOpts), ROLE_OFFICE_MANAGER, ROLE_ADULT, ROLE_ADMIN, ROLE_TEACHER, ROLE_SERVICE)).Methods(http.MethodGet)
	apiRouterV1.Handle("/children/{childId}", authenticator.Roles(childrenHandlerFactory.Update(childrenOpts), ROLE_OFFICE_MANAGER, ROLE_ADULT, ROLE_ADMIN)).Methods(http.MethodPatch)
	apiRouterV1.Handle("/children/{childId}", authenticator.Roles(childrenHandlerFactory.Delete(childrenOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodDelete)
	apiRouterV1.Handle("/children/{childId}/photos", authenticator.Roles(childrenHandlerFactory.AddPhoto(childrenOpts), ROLE_SERVICE)).Methods(http.MethodPost)
	apiRouterV1.Handle("/children/{childId}/schedules", authenticator.Roles(schedulesHandlerFactory.Add(schedulesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodPost)
	apiRouterV1.Handle("/children/{childId}/schedules/{scheduleId}", authenticator.Roles(schedulesHandlerFactory.Get(schedulesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodGet)
	apiRouterV1.Handle("/children/{childId}/schedules/{scheduleId}", authenticator.Roles(schedulesHandlerFactory.Update(schedulesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodPatch)
	apiRouterV1.Handle("/children/{childId}/schedules/{scheduleId}", authenticator.Roles(schedulesHandlerFactory.Delete(schedulesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodDelete)

	apiRouterV1.Handle("/age-ranges", authenticator.Roles(ageRangesHandlerFactory.Add(ageRangesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodPost)
	apiRouterV1.Handle("/age-ranges", authenticator.Roles(ageRangesHandlerFactory.List(ageRangesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodGet)
	apiRouterV1.Handle("/age-ranges/{ageRangeId}", authenticator.Roles(ageRangesHandlerFactory.Get(ageRangesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodGet)
	apiRouterV1.Handle("/age-ranges/{ageRangeId}", authenticator.Roles(ageRangesHandlerFactory.Update(ageRangesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodPatch)
	apiRouterV1.Handle("/age-ranges/{ageRangeId}", authenticator.Roles(ageRangesHandlerFactory.Delete(ageRangesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodDelete)

	apiRouterV1.Handle("/classes", authenticator.Roles(classesHandlerFactory.Add(classesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodPost)
	apiRouterV1.Handle("/classes", authenticator.Roles(classesHandlerFactory.List(classesOpts), ROLE_OFFICE_MANAGER, ROLE_ADULT, ROLE_ADMIN, ROLE_TEACHER)).Methods(http.MethodGet)
	apiRouterV1.Handle("/classes/{classId}", authenticator.Roles(classesHandlerFactory.Get(classesOpts), ROLE_OFFICE_MANAGER, ROLE_ADULT, ROLE_ADMIN, ROLE_TEACHER)).Methods(http.MethodGet)
	apiRouterV1.Handle("/classes/{classId}", authenticator.Roles(classesHandlerFactory.Update(classesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodPatch)
	apiRouterV1.Handle("/classes/{classId}", authenticator.Roles(classesHandlerFactory.Delete(classesOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodDelete)

	checkErrAndExit(http.ListenAndServe("0.0.0.0:8080",
		logger.RequestLoggerMiddleware(
			authenticator.Firebase(router, []string{"/healthz", "/readyz", "/auth/login", "/auth/success", "/swagger.yaml", "/api/v1/"}),
		),
	))
}

func checkErrAndExit(err error) {
	if err == nil {
		return
	}
	fmt.Println(err.Error())
	os.Exit(1)
}
