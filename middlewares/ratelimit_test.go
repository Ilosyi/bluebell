package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bluebell/controller"
	"bluebell/settings"

	"github.com/gin-gonic/gin"
)

func TestRateLimitMiddlewareAllowsRequestWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RateLimitMiddleware(&settings.RateLimitConfig{Enabled: false}))
	r.GET("/ping", func(c *gin.Context) {
		controller.ResponseSuccess(c, gin.H{"message": "pong"})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRateLimitMiddlewareRejectsWhenBucketIsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RateLimitMiddleware(&settings.RateLimitConfig{
		Enabled:  true,
		Rate:     0.1,
		Capacity: 1,
	}))
	r.GET("/ping", func(c *gin.Context) {
		controller.ResponseSuccess(c, gin.H{"message": "pong"})
	})

	firstReq := httptest.NewRequest(http.MethodGet, "/ping", nil)
	firstResp := httptest.NewRecorder()
	r.ServeHTTP(firstResp, firstReq)
	if firstResp.Code != http.StatusOK {
		t.Fatalf("first status = %d, want %d", firstResp.Code, http.StatusOK)
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/ping", nil)
	secondResp := httptest.NewRecorder()
	r.ServeHTTP(secondResp, secondReq)
	if secondResp.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want %d", secondResp.Code, http.StatusTooManyRequests)
	}

	var resp errorResp
	if err := json.Unmarshal(secondResp.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Code != int(controller.CodeTooManyRequests) {
		t.Fatalf("code = %d, want %d", resp.Code, controller.CodeTooManyRequests)
	}
}
