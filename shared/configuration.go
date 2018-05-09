package shared

import (
	"fmt"

	"encoding/json"
	"github.com/kelseyhightower/envconfig"
	"io/ioutil"
)

const CONFIG_PREFIX = "TEDDYCARE"

type AppConfig struct {
	PgUsername             string `split_words:"true" default:"postgres"`
	PgPassword             string `split_words:"true" default:"postgres"`
	PgContactPoint         string `split_words:"true" default:"127.0.0.1"`
	PgContactPort          string `split_words:"true" default:"5432"`
	PgDbName               string `split_words:"true" default:"teddycare"`
	SqlMigrationsSourceDir string `split_words:"true" default:"C:\\Users\\arthur\\gocode\\src\\github.com\\Vinubaba\\SANTC-API\\sql"`
	GcpProjectID           string `split_words:"true" default:"teddy-care"`
	LocalStoragePath       string `split_words:"true"`

	BucketImagesName            string `split_words:"true" default:"teddycare-images"`
	BucketServiceAccount        string `split_words:"true" default:"C:\\Users\\arthur\\gocode\\src\\github.com\\Vinubaba\\deployment\\bucket-sa.json"`
	BucketServiceAccountDetails ServiceAccountDetails

	FirebaseServiceAccount string `split_words:"true" default:"C:\\Users\\arthur\\gocode\\src\\github.com\\Vinubaba\\deployment\\firebase-sa.json"`

	TestAuthMode     bool `split_words:"true" default:"true"`
	StartupMigration bool `split_words:"true" default:"false"`

	PublicDaycareId string `split_words:"true" default:"PUBLIC"`
	SwaggerFilePath string `split_words:"true" default:"C:\\Users\\arthur\\gocode\\src\\github.com\\Vinubaba\\SANTC-API\\.docs\\swagger.yml"`
}

func InitAppConfiguration() (config *AppConfig, err error) {
	config = &AppConfig{}

	if err := envconfig.Process(CONFIG_PREFIX, config); err != nil {
		return nil, fmt.Errorf("failed to parse env vars: %v", err)
	}

	if config.BucketServiceAccount != "" {
		b, err := ioutil.ReadFile(config.BucketServiceAccount)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(b, &config.BucketServiceAccountDetails); err != nil {
			return nil, err
		}
	}

	return
}

type ServiceAccountDetails struct {
	Type                    string `json:"type"`
	ProjectId               string `json:"project_id"`
	PrivateKeyId            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientId                string `json:"client_id"`
	AuthUri                 string `json:"auth_uri"`
	TokenUri                string `json:"token_uri"`
	AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
	ClientX509CertUrl       string `json:"client_x509_cert_url"`
}
