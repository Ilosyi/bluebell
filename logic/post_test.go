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
	origDeletePostIndex := deletePostIndex
	origMovePostCommunityIndex := movePostCommunityIndex
	origGetPostBundleByIDFromMySQL := getPostBundleByIDFromMySQL
	origGetPostIDsInOrder := getPostIDsInOrder
	origGetCommunityPostIDsInOrder := getCommunityPostIDsInOrder
	origGetPostBundlesByIDsFromMySQL := getPostBundlesByIDsFromMySQL
	origGetPostVoteData := getPostVoteData
	origGetPostListFromMySQL := getPostListFromMySQL
	origCountPostsFromMySQL := countPostsFromMySQL
	origSearchPostBundlesFromMySQL := searchPostBundlesFromMySQL
	origCountSearchPostsFromMySQL := countSearchPostsFromMySQL
	origCountPostsInCommunity := countPostsInCommunity
	origGetPostForManageByID := getPostForManageByID
	origGetMyPostBundles := getMyPostBundles
	origCountMyPosts := countMyPosts
	origUpdatePostInMySQL := updatePostInMySQL
	origPublishPostInMySQL := publishPostInMySQL
	origDeletePostInMySQL := deletePostInMySQL
	t.Cleanup(func() {
		genPostID = origGenPostID
		createPostInMySQL = origCreatePostInMySQL
		savePostTimeAndScore = origSavePostTimeAndScore
		deletePostIndex = origDeletePostIndex
		movePostCommunityIndex = origMovePostCommunityIndex
		getPostBundleByIDFromMySQL = origGetPostBundleByIDFromMySQL
		getPostIDsInOrder = origGetPostIDsInOrder
		getCommunityPostIDsInOrder = origGetCommunityPostIDsInOrder
		getPostBundlesByIDsFromMySQL = origGetPostBundlesByIDsFromMySQL
		getPostVoteData = origGetPostVoteData
		getPostListFromMySQL = origGetPostListFromMySQL
		countPostsFromMySQL = origCountPostsFromMySQL
		searchPostBundlesFromMySQL = origSearchPostBundlesFromMySQL
		countSearchPostsFromMySQL = origCountSearchPostsFromMySQL
		countPostsInCommunity = origCountPostsInCommunity
		getPostForManageByID = origGetPostForManageByID
		getMyPostBundles = origGetMyPostBundles
		countMyPosts = origCountMyPosts
		updatePostInMySQL = origUpdatePostInMySQL
		publishPostInMySQL = origPublishPostInMySQL
		deletePostInMySQL = origDeletePostInMySQL
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
	// 这里把 MySQL 聚合查询和 Redis 票数查询都 stub 掉，
	// 重点验证 logic 层是否能把两部分数据正确合并成详情响应。
	getPostBundleByIDFromMySQL = func(pid int64) (*models.ApiPostDetail, error) {
		if pid != 100 {
			t.Fatalf("pid = %d, want 100", pid)
		}
		return &models.ApiPostDetail{
			AuthorName: "alice",
			Post:       &models.Post{ID: pid, AuthorID: 9, CommunityID: 2, Title: "post"},
			CommunityDetail: &models.CommunityDetail{
				ID:   2,
				Name: "go",
			},
		}, nil
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
		total         int64
	}{
		{name: "global", wantGlobal: true, total: 42},
		{name: "community", communityID: 3, wantCommunity: true, total: 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetPostDeps(t)
			globalCalled := false
			communityCalled := false
			// 列表的主体数据这里统一 stub，测试重点放在“路由到哪条榜单逻辑”和“分页元数据是否正确”。
			stubPostsByIDs(t)
			countPostsFromMySQL = func() (int64, error) { return tt.total, nil }
			countPostsInCommunity = func(communityID int64) (int64, error) {
				if communityID != tt.communityID {
					t.Fatalf("communityID = %d, want %d", communityID, tt.communityID)
				}
				return tt.total, nil
			}

			getPostIDsInOrder = func(p *models.ParamPostList) ([]string, error) {
				globalCalled = true
				return []string{"11"}, nil
			}
			getCommunityPostIDsInOrder = func(p *models.ParamPostList) ([]string, error) {
				communityCalled = true
				return []string{"11"}, nil
			}

			got, err := GetPostListNew(&models.ParamPostList{CommunityID: tt.communityID, Page: 1, Size: 10, Order: "time"})
			if err != nil {
				t.Fatalf("GetPostListNew error: %v", err)
			}
			if globalCalled != tt.wantGlobal || communityCalled != tt.wantCommunity {
				t.Fatalf("globalCalled=%v communityCalled=%v, want %v/%v", globalCalled, communityCalled, tt.wantGlobal, tt.wantCommunity)
			}
			if got.Pagination.Total != tt.total {
				t.Fatalf("total = %d, want %d", got.Pagination.Total, tt.total)
			}
		})
	}
}

