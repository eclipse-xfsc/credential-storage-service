package main

import (
	"context"
	"path"
	"path/filepath"

	"github.com/eclipse-xfsc/credential-storage-service/internal/api"
	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/connection"
	"github.com/eclipse-xfsc/credential-storage-service/internal/crypto"
	"github.com/eclipse-xfsc/credential-storage-service/internal/middleware"
	core "github.com/eclipse-xfsc/crypto-provider-core"

	"os"

	"github.com/eclipse-xfsc/credential-storage-service/internal/config"
	"github.com/eclipse-xfsc/credential-storage-service/internal/event"

	"github.com/eclipse-xfsc/crypto-provider-core/types"
	logPkg "github.com/eclipse-xfsc/microservice-core-go/pkg/logr"
	serverPkg "github.com/eclipse-xfsc/microservice-core-go/pkg/server"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var env *common.Environment

func init() {
	env = common.GetEnvironment()

	err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %t", err)
	}

	currentConf := config.CurrentStorageConfig

	logger, err := logPkg.New(currentConf.LogLevel, currentConf.IsDev, nil)
	if err != nil {
		log.Fatalf("failed to init logger: %t", err)
	}

	env.SetLogger(*logger)
	env.SetCryptoNamespace(currentConf.Crypto.Namespace)
	env.SetCryptoSignKey(currentConf.Crypto.SignKey)
	env.SetMode(currentConf.Mode)
	env.SetUnitTestModeOn(currentConf.UnitTestModeOn)
	env.SetHealthy(true)
}

func startDbConnection() error {
	// Establish connections
	dbSession, err := connection.Connection()
	if err != nil {
		log.Fatal("Database could not be connected")
		return err
	} else {
		env.SetSession(dbSession)
		log.Info("Database connected")
	}
	return nil
}

func refineRoutes(rg *gin.RouterGroup) {
	storageGroup := rg.Group("/storage")
	accountGroup := storageGroup.Group("/:account")

	switch env.GetMode() {
	case "REMOTE":
		addRemoteRouterGroup(accountGroup)
	case "DIRECT":
		addDirectRouterGroup(accountGroup)
	}
}

func addDirectRouterGroup(rg *gin.RouterGroup) {
	env.SetContentType("application/json")

	rg.Use(middleware.AuthModel())
	credentialGroup := rg.Group("/credentials")
	credentialGroup.Use(middleware.AuthModel())
	api.AddCredentialRoutes(credentialGroup, env)
	presentationGroup := rg.Group("/presentations")
	presentationGroup.Use(middleware.AuthModel())
	api.AddPresentationRoutes(presentationGroup, env)
}

func addRemoteRouterGroup(rg *gin.RouterGroup) {
	env.SetContentType("application/jose")

	deviceGroup := rg.Group("/device")
	remoteGroup := deviceGroup.Group("/remote")
	remoteGroup.Use(middleware.AuthModel())
	remoteGroup.Use(middleware.Auth(env, false, false))
	api.AddRemoteRoutes(remoteGroup, env)

	registrationGroup := deviceGroup.Group("/registration")
	registrationGroup.Use(middleware.AuthModel())
	registrationGroup.Use(middleware.SelfSignedAuth(env))

	api.AddRegistrationRoutes(registrationGroup, env)

	recoverGroup := deviceGroup.Group("/recovery")
	recoverGroup.Use(middleware.AuthModel())
	recoverGroup.Use(middleware.Auth(env, true, true))

	api.AddRecoverRoutes(recoverGroup, env)

	credentialGroup := rg.Group("/credentials")

	credentialGroup.Use(middleware.AuthModel())
	credentialGroup.Use(middleware.Auth(env, false, true))

	api.AddCredentialRoutes(credentialGroup, env)

	presentationGroup := rg.Group("/presentations")

	presentationGroup.Use(middleware.AuthModel())
	presentationGroup.Use(middleware.Auth(env, false, true))

	api.AddPresentationRoutes(presentationGroup, env)
}

func startServer() error {
	server := serverPkg.New(env, config.CurrentStorageConfig.ServerMode)
	server.Add(refineRoutes)

	// Run server
	return server.Run(config.CurrentStorageConfig.ListenPort)
}

func initializeCrypto() error {
	var err error
	var exists bool

	var engine types.CryptoProvider
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exPath := filepath.Dir(ex)
	enginePath := config.CurrentStorageConfig.Crypto.PluginPath

	if !config.CurrentStorageConfig.UnitTestModeOn {
		if config.CurrentStorageConfig.Profile == "DEBUG:LOCAL" {
			engine = core.CreateCryptoEngine(path.Join(exPath, ".engines/.local/crypto-provider-local-plugin.so"))
		} else {
			if config.CurrentStorageConfig.Profile == "DEBUG:VAULT" {
				engine = core.CreateCryptoEngine(path.Join(exPath, ".engines/.vault/crypto-provider-hashicorp-vault-plugin.so"))
			} else {
				if _, err := os.Stat(enginePath); err == nil || os.IsExist(err) {
					env.GetLogger().Debug("Load Engine...")
					engine = core.CreateCryptoEngine(enginePath)
				} else {
					panic("Engine not exists.")
				}
			}
		}
	} else {
		engine = new(crypto.TestProvider)
	}

	crypto.CreateCryptoProvider(config.CurrentStorageConfig.UnitTestModeOn, engine)

	ctx := types.CryptoContext{
		Namespace: env.GetCryptoNamespace(),
		Context:   context.Background(),
		Group:     common.StorageCryptoContext,
	}

	if exists, err = env.GetCryptoProvider().IsCryptoContextExisting(ctx); err == nil && !exists {
		err = env.GetCryptoProvider().CreateCryptoContext(ctx)
	}

	if err != nil {
		return err
	}

	identifier := types.CryptoIdentifier{
		KeyId:         env.GetCryptoSignKey(),
		CryptoContext: ctx,
	}

	if exists, err = env.GetCryptoProvider().IsKeyExisting(identifier); err == nil && !exists {
		err = env.GetCryptoProvider().GenerateKey(types.CryptoKeyParameter{
			Identifier: identifier,
			KeyType:    types.Ecdsap256,
		})
	}

	return err
}

// @title			Storage service API
// @version		1.0
// @description	Service responsible for storing and retrieving credentials and presentations
// @license.name	Apache 2.0
// @license.url	http://www.apache.org/licenses/LICENSE-2.0.html
// @host			localhost:8080

func main() {
	logger := env.GetLogger()

	err := initializeCrypto()
	if err != nil {
		logger.Error(err, "Failed initializing crypto keys")
		os.Exit(1)
	}

	if err := startDbConnection(); err != nil {
		return
	}

	if config.CurrentStorageConfig.Messaging.Enabled {
		if err := event.StartCloudEvents(); err != nil {
			return
		}
	}

	if env.GetMode() == "REMOTE" || env.GetMode() == "DIRECT" {
		if err := startServer(); err != nil {
			return
		}
	}

	defer env.GetSession().Close()
}
