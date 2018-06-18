package shared

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

const CONFIG_PREFIX = "TEDDYCARE"

type AppConfig struct {
	PgUsername             string `split_words:"true" default:"postgres"`
	PgPassword             string `split_words:"true" default:"postgres"`
	PgContactPoint         string `split_words:"true" default:"127.0.0.1"`
	PgContactPort          string `split_words:"true" default:"5432"`
	PgDbName               string `split_words:"true" default:"teddycare"`
	SqlMigrationsSourceDir string `split_words:"true" default:"C:\\Users\\arthur\\gocode\\src\\github.com\\Vinubaba\\SANTC-API\\api\\sql"`
	GcpProjectID           string `split_words:"true" default:"teddy-care"`
	LocalStoragePath       string `split_words:"true"`

	BucketImagesName     string `split_words:"true" default:"teddycare-profiles"`
	BucketServiceAccount string `split_words:"true" default:"C:\\Users\\arthur\\code\\kubernetes-configuration\\bucket-sa.json"`

	FirebaseServiceAccount string `split_words:"true" default:"C:\\Users\\arthur\\code\\kubernetes-configuration\\firebase-sa.json"`

	TestAuthMode     bool `split_words:"true" default:"true"`
	StartupMigration bool `split_words:"true" default:"false"`

	PublicDaycareId string `split_words:"true" default:"PUBLIC"`
	SwaggerFilePath string `split_words:"true" default:"C:\\Users\\arthur\\gocode\\src\\github.com\\Vinubaba\\SANTC-API\\api\\.docs\\swagger.yml"`
}

func InitAppConfiguration() (config *AppConfig, err error) {
	config = &AppConfig{}
	if err := envconfig.Process(CONFIG_PREFIX, config); err != nil {
		return nil, fmt.Errorf("failed to parse env vars: %v", err)
	}

	return
}
