package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPostJSONUsesStringIDs(t *testing.T) {
	post := Post{
		ID:          45193082978701312,
		AuthorID:    42651172753903616,
		CommunityID: 3,
		Title:       "HUST牛逼",
		Content:     "打得好",
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	text := string(data)
	if !strings.Contains(text, `"id":"45193082978701312"`) {
		t.Fatalf("id not encoded as string: %s", text)
	}
	if !strings.Contains(text, `"author_id":"42651172753903616"`) {
		t.Fatalf("author_id not encoded as string: %s", text)
	}
}
