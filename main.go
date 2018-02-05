package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/DigitalFrameworksLLC/teddycare/adult_responsible"
	"github.com/DigitalFrameworksLLC/teddycare/authentication"
	"github.com/DigitalFrameworksLLC/teddycare/children"
	"github.com/DigitalFrameworksLLC/teddycare/office_manager"
	"github.com/DigitalFrameworksLLC/teddycare/shared"
	"github.com/DigitalFrameworksLLC/teddycare/storage"
	"github.com/DigitalFrameworksLLC/teddycare/store"

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

	checkErrAndExit(http.ListenAndServe(":8080", router))
}

func checkErrAndExit(err error) {
	if err == nil {
		return
	}
	fmt.Println(err.Error())
	os.Exit(1)
}
