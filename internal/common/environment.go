package common

import (
	"github.com/eclipse-xfsc/credential-storage-service/docs"
	"github.com/eclipse-xfsc/credential-storage-service/internal/config"
	"github.com/eclipse-xfsc/credential-storage-service/internal/connection"
	cryptoProvider "github.com/eclipse-xfsc/credential-storage-service/internal/crypto"
	ginSwagger "github.com/swaggo/gin-swagger"

	logPkg "github.com/eclipse-xfsc/microservice-core-go/pkg/logr"

	"github.com/eclipse-xfsc/crypto-provider-core/types"
)

type Environment struct {
	session         connection.SessionInterface
	mode            string
	cryptoNamespace string
	signKey         string
	unitTestModeOn  bool
	contentType     string
	logger          logPkg.Logger
	isHealthy       bool
}

var env *Environment

func init() {
	env = new(Environment)
}

func GetEnvironment() *Environment {
	return env
}

func (e *Environment) SetSession(session connection.SessionInterface) {
	e.session = session
}

func (e *Environment) GetSession() connection.SessionInterface {
	return e.session
}

func (e *Environment) GetCryptoProvider() types.CryptoProvider {
	return cryptoProvider.GetCryptoProvider()
}

func (e *Environment) GetRegion() string {
	return config.CurrentStorageConfig.Region
}

func (e *Environment) GetCountry() string {
	return config.CurrentStorageConfig.Country
}

func (e *Environment) GetAccountPartition(account string) string {
	if len(account) < AccountPartitionLength {
		return account
	}
	return account[0:AccountPartitionLength]
}

func (e *Environment) SetCryptoNamespace(namespace string) {
	e.cryptoNamespace = namespace
}

func (e *Environment) GetCryptoNamespace() string {
	return e.cryptoNamespace
}

func (e *Environment) SetContentType(contentType string) {
	e.contentType = contentType
}

func (e *Environment) GetContentType() string {
	return e.contentType
}

func (e *Environment) SetCryptoSignKey(signKey string) {
	e.signKey = signKey
}

func (e *Environment) GetCryptoSignKey() string {
	return e.signKey
}

func (e *Environment) SetMode(mode string) {
	e.mode = mode
}

func (e *Environment) GetMode() string {
	return e.mode
}

func (e *Environment) SetUnitTestModeOn(mode bool) {
	e.unitTestModeOn = mode
}

func (e *Environment) GetUnitTestModeOn() bool {
	return e.unitTestModeOn
}

func (e *Environment) SetLogger(logger logPkg.Logger) {
	e.logger = logger
}

func (e *Environment) GetLogger() logPkg.Logger { return e.logger }

func (e *Environment) SetHealthy(isHealthy bool) {
	e.isHealthy = isHealthy
}

func (e *Environment) IsHealthy() bool {
	return !e.session.Closed()
}

func (env *Environment) SetSwaggerBasePath(path string) {
	docs.SwaggerInfo.BasePath = path + BasePath
}

// SwaggerOptions swagger config options. See https://github.com/swaggo/gin-swagger?tab=readme-ov-file#configuration
func (env *Environment) SwaggerOptions() []func(config *ginSwagger.Config) {
	return []func(config *ginSwagger.Config){
		ginSwagger.DefaultModelsExpandDepth(10),
	}
}

func WithTestEnvironment(e *Environment, f func()) {
	realEnv := GetEnvironment()
	env = e
	defer func() {
		env = realEnv
	}()
	f()
}
