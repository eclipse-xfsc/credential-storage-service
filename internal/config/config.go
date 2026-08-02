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

	Nats struct {
		StorageTopic string `mapstructure:"storageTopic" envconfig:"STORAGESERVICE_NATS_STORAGETOPIC"`
		Url          string `mapstructure:"url" envconfig:"STORAGESERVICE_NATS_URL"`
		QueueGroup   string `mapstructure:"queueGroup" envconfig:"STORAGESERVICE_NATS_QUEUEGROUP"`
		TimeoutInSec string `mapstructure:"requestTimeout" envconfig:"STORAGESERVICE_NATS_REQUEST_TIMEOUT"`
	} `mapstructure:"nats"`

	Crypto struct {
		Namespace string `mapstructure:"namespace" envconfig:"STORAGESERVICE_CRYPTO_NAMESPACE"`
		SignKey   string `mapstructure:"signKey" envconfig:"STORAGESERVICE_CRYPTO_SIGNKEY"`
		GrpcAddr  string `mapstructure:"grpcAddr" envconfig:"STORAGESERVICE_CRYPTO_GRPC_ADDR" default:"crypto-provider:50051"`
	} `mapstructure:"crypto"`

	Cassandra struct {
		Host     string `mapstructure:"host" envconfig:"STORAGESERVICE_CASSANDRA_HOSTS"`
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
