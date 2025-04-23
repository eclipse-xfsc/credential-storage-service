package tests

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eclipse-xfsc/credential-storage-service/internal/api"
	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/crypto"
	"github.com/eclipse-xfsc/credential-storage-service/internal/model"
	"github.com/stretchr/testify/mock"

	"github.com/gin-gonic/gin"
)

var registrationEngine *gin.Engine
var registrationEnv *common.Environment

func init() {
	registrationEnv = new(common.Environment)

	key, _ := CreateTestJWK()
	var rawKey interface{}
	key.Raw(&rawKey)
	provider := new(crypto.TestProvider)

	provider.AddKey("test", rawKey)
	crypto.CreateCryptoProvider(true, provider)

	registrationEngine = gin.Default()
	tenantGroup := registrationEngine.Group("/:tenantId")
	accountGroup := tenantGroup.Group("/:account")

	group := accountGroup.Group("/test")
	api.AddRegistrationRoutes(group, registrationEnv)

	group2 := accountGroup.Group("/test2")
	api.AddRemoteRoutes(group2, registrationEnv)
	registrationEnv.SetCryptoNamespace("transit")
	registrationEnv.SetCryptoSignKey("test")
}

func TestAddNewDevice(t *testing.T) {
	mockDb := &SessionMock{}
	mockQ := &QueryMock{}
	mockDb.
		On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(mockQ)
	registrationEnv.SetSession(mockDb)

	jwk, _ := CreateTestJWK()
	pub, _ := jwk.PublicKey()
	authModel := model.AuthModel{
		Account:    "ABCD123",
		TenantId:   "tenant_space",
		Device_Key: &pub,
	}

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/tenant_space/ABCD123/test/register", nil)
	request.Header.Add("Content-Type", "application/json")
	request = request.WithContext(context.WithValue(request.Context(), model.AuthModelKey, authModel))
	if err != nil {
		t.Error()
	}

	registrationEngine.ServeHTTP(recorder, request)

	if recorder.Result().StatusCode != 200 {
		body := recorder.Result().Body
		data, err := io.ReadAll(body)
		if err != nil {
			t.Error(err)
		}
		t.Error("Here should be a 200", string(data))
	}
}

func TestGetSession(t *testing.T) {
	mockDb := &SessionMock{}
	mockQ := &QueryMock{}
	mockDb.
		On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(mockQ)
	registrationEnv.SetSession(mockDb)

	jwk, _ := CreateTestJWK()
	pub, _ := jwk.PublicKey()
	authModel := model.AuthModel{
		Account:    "ABCD123",
		TenantId:   "tenant_space",
		Device_Key: &pub,
	}

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/tenant_space/ABCD123/test2/session", nil)
	request.Header.Add("Content-Type", "application/json")
	request = request.WithContext(context.WithValue(request.Context(), model.AuthModelKey, authModel))
	if err != nil {
		t.Error()
	}

	registrationEngine.ServeHTTP(recorder, request)

	if recorder.Result().StatusCode != 200 {
		t.Error("Here should be a 200")
	}
}
