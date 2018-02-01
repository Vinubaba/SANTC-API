package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"arthurgustin.fr/teddycare/adult_responsible"
	"arthurgustin.fr/teddycare/children"
	"arthurgustin.fr/teddycare/shared"
	"arthurgustin.fr/teddycare/store"

	"arthurgustin.fr/teddycare/authentication"
	"arthurgustin.fr/teddycare/office_manager"
	"arthurgustin.fr/teddycare/storage"
	"github.com/facebookgo/inject"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
)

var (
	ctx                     = context.Background()
	config                  *shared.AppConfig
	db                      *gorm.DB
	stringGenerator         = &shared.StringGenerator{}
	childService            = &children.ChildService{}
	adultResponsibleService = &adult_responsible.AdultResponsibleService{}
	officeManagerService    = office_manager.NewDefaultService()
	authenticationService   = &authentication.AuthenticationService{}
	dbStore                 = &store.Store{}
	gcsStorage              = &storage.GoogleStorage{}
)

func init() {
	checkErrAndExit(initAppConfiguration())
	checkErrAndExit(initPostgresConnection())
	checkErrAndExit(initApplicationGraph())
}

func initAppConfiguration() (err error) {
	config, err = shared.InitAppConfiguration()
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
	return
}

func initApplicationGraph() error {
	if err := inject.Populate(
		config,
		childService,
		adultResponsibleService,
		officeManagerService,
		authenticationService,
		db,
		stringGenerator,
		dbStore,
		gcsStorage,
	); err != nil {
		return errors.Wrap(err, "failed to populate")
	}
	return nil
}

func main() {
	ctx := context.Background()
	applySqlSchemaMigrations(ctx)
	startHttpServer(ctx)
}

func applySqlSchemaMigrations(ctx context.Context) {
	/*log.Info(ctx, "applying sql schema migrations", nil)
	migrationResult := migrations.Up(migrations.ApplyOptions{
		SourceURL: fmt.Sprintf("file://%s", config.SqlMigrationsSourceDir),
		DatabaseURL: fmt.Sprintf("postgres://%v:%v/%v?sslmode=disable&user=%s&password=%s",
			config.PgContactPoint, config.PgContactPort, config.PgDbName, config.PgUsername, config.PgPassword),
	})
	checkErrAndExit(migrationResult.Err)
	if !migrationResult.Changes {
		logger.Info(ctx, "no new migrations applied", nil)
	}*/
}

func startHttpServer(ctx context.Context) {
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	httpLogger := log.With(logger, "component", "http")

	router := mux.NewRouter()
	authentication.MakeAuthHandler(router, authenticationService, httpLogger)

	apiRouterV1 := router.PathPrefix("/api/v1").Subrouter()
	children.MakeHandler(apiRouterV1, childService, httpLogger)
	adult_responsible.MakeHandler(apiRouterV1, adultResponsibleService, httpLogger)
	office_manager.MakeHandler(apiRouterV1, officeManagerService, httpLogger)

	checkErrAndExit(http.ListenAndServe(":8084", router))

	/*router := mux.NewRouter()
	apiRouterV1 := router.PathPrefix("/api/v1").Subrouter()
	apiRouterV1.Handle("/", childs.MakeHandler(childService, httpLogger)).Methods(http.MethodPost)
	apiRouterV1.Handle("/childs/", childs.MakeHandler(childService, httpLogger)).Methods(http.MethodPost)
	apiRouterV1.Handle("/childs", childs.MakeHandler(childService, httpLogger)).Methods(http.MethodPost)
	apiRouterV1.Handle("/childs", api.NotImplemented).Methods(http.MethodGet)
	apiRouterV1.Handle("/childs/{id}", api.NotImplemented).Methods(http.MethodPatch)
	apiRouterV1.Handle("/childs/{id}", api.NotImplemented).Methods(http.MethodDelete)
	apiRouterV1.Handle("/childs/{id}", api.NotImplemented).Methods(http.MethodGet)

	http.Handle("/", router)

	//logger.Info(ctx, "starting http server", log.Fields{"port": 8080})
	checkErrAndExit(http.ListenAndServe(":8083", nil))*/
}

func checkErrAndExit(err error) {
	if err == nil {
		return
	}
	fmt.Println(err.Error())
	os.Exit(1)
}
