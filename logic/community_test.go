package logic

import (
	"errors"
	"testing"

	"bluebell/models"
)

func resetCommunityDeps(t *testing.T) {
	t.Helper()
	origGetList := getCommunityListFromMySQL
	origGetDetail := getCommunityDetailByIDFromMySQL
	t.Cleanup(func() {
		getCommunityListFromMySQL = origGetList
		getCommunityDetailByIDFromMySQL = origGetDetail
	})
}

func TestGetCommunityList(t *testing.T) {
	tests := []struct {
		name    string
		data    []*models.Community
		err     error
		wantLen int
	}{
		{
			name: "success",
			data: []*models.Community{
				{ID: 1, Name: "go"},
				{ID: 2, Name: "vue"},
			},
			wantLen: 2,
		},
		{
			name: "mysql error",
			err:  errors.New("select failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCommunityDeps(t)
			getCommunityListFromMySQL = func() ([]*models.Community, error) {
				return tt.data, tt.err
			}

			got, err := GetCommunityList()
			if (err != nil) != (tt.err != nil) {
				t.Fatalf("GetCommunityList error = %v, wantErr %v", err, tt.err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("len(GetCommunityList) = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestGetCommunityDetail(t *testing.T) {
	resetCommunityDeps(t)
	getCommunityDetailByIDFromMySQL = func(id int64) (*models.CommunityDetail, error) {
		if id != 10 {
			t.Fatalf("id = %d, want 10", id)
		}
		return &models.CommunityDetail{ID: id, Name: "go"}, nil
	}

	got, err := GetCommunityDetail(10)
	if err != nil {
		t.Fatalf("GetCommunityDetail error: %v", err)
	}
	if got.ID != 10 || got.Name != "go" {
		t.Fatalf("unexpected detail: %#v", got)
	}
}
