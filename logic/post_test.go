package logic

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"bluebell/models"
)

func resetPostDeps(t *testing.T) {
	t.Helper()
	origGenPostID := genPostID
	origCreatePostInMySQL := createPostInMySQL
	origSavePostTimeAndScore := savePostTimeAndScore
	origGetPostByIDFromMySQL := getPostByIDFromMySQL
	origGetUserByIDFromMySQL := getUserByIDFromMySQL
	origGetCommunityDetailByID := getCommunityDetailByID
	origGetPostIDsInOrder := getPostIDsInOrder
	origGetCommunityPostIDsInOrder := getCommunityPostIDsInOrder
	origGetPostListByIDsFromMySQL := getPostListByIDsFromMySQL
	origGetPostVoteData := getPostVoteData
	origGetPostListFromMySQL := getPostListFromMySQL
	t.Cleanup(func() {
		genPostID = origGenPostID
		createPostInMySQL = origCreatePostInMySQL
		savePostTimeAndScore = origSavePostTimeAndScore
		getPostByIDFromMySQL = origGetPostByIDFromMySQL
		getUserByIDFromMySQL = origGetUserByIDFromMySQL
		getCommunityDetailByID = origGetCommunityDetailByID
		getPostIDsInOrder = origGetPostIDsInOrder
		getCommunityPostIDsInOrder = origGetCommunityPostIDsInOrder
		getPostListByIDsFromMySQL = origGetPostListByIDsFromMySQL
		getPostVoteData = origGetPostVoteData
		getPostListFromMySQL = origGetPostListFromMySQL
	})
}

func TestCreatePost(t *testing.T) {
	tests := []struct {
		name      string
		createErr error
		saveErr   error
		wantErr   bool
		wantSave  bool
	}{
		{name: "success", wantSave: true},
		{name: "mysql error", createErr: errors.New("insert failed"), wantErr: true},
		{name: "redis error", saveErr: errors.New("redis failed"), wantErr: true, wantSave: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetPostDeps(t)
			const postID int64 = 88
			saveCalled := false

			genPostID = func() int64 { return postID }
			createPostInMySQL = func(p *models.Post) error {
				if p.ID != postID {
					t.Fatalf("post id = %d, want %d", p.ID, postID)
				}
				return tt.createErr
			}
			savePostTimeAndScore = func(id int64, communityID int64, now time.Time) error {
				saveCalled = true
				if id != postID || communityID != 3 {
					t.Fatalf("unexpected save args: id=%d communityID=%d", id, communityID)
				}
				if now.IsZero() {
					t.Fatal("save time is zero")
				}
				return tt.saveErr
			}

			err := CreatePost(&models.Post{CommunityID: 3, Title: "title", Content: "content"})
			if (err != nil) != tt.wantErr {
				t.Fatalf("CreatePost error = %v, wantErr %v", err, tt.wantErr)
			}
			if saveCalled != tt.wantSave {
				t.Fatalf("save called = %v, want %v", saveCalled, tt.wantSave)
			}
		})
	}
}

func TestGetPostById(t *testing.T) {
	resetPostDeps(t)
	getPostByIDFromMySQL = func(pid int64) (*models.Post, error) {
		if pid != 100 {
			t.Fatalf("pid = %d, want 100", pid)
		}
		return &models.Post{ID: pid, AuthorID: 9, CommunityID: 2, Title: "post"}, nil
	}
	getUserByIDFromMySQL = func(userID int64) (*models.User, error) {
		if userID != 9 {
			t.Fatalf("userID = %d, want 9", userID)
		}
		return &models.User{UserID: userID, Username: "alice"}, nil
	}
	getCommunityDetailByID = func(id int64) (*models.CommunityDetail, error) {
		if id != 2 {
			t.Fatalf("community id = %d, want 2", id)
		}
		return &models.CommunityDetail{ID: id, Name: "go"}, nil
	}
	getPostVoteData = func(ids []string) ([]int64, error) {
		if !reflect.DeepEqual(ids, []string{"100"}) {
			t.Fatalf("vote ids = %#v, want %#v", ids, []string{"100"})
		}
		return []int64{7}, nil
	}

	got, err := GetPostById(100)
	if err != nil {
		t.Fatalf("GetPostById error: %v", err)
	}
	if got.Post.ID != 100 || got.AuthorName != "alice" || got.CommunityDetail.Name != "go" || got.VoteNum != 7 {
		t.Fatalf("unexpected detail: %#v", got)
	}
}

