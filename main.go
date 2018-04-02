package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Vinubaba/SANTC-API/children"
	"github.com/Vinubaba/SANTC-API/classes"
	teddyFirebase "github.com/Vinubaba/SANTC-API/firebase"
	. "github.com/Vinubaba/SANTC-API/shared"
	"github.com/Vinubaba/SANTC-API/storage"
	. "github.com/Vinubaba/SANTC-API/store"
	"github.com/Vinubaba/SANTC-API/users"

	"firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/Vinubaba/SANTC-API/ageranges"
	"github.com/Vinubaba/SANTC-API/authentication"
	"github.com/Vinubaba/SANTC-API/store/migrations"
	"github.com/facebookgo/inject"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"google.golang.org/api/option"
)

var (
	ctx                     = context.Background()
	logger                  = NewLogger("teddycare")
	config                  *AppConfig
	db                      *gorm.DB
	stringGenerator         = &StringGenerator{}
	childService            = &children.ChildService{}
	userService             = &users.UserService{}
	userHandlerFactory      = &users.HandlerFactory{}
	childrenHandlerFactory  = &children.HandlerFactory{}
	classesHandlerFactory   = &classes.HandlerFactory{}
	ageRangesHandlerFactory = &ageranges.HandlerFactory{}
	teddyFirebaseClient     = &teddyFirebase.Client{}

	dbStore    = &Store{}
	gcsStorage = &storage.GoogleStorage{}

	firebaseClient *auth.Client
	authenticator  = &authentication.Authenticator{}
)

func init() {
	checkErrAndExit(initAppConfiguration())
	checkErrAndExit(initPostgresConnection())
	checkErrAndExit(initFirebase())
	checkErrAndExit(initApplicationGraph())
	checkErrAndExit(setPublicDaycare())
}

func initAppConfiguration() (err error) {
	config, err = InitAppConfiguration()
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
	config := &firebase.Config{ProjectID: "teddycare-193910"}

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
		&inject.Object{Value: childService},
		&inject.Object{Value: userService},
		&inject.Object{Value: userHandlerFactory},
		&inject.Object{Value: childrenHandlerFactory},
		&inject.Object{Value: classesHandlerFactory},
		&inject.Object{Value: ageRangesHandlerFactory},
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

func setPublicDaycare() error {
	publicDaycare, err := dbStore.GetPublicDaycare(db)
	if err != nil {
		return err
	}
	config.PublicDaycareId = publicDaycare.DaycareId.String
	return nil
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

	router := mux.NewRouter()

	apiRouterV1 := router.PathPrefix("/api/v1").Subrouter()

	apiRouterV1.Handle("/me", authenticator.Roles(userHandlerFactory.Me(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER, ROLE_ADULT, ROLE_TEACHER)).Methods(http.MethodGet)

	apiRouterV1.Handle("/office-managers", authenticator.Roles(userHandlerFactory.ListOfficeManager(userOpts), ROLE_ADMIN)).Methods(http.MethodGet)
	apiRouterV1.Handle("/office-managers/{id}", authenticator.Roles(userHandlerFactory.GetOfficeManager(userOpts), ROLE_ADMIN)).Methods(http.MethodGet)
	apiRouterV1.Handle("/office-managers/{id}", authenticator.Roles(userHandlerFactory.DeleteOfficeManager(userOpts), ROLE_ADMIN)).Methods(http.MethodDelete)
	apiRouterV1.Handle("/office-managers/{id}", authenticator.Roles(userHandlerFactory.UpdateOfficeManager(userOpts), ROLE_ADMIN)).Methods(http.MethodPatch)

	apiRouterV1.Handle("/teachers", authenticator.Roles(userHandlerFactory.CreateTeacher(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodPost)
	apiRouterV1.Handle("/teachers", authenticator.Roles(userHandlerFactory.ListTeacher(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodGet)
	apiRouterV1.Handle("/teachers/{id}", authenticator.Roles(userHandlerFactory.GetTeacher(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodGet)
	apiRouterV1.Handle("/teachers/{id}", authenticator.Roles(userHandlerFactory.DeleteTeacher(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodDelete)
	apiRouterV1.Handle("/teachers/{id}", authenticator.Roles(userHandlerFactory.UpdateTeacher(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodPatch)

	apiRouterV1.Handle("/adults", authenticator.Roles(userHandlerFactory.CreateAdult(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodPost)
	apiRouterV1.Handle("/adults", authenticator.Roles(userHandlerFactory.ListAdult(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodGet)
	apiRouterV1.Handle("/adults/{id}", authenticator.Roles(userHandlerFactory.GetAdult(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodGet)
	apiRouterV1.Handle("/adults/{id}", authenticator.Roles(userHandlerFactory.DeleteAdult(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodDelete)
	apiRouterV1.Handle("/adults/{id}", authenticator.Roles(userHandlerFactory.UpdateAdult(userOpts), ROLE_ADMIN, ROLE_OFFICE_MANAGER)).Methods(http.MethodPatch)

	apiRouterV1.Handle("/children", authenticator.Roles(childrenHandlerFactory.Add(childrenOpts), ROLE_OFFICE_MANAGER, ROLE_ADULT, ROLE_ADMIN)).Methods(http.MethodPost)
	apiRouterV1.Handle("/children", authenticator.Roles(childrenHandlerFactory.List(childrenOpts), ROLE_OFFICE_MANAGER, ROLE_ADULT, ROLE_ADMIN, ROLE_TEACHER)).Methods(http.MethodGet)
	apiRouterV1.Handle("/children/{childId}", authenticator.Roles(childrenHandlerFactory.Get(childrenOpts), ROLE_OFFICE_MANAGER, ROLE_ADULT, ROLE_ADMIN, ROLE_TEACHER)).Methods(http.MethodGet)
	apiRouterV1.Handle("/children/{childId}", authenticator.Roles(childrenHandlerFactory.Update(childrenOpts), ROLE_OFFICE_MANAGER, ROLE_ADULT, ROLE_ADMIN)).Methods(http.MethodPatch)
	apiRouterV1.Handle("/children/{childId}", authenticator.Roles(childrenHandlerFactory.Delete(childrenOpts), ROLE_OFFICE_MANAGER, ROLE_ADMIN)).Methods(http.MethodDelete)

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

	if config.TestAuthMode {
		testAuthRouter := mux.NewRouter()
		testAuthRouter.HandleFunc("/test-auth-login", authentication.ServeTestAuth).Methods(http.MethodGet)
		testAuthRouter.HandleFunc("/test-auth-on-success", authentication.ServeTestAuthOnSuccess)
		go func() {
			checkErrAndExit(http.ListenAndServe(":8082", testAuthRouter))
		}()
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://teddy-care-project.firebaseapp.com", "teddy-care-project.firebaseapp.com", "http://localhost:4200"},
		AllowCredentials: true,
		Debug:            true,
	})
	checkErrAndExit(http.ListenAndServe(":8083",
		logger.RequestLoggerMiddleware(
			c.Handler(
				authenticator.Firebase(router),
			),
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
