package shared

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

const CONFIG_PREFIX = "EVENT_MANAGER"

type AppConfig struct {
	PgUsername             string `split_words:"true" default:"postgres"`
	PgPassword             string `split_words:"true" default:"postgres"`
	PgContactPoint         string `split_words:"true" default:"127.0.0.1"`
	PgContactPort          string `split_words:"true" default:"5432"`
	PgDbName               string `split_words:"true" default:"teddycare"`
	SqlMigrationsSourceDir string `split_words:"true" default:"C:\\Users\\arthur\\gocode\\src\\github.com\\Vinubaba\\SANTC-API\\api\\sql"`
	LocalStoragePath       string `split_words:"true"`
	ApiServerHostname      string `split_words:"true" default:"teddycare"`

	GcpProjectID    string `split_words:"true" default:"teddy-care"`
	GcpSubscription string `split_words:"true" default:"events"`
	GcpTopic        string `split_words:"true" default:"events"`

	BucketName string `split_words:"true" default:"teddycare"`

	ServiceAccount string `split_words:"true" default:"C:\\Users\\arthur\\code\\kubernetes-configuration\\event-manager-sa.json"`

	StartupMigration bool `split_words:"true" default:"false"`
}

func InitAppConfiguration() (config *AppConfig, err error) {
	config = &AppConfig{}

	if err := envconfig.Process(CONFIG_PREFIX, config); err != nil {
		return nil, fmt.Errorf("failed to parse env vars: %v", err)
	}

	return
}
