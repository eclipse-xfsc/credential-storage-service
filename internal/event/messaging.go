package event

import (
	"context"
	"encoding/json"

	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/config"
	handlers "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/model"
	"github.com/eclipse-xfsc/credential-storage-service/internal/services"
	"github.com/eclipse-xfsc/credential-storage-service/pkg/messaging"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/eclipse-xfsc/cloud-event-provider"
	logPkg "github.com/eclipse-xfsc/microservice-core-go/pkg/logr"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	log "github.com/sirupsen/logrus"
)

type StorageMessaging struct {
	client *cloudeventprovider.CloudEventProviderClient
	logger *logPkg.Logger
}

var storagemessaging = new(StorageMessaging)

func StartCloudEvents() error {
	log.Info("start messaging!")
	client, err := cloudeventprovider.New(cloudeventprovider.Config{
		Protocol: cloudeventprovider.ProtocolTypeNats,
		Settings: cloudeventprovider.NatsConfig{
			Url:        config.CurrentStorageConfig.Messaging.Url,
			QueueGroup: config.CurrentStorageConfig.Messaging.QueueGroup,
		},
	}, cloudeventprovider.ConnectionTypeSub, config.CurrentStorageConfig.Messaging.StorageTopic)
	if err != nil {
		log.Fatal(err)
	}

	storagemessaging.client = client

	/*defer func() {
		if err := client.Close(); err != nil {
			log.Error(err)
		}
	}()*/

	go storagemessaging.listen()

	return nil
}

func (s *StorageMessaging) listen() {
	for {
		if err := s.client.SubCtx(context.Background(), handler); err != nil {
			s.logger.Error(err, "error retrieving message")
		}
	}
}

func handler(event event.Event) {
	var newMessage messaging.StorageServiceStoreMessage
	err := json.Unmarshal(event.Data(), &newMessage)
	if err != nil {
		log.Errorf("error occured while unmarshal Message %v: %v", event, err)
	}

	log.Debugf("new Message received: %v", newMessage)

	getType(newMessage, common.GetEnvironment())
}

func getType(msg messaging.StorageServiceStoreMessage, env *common.Environment) {
	logger := env.GetLogger()

	authModel := model.AuthModel{
		Account:  msg.AccountId,
		TenantId: msg.TenantId,
	}

	presentation := false

	if msg.Type == messaging.StorePresentationType {
		presentation = true
	}

	if msg.ContentType == common.EncryptedContentType {
		message, err := jwe.Parse(msg.Payload)
		if err != nil {
			logger.Error(err, "body parse error")
			return
		}
		if message.ProtectedHeaders().Algorithm() != jwa.ECDH_ES_A256KW {
			logger.Error(nil, handlers.InvalidKeyEncryptionAlgorithm)
			return
		}
		if message.ProtectedHeaders().ContentEncryption() != jwa.ContentEncryptionAlgorithm(jwa.A256GCM) {
			logger.Error(nil, handlers.InvalidContentEncryptionAlgorithm)
			return
		}
		recipients := message.Recipients()
		if len(recipients) == 1 {
			if _, err := services.StoreMessage(context.Background(), msg.Id, msg.Payload, authModel, env, presentation); err != nil {
				logger.Error(err, "could not store message")
				return
			}
		}
	} else {
		/*var message map[string]interface{} // DO NOT PARSE, SD-JWT is just a string
		if err := json.Unmarshal(msg.Payload, &message); err != nil {
			logger.Error(err, "not a json body")
			return
		}*/
		if _, err := services.StoreMessage(context.Background(), msg.Id, msg.Payload, authModel, env, presentation); err != nil {
			logger.Error(err, "could not store message")
			return
		}
	}
}
