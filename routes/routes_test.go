package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bluebell/settings"
	"github.com/gin-gonic/gin"
)

func TestSetupPing(t *testing.T) {
	orig := settings.GlobalConfig
	t.Cleanup(func() {
		settings.GlobalConfig = orig
	})
	settings.GlobalConfig = &settings.Config{App: settings.AppConfig{Mode: "dev"}}
	gin.SetMode(gin.TestMode)

	r := Setup()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}
