package tests

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/eclipse-xfsc/credential-storage-service/internal/api"
	"github.com/eclipse-xfsc/credential-storage-service/internal/common"

	"github.com/gin-gonic/gin"
)

var healthEnv *common.Environment

func TestHealth(t *testing.T) {
	healthEnv = new(common.Environment)
	engine := gin.Default()
	group := engine.Group("/test")

	api.AddHealth(group, healthEnv)

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/test/health", nil)

	if err != nil {
		t.Error("Request Building failed.")
	}

	engine.ServeHTTP(recorder, request)

	if recorder.Code != 400 {
		t.Error("Health Request shows wrong statuscode. Code was: " + strconv.Itoa(recorder.Code))
	}
}
