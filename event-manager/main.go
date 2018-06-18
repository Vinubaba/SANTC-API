package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	teddyFirebase "github.com/Vinubaba/SANTC-API/common/firebase"
	"github.com/Vinubaba/SANTC-API/common/log"
	"github.com/Vinubaba/SANTC-API/common/messaging"
	"github.com/Vinubaba/SANTC-API/common/storage"
	"github.com/Vinubaba/SANTC-API/common/store"
	. "github.com/Vinubaba/SANTC-API/event-manager/shared"

	"cloud.google.com/go/pubsub"
	"firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/Vinubaba/SANTC-API/common/api"
	"github.com/Vinubaba/SANTC-API/common/generator"
	"github.com/Vinubaba/SANTC-API/event-manager/consumers"
	"github.com/facebookgo/inject"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var (
	ctx     = context.Background()
	swagger []byte
	logger  = log.NewLogger("event-manager")
	config  *AppConfig
	db      *gorm.DB

	teddyFirebaseClient = &teddyFirebase.Client{}

	dbStore    = &store.Store{}
	gcsStorage *storage.GoogleStorage

	firebaseClient *auth.Client

	pubSubClient *messaging.Client

	consumer             *consumers.Consumer
	imageApprovalHandler *consumers.ImageApprovalHandler
	stringGenerator      = &generator.StringGenerator{}
	apiClient            api.Client
)

func init() {
	checkErrAndExit(initAppConfiguration())
	checkErrAndExit(initApiClient())
	checkErrAndExit(initStorage())
	checkErrAndExit(initConsumerStarter())
	checkErrAndExit(initPubSubClient())
	checkErrAndExit(initPostgresConnection())
	checkErrAndExit(initFirebase())
	checkErrAndExit(initApplicationGraph())
}

func initConsumerStarter() (err error) {
	imageApprovalHandler = &consumers.ImageApprovalHandler{}
	consumer = &consumers.Consumer{}
	consumer.EventHandlers = append(consumer.EventHandlers, imageApprovalHandler)
	return
}

func initStorage() (err error) {
	gcsStorage, err = storage.New(ctx, storage.Options{
		BucketName:      config.BucketApprovalsName,
		CredentialsFile: config.ServiceAccount,
	})
	return
}

func initApiClient() (err error) {
	apiClient, err = api.NewDefaultClient("http", config.ApiServerHostname)
	return
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
	opt := option.WithCredentialsFile(config.ServiceAccount)
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
		&inject.Object{Value: db},
		&inject.Object{Value: dbStore},
		&inject.Object{Value: gcsStorage},
		&inject.Object{Value: teddyFirebaseClient, Name: "teddyFirebaseClient"},
		&inject.Object{Value: firebaseClient},
		&inject.Object{Value: logger},
		&inject.Object{Value: imageApprovalHandler},
		&inject.Object{Value: consumer},
		&inject.Object{Value: stringGenerator},
		&inject.Object{Value: pubSubClient},
		&inject.Object{Value: apiClient},
	)
	if err := g.Populate(); err != nil {
		return errors.Wrap(err, "failed to populate")
	}
	return nil
}

func main() {
	go consumer.Start(ctx)
	/*	if config.StartupMigration {
		applySqlSchemaMigrations(ctx)
	}*/
	startHttpServer(ctx)

}

func startHttpServer(ctx context.Context) {

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

	checkErrAndExit(http.ListenAndServe("0.0.0.0:8080", router))
}

func checkErrAndExit(err error) {
	if err == nil {
		return
	}
	fmt.Println(err.Error())
	os.Exit(1)
}

func initPubSubClient() (err error) {
	pubSubClient, err = messaging.New(messaging.ClientOptions{
		ProjectID:      config.GcpProjectID,
		Subscription:   config.GcpSubscription,
		Topic:          config.GcpTopic,
		CredentialPath: config.ServiceAccount,
	})
	if err != nil {
		return err
	}

	ensureTopicAndSubscriptionsAreCreated()

	return nil
}

func ensureTopicAndSubscriptionsAreCreated() {
	ensure := func() bool {
		it := pubSubClient.Topics(ctx)
		for {
			topic, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				logger.Err(ctx, "errors in listing topics", "err", err)
				return false
			}
			if topic.ID() == config.GcpTopic {
				if !subscriptionExist(config.GcpSubscription, topic) {
					_, err := pubSubClient.GetPubSubClient().CreateSubscription(ctx, config.GcpSubscription, pubsub.SubscriptionConfig{
						Topic:               topic,
						RetainAckedMessages: false,
						AckDeadline:         20 * time.Second,
					})
					if err != nil {
						logger.Err(ctx, "errors while creating subscription", "err", err)
						return false
					}
				}
				logger.Info(ctx, "subscription "+config.GcpTopic+" found !")
				return true
			}
		}
		logger.Info(ctx, "no existing topic with the name"+config.GcpTopic+". try to create one...")
		_, err := pubSubClient.GetPubSubClient().CreateTopic(ctx, config.GcpTopic)
		if err != nil {
			logger.Err(ctx, "failed to create topic "+config.GcpTopic)
		}
		return false
	}

	for !ensure() {
		time.Sleep(1 * time.Second)
	}
}

func subscriptionExist(subscriptionName string, topic *pubsub.Topic) bool {
	it := topic.Subscriptions(ctx)
	for {
		subscription, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.Err(ctx, "errors in listing subscriptions", "err", err)
			return false
		}
		if subscription.ID() == subscriptionName {
			return true
		}
	}
	return false
}
