package messaging

import "github.com/eclipse-xfsc/nats-message-library/common"

const (
	StorePresentationType = "storage.service.presentation"
	StoreCredentialType   = "storage.service.credential"
)

type StorageServiceStoreMessage struct {
	common.Request
	AccountId   string `json:"accountId"`
	Type        string `json:"type"`
	Payload     []byte `json:"payload"`
	ContentType string `json:"contentType"`
	Id          string `json:"id"`
}