func TestGetPostListNewRoutesByCommunity(t *testing.T) {
	tests := []struct {
		name          string
		communityID   int64
		wantGlobal    bool
		wantCommunity bool
	}{
		{name: "global", wantGlobal: true},
		{name: "community", communityID: 3, wantCommunity: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetPostDeps(t)
			globalCalled := false
			communityCalled := false
			stubPostsByIDs(t)

			getPostIDsInOrder = func(p *models.ParamPostList) ([]string, error) {
				globalCalled = true
				return []string{"11"}, nil
			}
			getCommunityPostIDsInOrder = func(p *models.ParamPostList) ([]string, error) {
				communityCalled = true
				return []string{"11"}, nil
			}

			_, err := GetPostListNew(&models.ParamPostList{CommunityID: tt.communityID, Page: 1, Size: 10, Order: "time"})
			if err != nil {
				t.Fatalf("GetPostListNew error: %v", err)
			}
			if globalCalled != tt.wantGlobal || communityCalled != tt.wantCommunity {
				t.Fatalf("globalCalled=%v communityCalled=%v, want %v/%v", globalCalled, communityCalled, tt.wantGlobal, tt.wantCommunity)
			}
		})
	}
}

func TestGetPostListByIDs(t *testing.T) {
	tests := []struct {
		name      string
		ids       []string
		wantIDs   []int64
		wantLen   int
		wantVotes []int64
	}{
		{
			name:      "empty ids",
			wantLen:   0,
			wantVotes: nil,
		},
		{
			name:      "invalid ids are skipped for mysql query",
			ids:       []string{"10", "bad", "11"},
			wantIDs:   []int64{10, 11},
			wantLen:   2,
			wantVotes: []int64{3, 5, 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetPostDeps(t)
			getPostListByIDsFromMySQL = func(ids []int64) ([]*models.Post, error) {
				if !reflect.DeepEqual(ids, tt.wantIDs) {
					t.Fatalf("ids = %#v, want %#v", ids, tt.wantIDs)
				}
				posts := make([]*models.Post, 0, len(ids))
				for _, id := range ids {
					posts = append(posts, &models.Post{ID: id, AuthorID: 1, CommunityID: 2})
				}
				return posts, nil
			}
			getPostVoteData = func(ids []string) ([]int64, error) {
				if !reflect.DeepEqual(ids, tt.ids) {
					t.Fatalf("vote ids = %#v, want %#v", ids, tt.ids)
				}
				return tt.wantVotes, nil
			}
			getUserByIDFromMySQL = func(userID int64) (*models.User, error) {
				return &models.User{UserID: userID, Username: "alice"}, nil
			}
			getCommunityDetailByID = func(id int64) (*models.CommunityDetail, error) {
				return &models.CommunityDetail{ID: id, Name: "go"}, nil
			}

			got, err := getPostListByIDs(tt.ids)
			if err != nil {
				t.Fatalf("getPostListByIDs error: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("len(getPostListByIDs) = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestGetPostList(t *testing.T) {
	resetPostDeps(t)
	getPostListFromMySQL = func(page, size int64) ([]*models.Post, error) {
		if page != 2 || size != 3 {
			t.Fatalf("page/size = %d/%d, want 2/3", page, size)
		}
		return []*models.Post{
			{ID: 1, AuthorID: 10, CommunityID: 20},
			{ID: 2, AuthorID: 11, CommunityID: 21},
		}, nil
	}
	getUserByIDFromMySQL = func(userID int64) (*models.User, error) {
		if userID == 11 {
			return nil, errors.New("user failed")
		}
		return &models.User{UserID: userID, Username: "alice"}, nil
	}
	getCommunityDetailByID = func(id int64) (*models.CommunityDetail, error) {
		return &models.CommunityDetail{ID: id, Name: "go"}, nil
	}

	got, err := GetPostList(2, 3)
	if err != nil {
		t.Fatalf("GetPostList error: %v", err)
	}
	if len(got) != 1 || got[0].Post.ID != 1 {
		t.Fatalf("unexpected posts: %#v", got)
	}
}

func stubPostsByIDs(t *testing.T) {
	t.Helper()
	getPostListByIDsFromMySQL = func(ids []int64) ([]*models.Post, error) {
		return []*models.Post{{ID: ids[0], AuthorID: 1, CommunityID: 2}}, nil
	}
	getPostVoteData = func(ids []string) ([]int64, error) {
		return []int64{1}, nil
	}
	getUserByIDFromMySQL = func(userID int64) (*models.User, error) {
		return &models.User{UserID: userID, Username: "alice"}, nil
	}
	getCommunityDetailByID = func(id int64) (*models.CommunityDetail, error) {
		return &models.CommunityDetail{ID: id, Name: "go"}, nil
	}
}
