package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eclipse-xfsc/microservice-core-go/pkg/logr"

	"github.com/eclipse-xfsc/credential-storage-service/internal/api"
	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	cryptoProvider "github.com/eclipse-xfsc/credential-storage-service/internal/crypto"
	commonHandler "github.com/eclipse-xfsc/credential-storage-service/internal/handlers/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/middleware"
	"github.com/eclipse-xfsc/credential-storage-service/internal/model"

	"github.com/stretchr/testify/mock"

	"github.com/eclipse-xfsc/crypto-provider-core/types"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/sirupsen/logrus"
)

var credentialEngine *gin.Engine
var credentialEnv *common.Environment

func init() {
	credentialEnv = new(common.Environment)

	logger, err := logr.New("info", true, nil)
	if err != nil {
		log.Fatalf("failed to init logger: %t", err)
	}

	credentialEnv.SetLogger(*logger)
	privKey, err := CreateTestJWK()

	if err != nil {
		logrus.Fatal("Can't create test jwk.")
	}

	deviceKey, err := privKey.PublicKey()

	if err != nil {
		logrus.Fatal()
	}

	credentialEngine = gin.Default()
	tenantGroup := credentialEngine.Group("/:tenantId")
	accountGroup := tenantGroup.Group("/:account")

	group := accountGroup.Group("/test")
	group.Use(middleware.AuthTestModel(&deviceKey))
	credentialEnv.SetContentType("application/jose")
	api.AddCredentialRoutes(group, credentialEnv)

	cryptoProvider.CreateCryptoProvider(true, nil)
	credentialEnv.SetCryptoNamespace("unique")

	parameter := types.CryptoKeyParameter{
		Identifier: types.CryptoIdentifier{
			KeyId: "ABCD123",
			CryptoContext: types.CryptoContext{
				Namespace: credentialEnv.GetCryptoNamespace(),
				Context:   context.Background(),
				Group:     common.StorageCryptoContext,
			},
		},
		KeyType: types.Aes256GCM,
	}

	cryptoProvider.GetCryptoProvider().GenerateKey(parameter)
}

//func StartConnection(env *common.Environment) {
//
//	_, filename, _, _ := runtime.Caller(0)
//	dir := path.Join(path.Dir(filename), "..", "..")
//	err := os.Chdir(dir)
//
//	if err != nil {
//		logrus.Fatal()
//	}
//
//	data := path.Join(dir, ".env")
//
//	viper.SetConfigFile(data)
//	viper.ReadInConfig()
//	viper.AutomaticEnv()
//
//	if env.GetSession() == nil {
//
//		env.SetSession(&SessionMock{})
//	}
//}

func TestAddCredentialNoBody(t *testing.T) {
	common.WithTestEnvironment(credentialEnv, func() {
		credentialEnv.SetContentType("application/jose")
		recorder := httptest.NewRecorder()
		request, err := http.NewRequest("PUT", "/tenant_space/ABCD123/test/123", nil)
		request.Header.Add("Content-Type", "application/jose")
		if err != nil {
			t.Error()
		}

		credentialEngine.ServeHTTP(recorder, request)

		if recorder.Code == 400 {
			body := recorder.Body.String()
			var result map[string]string
			err := json.Unmarshal([]byte(body), &result)

			if err != nil {
				t.Error(err)
			}

			if result["message"] != commonHandler.NoBodyError {
				t.Error("Result Message is wrong.")
			}

		} else {
			t.Error("Result should be 400, but was something else.")
		}
	})
}

func TestWrongContentType(t *testing.T) {
	common.WithTestEnvironment(credentialEnv, func() {

		recorder := httptest.NewRecorder()
		request, err := http.NewRequest("PUT", "/tenant_space/ABCD123/test/123", nil)
		request.Header.Add("Content-Type", "application/json")
		if err != nil {
			t.Error()
		}

		credentialEngine.ServeHTTP(recorder, request)

		if recorder.Code == 400 {
			body := recorder.Body.String()
			var result map[string]string
			err := json.Unmarshal([]byte(body), &result)

			if err != nil {
				t.Error(err)
			}

			if result["message"] != commonHandler.WrongContentType {
				t.Error("Result Message is wrong.")
			}
		} else {
			t.Error("Result should be 400, but was something else.")
		}
	})
}

//func TestAddCredentialWithEmptyBody(t *testing.T) {
//
//	credential := "credential"
//
//	jwk, _ := CreateTestJWK()
//	var rawKey rsa.PublicKey
//	jwk.Raw(&rawKey)
//	encrypted, err := jwe.Encrypt([]byte(credential), jwe.WithKey(jwa.RSA1_5, &rawKey)) // usage of private results in empty msg
//	if err != nil {
//		t.Error(err)
//	}
//	recorder := httptest.NewRecorder()
//
//	request, err := http.NewRequest("PUT", "/tenant_space/ABCD123/test/123", bytes.NewReader(encrypted))
//	request.Header.Add("Content-Type", "application/jose")
//
//	if err != nil {
//		t.Error()
//	}
//
//	credentialEngine.ServeHTTP(recorder, request)
//
//	if recorder.Code == 400 {
//		body := recorder.Body.String()
//		var result map[string]string
//		err := json.Unmarshal([]byte(body), &result)
//
//		if err != nil {
//			t.Error(err)
//		}
//
//		if result["message"] != commonHandler.BodyParseError {
//			t.Error("Result Message is wrong.")
//		}
//	} else {
//		t.Error("Result should be 400, but was something else.")
//	}
//}

