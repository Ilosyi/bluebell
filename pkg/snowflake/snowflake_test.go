package snowflake

import "testing"

func TestInitAndGenID(t *testing.T) {
	if err := Init("2024-01-01", 1); err != nil {
		t.Fatalf("Init error: %v", err)
	}
	first := GenID()
	second := GenID()
	if first == 0 || second == 0 {
		t.Fatalf("generated zero id: %d, %d", first, second)
	}
	if first == second {
		t.Fatalf("generated duplicate ids: %d", first)
	}
}

func TestInitInvalidStartTime(t *testing.T) {
	if err := Init("bad-date", 1); err == nil {
		t.Fatal("expected invalid start time error")
	}
}
