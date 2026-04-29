package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bluebell/controller"
	"bluebell/pkg/jwt"
	"bluebell/settings"

	"github.com/gin-gonic/gin"
)

type pingResp struct {
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
	Data struct {
		Message  string `json:"message"`
		Username string `json:"username"`
	} `json:"data"`
}

type errorResp struct {
	Code int `json:"code"`
	Msg  any `json:"msg"`
}

func setupTestConfig() {
	settings.GlobalConfig = &settings.Config{
		JWT: settings.JWTConfig{
			Secret:        "test-secret",
			ExpireSeconds: 60,
		},
	}
}

func TestJWTAuthMiddlewareAnonymous(t *testing.T) {
	setupTestConfig()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/ping", JWTAuthMiddleware(), func(c *gin.Context) {
		controller.ResponseSuccess(c, gin.H{"message": "pong"})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp errorResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Code != int(controller.CodeNeedLogin) {
		t.Fatalf("expected code %d, got %d", controller.CodeNeedLogin, resp.Code)
	}
}

func TestJWTAuthMiddlewareWithToken(t *testing.T) {
	setupTestConfig()
	gin.SetMode(gin.TestMode)

	claims := &jwt.Myclaims{UserID: 1, Username: "bob"}
	token, err := jwt.GenToken(claims)
	if err != nil {
		t.Fatalf("GenToken error: %v", err)
	}

	r := gin.New()
	r.GET("/ping", JWTAuthMiddleware(), func(c *gin.Context) {
		userID, _ := c.Get(controller.CtxUserIDKey)
		username, _ := c.Get(controller.CtxUsernameKey)
		controller.ResponseSuccess(c, gin.H{"message": "pong", "user_id": userID, "username": username})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp pingResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Code != int(controller.CodeSuccess) {
		t.Fatalf("expected code %d, got %d", controller.CodeSuccess, resp.Code)
	}
	if resp.Data.Message != "pong" {
		t.Fatalf("expected message pong, got %s", resp.Data.Message)
	}
	if resp.Data.Username != "bob" {
		t.Fatalf("expected username bob, got %s", resp.Data.Username)
	}
}

func TestJWTAuthMiddlewareInvalidToken(t *testing.T) {
	setupTestConfig()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/ping", JWTAuthMiddleware(), func(c *gin.Context) {
		controller.ResponseSuccess(c, gin.H{"message": "pong"})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp errorResp
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if resp.Code != int(controller.CodeInvalidToken) {
		t.Fatalf("expected code %d, got %d", controller.CodeInvalidToken, resp.Code)
	}
}
