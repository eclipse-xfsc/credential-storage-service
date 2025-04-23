package tests

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eclipse-xfsc/microservice-core-go/pkg/logr"

	"github.com/eclipse-xfsc/credential-storage-service/internal/common"
	"github.com/eclipse-xfsc/credential-storage-service/internal/crypto"
	"github.com/eclipse-xfsc/credential-storage-service/internal/middleware"
	"github.com/stretchr/testify/mock"

	"github.com/gin-gonic/gin"
)

var authEnv *common.Environment
var authEngine *gin.Engine

func init() {
	authEnv = new(common.Environment)
	logger, err := logr.New("info", true, nil)
	if err != nil {
		log.Fatalf("failed to init logger: %t", err)
	}

	authEnv.SetLogger(*logger)
	key, _ := CreateTestJWK()
	var rawKey interface{}
	key.Raw(&rawKey)

	provider := new(crypto.TestProvider)

	provider.AddKey("test", rawKey)

	crypto.CreateCryptoProvider(true, provider)
	authEnv.SetCryptoNamespace("transit")
	authEnv.SetCryptoSignKey("test")

	authEngine = gin.Default()

	tenantGroup := authEngine.Group("/:tenantId")
	accGroup := tenantGroup.Group("/:account")

	group := accGroup.Group("/test")
	group.Use(middleware.Auth(authEnv, false, false))

	group.GET("", func(c *gin.Context) {
		c.JSON(200, "test")
	})

	group2 := accGroup.Group("/test2")
	k, _ := CreateTestJWK()
	group2.Use(middleware.AuthTestModel(&k))
	group2.Use(middleware.Auth(authEnv, false, false))

	group2.GET("", func(c *gin.Context) {
		c.JSON(200, "test")
	})

	group3 := accGroup.Group("/test3")
	group3.Use(middleware.AuthModel())
	group3.Use(middleware.SelfSignedAuth(authEnv))

	group3.GET("", func(c *gin.Context) {
		c.JSON(200, "test")
	})

}

func TestTokenCheckupWithWrongRoute(t *testing.T) {
	mockDb := &SessionMock{}
	mockQ := &QueryMock{}
	mockDb.
		On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(mockQ)
	authEnv.SetSession(mockDb)
	recorder := httptest.NewRecorder()

	request, err := http.NewRequest("GET", "/tenant_space/ABCD123/test", nil)

	if err != nil {
		t.Error()
	}

	authEngine.ServeHTTP(recorder, request)

	if recorder.Result().StatusCode != 403 {
		t.Error("Here should be a 403")
	}

	if recorder.Code == 403 {
		body := recorder.Body.String()
		var result map[string]string
		err := json.Unmarshal([]byte(body), &result)

		if err != nil {
			t.Error(err)
		}

		if result["message"] != middleware.RouteDataInvald {
			t.Error("Result Message is wrong.")
		}
	} else {
		t.Error("Result should be 400, but was something else.")
	}
}

func TestTokenCheckupWithoutToken(t *testing.T) {
	mockDb := &SessionMock{}
	mockQ := &QueryMock{}
	mockDb.
		On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(mockQ)
	authEnv.SetSession(mockDb)
	recorder := httptest.NewRecorder()

	request, err := http.NewRequest("GET", "/tenant_space/ABCD123/test2", nil)

	if err != nil {
		t.Error()
	}

	authEngine.ServeHTTP(recorder, request)

	if recorder.Result().StatusCode != 403 {
		t.Error("Here should be a 403")
	}

	if recorder.Code == 403 {
		body := recorder.Body.String()
		var result map[string]string
		err := json.Unmarshal([]byte(body), &result)

		if err != nil {
			t.Error(err)
		}

		if result["message"] != middleware.TokenInvalid {
			t.Error("Result Message is wrong.")
		}
	} else {
		t.Error("Result should be 400, but was something else.")
	}
}

//func TestTokenCheckupWithToken(t *testing.T) {
//	var nonce = "5555"
//
//	mockDb := &SessionMock{}
//	mockQ := &QueryMock{}
//	mockQ.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
//		arg := args.String(0)
//		print(arg)
//		arg = nonce
//	})
//	mockDb.
//		On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//		Return(mockQ)
//	authEnv.SetSession(mockDb)
//	k, err := CreateTestJWK()
//	if err != nil {
//		t.Error()
//	}
//
//	session := authEnv.GetSession()
//
//	queryString := `UPDATE tenant_space.credentials SET nonce=? WHERE accountPartition=? AND
//	region=? AND
//	country=? AND
//	account=?`
//
//	err2 := session.Query(queryString,
//		b64.StdEncoding.EncodeToString([]byte(nonce)),
//		authEnv.GetAccountPartition("ABCD123"),
//		authEnv.GetRegion(),
//		authEnv.GetCountry(),
//		"ABCD123").WithContext(context.Background()).Exec()
//
//	if err2 != nil {
//		t.Error()
//	}
//
//	queryString = `SELECT nonce FROM tenant_space.credentials WHERE accountPartition=? AND
//																				region=? AND
//																				country=? AND
//																				account=? LIMIT 1`
//
//	query := session.Query(queryString,
//		authEnv.GetAccountPartition("ABCD123"),
//		authEnv.GetRegion(),
//		authEnv.GetCountry(),
//		"ABCD123")
//
//	err = query.Consistency(gocql.LocalQuorum).Scan(&nonce)
//
//	if err != nil {
//		t.Error()
//	}
//
//	b, err := CreateToken(k, "/tenant_space/ABCD123/test2", "ABCD123", nonce, true)
//	logrus.Info(string(b))
//	if err != nil {
//		t.Error()
//	}
//
//	recorder := httptest.NewRecorder()
//
//	request, err := http.NewRequest("GET", "/tenant_space/ABCD123/test2", nil)
//	request.Header.Add("Authorization", "Bearer "+string(b))
//	if err != nil {
//		t.Error()
//	}
//
//	authEngine.ServeHTTP(recorder, request)
//
//	if recorder.Result().StatusCode != 200 {
//		body := recorder.Result().Body
//		data, err := io.ReadAll(body)
//		if err != nil {
//			t.Error(err)
//		}
//		t.Error("Here should be a 200", string(data))
//	}
//}

func TestLockedAccount(t *testing.T) {
	//TODO
}

func TestSelfSignedAuth(t *testing.T) {
	jwk, err := CreateTestJWK()

	if err != nil {
		t.Error()
	}

	tok, err := CreateSelfSignedToken(jwk, "/tenant_space/ABCD123/test3", "ABCD123")

	if err != nil {
		t.Error()
	}

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/tenant_space/ABCD123/test3", nil)
	if err != nil {
		t.Error()
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+string(tok))
	authEngine.ServeHTTP(recorder, request)
	if recorder.Result().StatusCode != 200 {
		t.Error("Here should be a 200")
	}
}

func TestSelfSignedAuthWithoutBearer(t *testing.T) {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/tenant_space/ABCD123/test3", nil)
	request.Header.Add("Content-Type", "application/json")
	authEngine.ServeHTTP(recorder, request)
	if recorder.Result().StatusCode != 403 {
		t.Error("Here should be a 403")
	}
}
