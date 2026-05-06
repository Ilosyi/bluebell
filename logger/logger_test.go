package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"bluebell/settings"
	"github.com/gin-gonic/gin"
)

func TestInit(t *testing.T) {
	origConfig := settings.GlobalConfig
	t.Cleanup(func() {
		settings.GlobalConfig = origConfig
	})

	settings.GlobalConfig = nil
	if err := Init(&settings.LogConfig{Level: "debug"}, "dev"); err != nil {
		t.Fatalf("Init with nil GlobalConfig error: %v", err)
	}

	settings.GlobalConfig = &settings.Config{}
	if err := Init(&settings.LogConfig{Level: "bad-level"}, "release"); err == nil {
		t.Fatal("expected invalid log level error")
	}

	if err := Init(&settings.LogConfig{Level: "debug", Filename: t.TempDir() + "/app.log"}, "release"); err != nil {
		t.Fatalf("Init release error: %v", err)
	}
}

func TestGinMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(GinLogger(), GinRecovery(false))
	r.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusAccepted, "ok")
	})
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok?x=1", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("/ok status = %d", w.Code)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("/panic status = %d", w.Code)
	}
}
