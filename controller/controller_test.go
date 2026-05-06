package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bluebell/models"
	"github.com/gin-gonic/gin"
)

type responseBody struct {
	Code int64           `json:"code"`
	Msg  any             `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func resetControllerDeps(t *testing.T) {
	t.Helper()
	origSignUp := signUp
	origLogin := login
	origCreatePost := createPost
	origGetPostByID := getPostByID
	origGetPostList := getPostList
	origGetPostListNew := getPostListNew
	origGetCommunityList := getCommunityList
	origGetCommunityDetail := getCommunityDetail
	origVoteForPost := voteForPost
	t.Cleanup(func() {
		signUp = origSignUp
		login = origLogin
		createPost = origCreatePost
		getPostByID = origGetPostByID
		getPostList = origGetPostList
		getPostListNew = origGetPostListNew
		getCommunityList = origGetCommunityList
		getCommunityDetail = origGetCommunityDetail
		voteForPost = origVoteForPost
	})
}

func TestResponseHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ResponseError(c, CodeNeedLogin)
	var got responseBody
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if got.Code != int64(CodeNeedLogin) || got.Msg != CodeNeedLogin.Msg() {
		t.Fatalf("unexpected error response: %#v", got)
	}

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	ResponseSuccess(c, gin.H{"hello": "world"})
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if got.Code != int64(CodeSuccess) {
		t.Fatalf("unexpected success code: %#v", got)
	}
}

func TestGetCurrentUserIDAndPageInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	if _, err := getCurrentUserID(c); !errors.Is(err, ErrorUserNotLogin) {
		t.Fatalf("expected ErrorUserNotLogin, got %v", err)
	}

	c.Set(CtxUserIDKey, int64(88))
	id, err := getCurrentUserID(c)
	if err != nil || id != 88 {
		t.Fatalf("getCurrentUserID = %d, %v", id, err)
	}

	req := httptest.NewRequest(http.MethodGet, "/?page=2&size=5", nil)
	c, _ = gin.CreateTestContext(w)
	c.Request = req
	page, size := getPageInfo(c)
	if page != 2 || size != 5 {
		t.Fatalf("getPageInfo = %d,%d", page, size)
	}
}

func TestRemoveTopStruct(t *testing.T) {
	got := removeTopStruct(map[string]string{
		"SignUpParam.username": "用户名错误",
	})
	if got["username"] != "用户名错误" {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestHandlerUserFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	if err := InitTrans("zh"); err != nil {
		t.Fatalf("InitTrans error: %v", err)
	}

	tests := []struct {
		name   string
		path   string
		method string
		body   string
		setup  func()
		want   int64
	}{
		{
			name:   "signup success",
			method: http.MethodPost,
			path:   "/signup",
			body:   `{"username":"alice","password":"password123","re_password":"password123"}`,
			setup: func() {
				signUp = func(p *models.SignUpParam) error { return nil }
			},
			want: int64(CodeSuccess),
		},
		{
			name:   "signup exists",
			method: http.MethodPost,
			path:   "/signup",
			body:   `{"username":"alice","password":"password123","re_password":"password123"}`,
			setup: func() {
				signUp = func(p *models.SignUpParam) error { return errors.New("用户已存在") }
			},
			want: int64(CodeUserExist),
		},
		{
			name:   "login success",
			method: http.MethodPost,
			path:   "/login",
			body:   `{"username":"alice","password":"password123"}`,
			setup: func() {
				login = func(p *models.LoginParam) (*models.User, error) {
					return &models.User{UserID: 7, Username: "alice", Token: "tok"}, nil
				}
			},
			want: int64(CodeSuccess),
		},
		{
			name:   "login bad password",
			method: http.MethodPost,
			path:   "/login",
			body:   `{"username":"alice","password":"password123"}`,
			setup: func() {
				login = func(p *models.LoginParam) (*models.User, error) {
					return nil, errors.New("用户名或密码错误")
				}
			},
			want: int64(CodeInvalidPassword),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetControllerDeps(t)
			tt.setup()

			r := gin.New()
			switch tt.path {
			case "/signup":
				r.POST("/signup", SignUpHandler)
			case "/login":
				r.POST("/login", LoginHandler)
			}

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			var got responseBody
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}
			if got.Code != tt.want {
				t.Fatalf("code = %d, want %d, body=%s", got.Code, tt.want, w.Body.String())
			}
		})
	}
}

func TestHandlerPostCommunityVote(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		setup  func()
		want   int64
	}{
		{
			name:   "create post success",
			method: http.MethodPost,
			path:   "/post",
			body:   `{"title":"hello","content":"world","community_id":1}`,
			setup: func() {
				createPost = func(p *models.Post) error {
					return nil
				}
			},
			want: int64(CodeSuccess),
		},
		{
			name:   "community detail invalid id",
			method: http.MethodGet,
			path:   "/community/abc",
			want:   int64(CodeInvalidParam),
		},
		{
			name:   "community list success",
			method: http.MethodGet,
			path:   "/community",
			setup: func() {
				getCommunityList = func() ([]*models.Community, error) {
					return []*models.Community{{ID: 1, Name: "go"}}, nil
				}
			},
			want: int64(CodeSuccess),
		},
		{
			name:   "post detail success",
			method: http.MethodGet,
			path:   "/post/1",
			setup: func() {
				getPostByID = func(id int64) (*models.ApiPostDetail, error) {
					return &models.ApiPostDetail{AuthorName: "alice", Post: &models.Post{ID: id}}, nil
				}
			},
			want: int64(CodeSuccess),
		},
		{
			name:   "post list success",
			method: http.MethodGet,
			path:   "/posts?page=1&size=10",
			setup: func() {
				getPostList = func(page, size int64) ([]*models.ApiPostDetail, error) {
					return []*models.ApiPostDetail{{AuthorName: "alice"}}, nil
				}
			},
			want: int64(CodeSuccess),
		},
		{
			name:   "post list new success",
			method: http.MethodGet,
			path:   "/posts2?page=1&size=10&order=time",
			setup: func() {
				getPostListNew = func(p *models.ParamPostList) ([]*models.ApiPostDetail, error) {
					return []*models.ApiPostDetail{{AuthorName: "alice"}}, nil
				}
			},
			want: int64(CodeSuccess),
		},
		{
			name:   "vote success",
			method: http.MethodPost,
			path:   "/vote",
			body:   `{"post_id":"1","direction":"1"}`,
			setup: func() {
				voteForPost = func(userID int64, p *models.ParamVoteData) error {
					if userID != 1 {
						t.Fatalf("userID = %d, want 1", userID)
					}
					return nil
				}
			},
			want: int64(CodeSuccess),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetControllerDeps(t)
			if tt.setup != nil {
				tt.setup()
			}

			r := gin.New()
			if tt.path == "/post" || tt.path == "/vote" {
				r.Use(func(c *gin.Context) {
					c.Set(CtxUserIDKey, int64(1))
					c.Next()
				})
			}

			switch {
			case tt.method == http.MethodPost && tt.path == "/post":
				r.POST("/post", CreatePostHandler)
			case strings.HasPrefix(tt.path, "/community/"):
				r.GET("/community/:id", CommunityDetailHandler)
			case tt.path == "/community":
				r.GET("/community", CommunityHandler)
			case tt.method == http.MethodGet && strings.HasPrefix(tt.path, "/post/"):
				r.GET("/post/:id", GetPostDetailHandler)
			case strings.HasPrefix(tt.path, "/posts?"):
				r.GET("/posts", GetPostListHandler)
			case strings.HasPrefix(tt.path, "/posts2?"):
				r.GET("/posts2", GetPostListHandler2)
			case tt.path == "/vote":
				r.POST("/vote", PostVoteHandler)
			}

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			var got responseBody
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}
			if got.Code != tt.want {
				t.Fatalf("code = %d, want %d, body=%s", got.Code, tt.want, w.Body.String())
			}
		})
	}
}