func TestGetPostListNewRoutesToSearch(t *testing.T) {
	resetPostDeps(t)
	searchCalled := false
	searchPostBundlesFromMySQL = func(p *models.ParamPostList) ([]*models.ApiPostDetail, error) {
		searchCalled = true
		if p.Keyword != "Redis" {
			t.Fatalf("keyword = %q, want Redis", p.Keyword)
		}
		return []*models.ApiPostDetail{{
			AuthorName: "alice",
			Post:       &models.Post{ID: 10, AuthorID: 1, CommunityID: 2},
			CommunityDetail: &models.CommunityDetail{
				ID:   2,
				Name: "go",
			},
		}}, nil
	}
	getPostVoteData = func(ids []string) ([]int64, error) {
		if !reflect.DeepEqual(ids, []string{"10"}) {
			t.Fatalf("vote ids = %#v", ids)
		}
		return []int64{3}, nil
	}
	countSearchPostsFromMySQL = func(p *models.ParamPostList) (int64, error) {
		return 1, nil
	}

	got, err := GetPostListNew(&models.ParamPostList{Page: 1, Size: 10, Order: "time", Keyword: " Redis "})
	if err != nil {
		t.Fatalf("GetPostListNew search error: %v", err)
	}
	if !searchCalled || got.Pagination.Total != 1 || got.Items[0].VoteNum != 3 {
		t.Fatalf("unexpected search result: %#v", got)
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
			// 这里验证两个点：
			// 1. 非法字符串 ID 会被跳过
			// 2. MySQL 返回的聚合详情与 Redis 票数会按顺序合并
			getPostBundlesByIDsFromMySQL = func(ids []int64) ([]*models.ApiPostDetail, error) {
				if !reflect.DeepEqual(ids, tt.wantIDs) {
					t.Fatalf("ids = %#v, want %#v", ids, tt.wantIDs)
				}
				posts := make([]*models.ApiPostDetail, 0, len(ids))
				for _, id := range ids {
					posts = append(posts, &models.ApiPostDetail{
						AuthorName: "alice",
						Post:       &models.Post{ID: id, AuthorID: 1, CommunityID: 2},
						CommunityDetail: &models.CommunityDetail{
							ID:   2,
							Name: "go",
						},
					})
				}
				return posts, nil
			}
			getPostVoteData = func(ids []string) ([]int64, error) {
				if !reflect.DeepEqual(ids, tt.ids) {
					t.Fatalf("vote ids = %#v, want %#v", ids, tt.ids)
				}
				return tt.wantVotes, nil
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
	getPostBundlesByIDsFromMySQL = func(ids []int64) ([]*models.ApiPostDetail, error) {
		return []*models.ApiPostDetail{{
			AuthorName: "alice",
			Post:       &models.Post{ID: ids[0], AuthorID: 1, CommunityID: 2},
			CommunityDetail: &models.CommunityDetail{
				ID:   2,
				Name: "go",
			},
		}}, nil
	}
	getPostVoteData = func(ids []string) ([]int64, error) {
		return []int64{1}, nil
	}
}