func TestAddCredential(t *testing.T) {
	mockDb := &SessionMock{}
	mockQ := &QueryMock{}
	mockDb.
		On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(mockQ)
	credentialEnv.SetSession(mockDb)
	jwk, _ := CreateTestJWK()

	var rawKey interface{}
	k, _ := jwk.PublicKey()
	k.Raw(&rawKey)

	//credential := "credential"

	/*encrypted, err := //jwe.Encrypt([]byte(credential), jwe.WithKey(jwa.ECDH_ES_A256KW, rawKey))

	if err != nil {
		t.Error()
	}*/
	tok := "eyJyZWNpcGllbnQiOiIwNDdmMzJiYzlhMTA5ZWVmMzI3OTczOGFlMDEwM2EyNTU4YzI3MTE1MWFlODA4NDRiMjEyNjc4ZWQ5MGI3YTdhIiwiY3R5IjoiSldUIiwiZXBrIjp7Imt0eSI6IkVDIiwiY3J2IjoiUC0yNTYiLCJ4IjoiR3lETlV2Y2gwcHVxVUY4cUNMM1ZSWUVRUE84ZmdXd0N5eXFBbFpoa0RRSSIsInkiOiJGeHByMnNESTdxV2Y1R21MOHg3M1RybUYzeFFzNlctMG9WeEdrUWRReWRVIn0sImVuYyI6IkEyNTZHQ00iLCJhbGciOiJFQ0RILUVTK0EyNTZLVyJ9.vMoj_MHEbwJ7XQ-0t6n2Wmj4-3CXVHl4lawN8KSKESypmg-SKFjmag.dqorQWUEp1_ThhGx.ON1vEAETJCnpzI3KdDQTbDjjZS08CXpa_BQkWKSYODEeCPXmUaGkbDzTY1nldYfwAT5OUzyi4fjh4j9pzYYklhKDWknSpRD5plXcX6qG20hpMTjDiEGKGpQSDfXlE4_yMntJKfiquWc0Cw9HJ6E9m800CSAxuLqgmTGa9F-2mFHKog.4kHTg5dUSFcUf1jwDD0fbQ"
	encrypted := []byte(tok)

	recorder := httptest.NewRecorder()

	request, err := http.NewRequest("PUT", "/tenant_space/ABCD123/test/123", bytes.NewReader(encrypted))
	request.Header.Add("Content-Type", "application/jose")
	if err != nil {
		t.Error()
	}

	credentialEngine.ServeHTTP(recorder, request)

	if recorder.Result().StatusCode != 200 {
		t.Error()
	}

	body := recorder.Body.String()

	jwk.Raw(&rawKey)

	plain, err := jwe.Decrypt([]byte(body), jwe.WithKey(jwa.ECDH_ES_A256KW, rawKey))

	if err != nil {
		t.Error("Message cant be decoded")
	}
	var j map[string]interface{}
	err = json.Unmarshal(plain, &j)

	if err != nil {
		t.Error("Format not serializable")
	}

	_, ok := j["nonce"]

	if !ok {
		t.Error()
	}

	_, ok = j["expire"]

	if !ok {
		t.Error()
	}
}

func TestGetCredential(t *testing.T) {
	common.WithTestEnvironment(credentialEnv, func() {

		mockDb := &SessionMock{}
		mockQ := &QueryMock{}
		mockQ.
			On("Scan", mock.Anything).
			Run(func(args mock.Arguments) {
				arg := args.Get(0).(*model.GetCredentialModel)
				arg.Credentials = map[string]interface{}{"cred1": "cred1"}
				arg.Receipt = "receipt"
			})
		mockDb.
			On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(mockQ)
		credentialEnv.SetSession(mockDb)
		jwk, _ := CreateTestJWK()

		var rawKey interface{}
		k, _ := jwk.PublicKey()
		k.Raw(&rawKey)

		recorder := httptest.NewRecorder()

		request, err := http.NewRequest("GET", "/tenant_space/ABCD123/test", nil)
		request.Header.Add("Content-Type", "application/jose")
		if err != nil {
			t.Error()
		}

		credentialEngine.ServeHTTP(recorder, request)

		if recorder.Result().StatusCode != 200 {
			t.Error()
		}

		body := recorder.Body.String()

		var j map[string]interface{}
		err = json.Unmarshal([]byte(body), &j)
		print(body)
		if err != nil {
			t.Error()
		}

		_, ok := j["receipt"]

		if !ok {
			t.Error()
		}

		_, ok = j["credentials"]

		if !ok {
			t.Error()
		}
	})
}

func TestDeleteCredential(t *testing.T) {
	mockDb := &SessionMock{}
	mockQ := &QueryMock{}
	mockDb.
		On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(mockQ)
	credentialEnv.SetSession(mockDb)
	jwk, _ := CreateTestJWK()

	var rawKey interface{}
	k, _ := jwk.PublicKey()
	k.Raw(&rawKey)

	recorder := httptest.NewRecorder()

	request, err := http.NewRequest("DELETE", "/tenant_space/ABCD123/test/123", nil)
	if err != nil {
		t.Error()
	}

	credentialEngine.ServeHTTP(recorder, request)

	if recorder.Result().StatusCode != 200 {
		t.Error()
	}

	body := recorder.Body.String()

	jwk.Raw(&rawKey)

	plain, err := jwe.Decrypt([]byte(body), jwe.WithKey(jwa.ECDH_ES_A256KW, rawKey))

	if err != nil {
		t.Error("Message cant be decoded")
	}
	var j map[string]interface{}
	err = json.Unmarshal(plain, &j)

	if err != nil {
		t.Error("Format not serializable")
	}

	_, ok := j["nonce"]

	if !ok {
		t.Error()
	}

	_, ok = j["expire"]

	if !ok {
		t.Error()
	}
}
