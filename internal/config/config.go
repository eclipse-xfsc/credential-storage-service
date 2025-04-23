package config

import (
	configPkg "github.com/eclipse-xfsc/microservice-core-go/pkg/config"
	"github.com/kelseyhightower/envconfig"
)

type storageConfiguration struct {
	configPkg.BaseConfig `mapstructure:",squash"`
	Profile              string `mapstructure:"profile" envconfig:"STORAGESERVICE_PROFILE" default:"DEBUG:LOCAL"`
	Mode                 string `mapstructure:"mode" envconfig:"STORAGESERVICE_MODE" default:"DIRECT"`
	UnitTestModeOn       bool   `mapstructure:"unitTestModeOn" envconfig:"STORAGESERVICE_UNITTESTMODEON" default:"false"`
	Country              string `mapstructure:"country" envconfig:"STORAGESERVICE_COUNTRY"`
	Region               string `mapstructure:"region" envconfig:"STORAGESERVICE_REGION"`

	Messaging struct {
		Enabled      bool   `mapstructure:"enabled" envconfig:"STORAGESERVICE_MESSAGING_ENABLED" default:"false"`
		StorageTopic string `mapstructure:"storageTopic" envconfig:"STORAGESERVICE_MESSAGING_STORAGETOPIC"`
		Url          string `mapstructure:"url" envconfig:"STORAGESERVICE_MESSAGING_URL"`
		QueueGroup   string `mapstructure:"queueGroup" envconfig:"STORAGESERVICE_MESSAGING_QUEUEGROUP"`
	} `mapstructure:"messaging"`

	Crypto struct {
		Namespace  string `mapstructure:"namespace" envconfig:"STORAGESERVICE_CRYPTO_NAMESPACE"`
		SignKey    string `mapstructure:"signKey" envconfig:"STORAGESERVICE_CRYPTO_SIGNKEY"`
		PluginPath string `mapstructure:"pluginPath" envconfig:"STORAGESERVICE_CRYPTO_PLUGINPATH" default:"/etc/plugins"`
	} `mapstructure:"crypto"`

	Cassandra struct {
		Host     string `mapstructure:"host" envconfig:"STORAGESERVICE_CASSANDRA_HOST"`
		KeySpace string `mapstructure:"keyspace" envconfig:"STORAGESERVICE_CASSANDRA_KEYSPACE"`
		User     string `mapstructure:"user, omitempty" envconfig:"STORAGESERVICE_CASSANDRA_USER"`
		Password string `mapstructure:"password, omitempty" envconfig:"STORAGESERVICE_CASSANDRA_PASSWORD"`
	} `mapstructure:"cassandra"`
}

var CurrentStorageConfig storageConfiguration

func Load() error {
	err := configPkg.LoadConfig("STORAGESERVICE", &CurrentStorageConfig, getDefaults())

	if err != nil {
		return err
	}

	return envconfig.Process("STORAGESERVICE", &CurrentStorageConfig)
}

func getDefaults() map[string]any {
	return map[string]any{
		"isDev":          false,
		"unitTestModeOn": false,
		"topic":          "storing",
	}
}
