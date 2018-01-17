package shared

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

const CONFIG_PREFIX = "TEDDYCARE"

type AppConfig struct {
	PgUsername                      string `split_words:"true" default:"postgres"`
	PgPassword                      string `split_words:"true" default:"postgres"`
	PgContactPoint                  string `split_words:"true" default:"localhost"`
	PgContactPort                   string `split_words:"true" default:"5432"`
	PgDbName                        string `split_words:"true" default:"teddycare"`
	SqlMigrationsSourceDir          string `split_words:"true"  default:"/go/src/airbus/datastore/workspace-manager/sql"`
	GcpProjectID                    string `split_words:"true" `
	PubSubServiceAccountKeyFilePath string `split_words:"true" `
}

func InitAppConfiguration() (config *AppConfig, err error) {
	config = &AppConfig{}

	if err := envconfig.Process(CONFIG_PREFIX, config); err != nil {
		return nil, fmt.Errorf("failed to parse env vars: %v", err)
	}

	return
}
