package jwt

import (
	"testing"

	"bluebell/settings"
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
