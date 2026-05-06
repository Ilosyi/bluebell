package jwt

import (
	"testing"
	"time"

	"bluebell/settings"
	gojwt "github.com/golang-jwt/jwt/v5"
)

func TestGenAndParseToken(t *testing.T) {
	settings.GlobalConfig = &settings.Config{
		JWT: settings.JWTConfig{
			Secret:        "test-secret",
			ExpireSeconds: 60,
		},
	}

	claims := &Myclaims{
		UserID:   123,
		Username: "alice",
	}

	token, err := GenToken(claims)
	if err != nil {
		t.Fatalf("GenToken error: %v", err)
	}
	if token == "" {
		t.Fatal("GenToken returned empty token")
	}

	parsed, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if parsed.UserID != claims.UserID {
		t.Fatalf("expected user_id %d, got %d", claims.UserID, parsed.UserID)
	}
	if parsed.Username != claims.Username {
		t.Fatalf("expected username %s, got %s", claims.Username, parsed.Username)
	}
}

func TestLoadConfigValidation(t *testing.T) {
	orig := settings.GlobalConfig
	t.Cleanup(func() {
		settings.GlobalConfig = orig
	})

	settings.GlobalConfig = nil
	if _, _, err := loadConfig(); err == nil {
		t.Fatal("expected nil config error")
	}

	settings.GlobalConfig = &settings.Config{}
	if _, _, err := loadConfig(); err == nil {
		t.Fatal("expected empty secret error")
	}

	settings.GlobalConfig = &settings.Config{
		JWT: settings.JWTConfig{Secret: "test-secret"},
	}
	_, expire, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}
	if expire != 24*time.Hour {
		t.Fatalf("expire = %v, want 24h", expire)
	}
}

func TestParseTokenErrors(t *testing.T) {
	orig := settings.GlobalConfig
	t.Cleanup(func() {
		settings.GlobalConfig = orig
	})
	settings.GlobalConfig = &settings.Config{
		JWT: settings.JWTConfig{Secret: "test-secret", ExpireSeconds: 60},
	}

	if _, err := ParseToken("bad.token"); err == nil {
		t.Fatal("expected malformed token error")
	}

	token, err := GenToken(&Myclaims{
		UserID:   1,
		Username: "alice",
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	})
	if err != nil {
		t.Fatalf("GenToken error: %v", err)
	}
	if _, err := ParseToken(token); err == nil {
		t.Fatal("expected expired token error")
	}
}
