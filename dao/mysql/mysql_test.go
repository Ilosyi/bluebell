package mysql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"bluebell/models"
	"github.com/jmoiron/sqlx"
)

type sqlExpectation struct {
	queryContains string
	args          []driver.Value
	columns       []string
	rows          [][]driver.Value
	err           error
	exec          bool
}

var sqlMock = struct {
	sync.Mutex
	expectations []sqlExpectation
}{}

type testConnector struct{}

func (testConnector) Connect(context.Context) (driver.Conn, error) { return testConn{}, nil }
func (testConnector) Driver() driver.Driver                        { return testDriver{} }

type testDriver struct{}

func (testDriver) Open(string) (driver.Conn, error) { return testConn{}, nil }

type testConn struct{}

func (testConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not implemented")
}
func (testConn) Close() error              { return nil }
func (testConn) Begin() (driver.Tx, error) { return nil, errors.New("tx not implemented") }

func (testConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	exp, err := popExpectation(query, args, true)
	if err != nil {
		return nil, err
	}
	if exp.err != nil {
		return nil, exp.err
	}
	return driver.RowsAffected(1), nil
}

func (testConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	exp, err := popExpectation(query, args, false)
	if err != nil {
		return nil, err
	}
	if exp.err != nil {
		return nil, exp.err
	}
	return &testRows{columns: exp.columns, rows: exp.rows}, nil
}

type testRows struct {
	columns []string
	rows    [][]driver.Value
	idx     int
}

func (r *testRows) Columns() []string { return r.columns }
func (r *testRows) Close() error      { return nil }
func (r *testRows) Next(dest []driver.Value) error {
	if r.idx >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.idx])
	r.idx++
	return nil
}

func setMockDB(t *testing.T, expectations ...sqlExpectation) {
	t.Helper()
	origDB := db
	sqlMock.Lock()
	sqlMock.expectations = append([]sqlExpectation(nil), expectations...)
	sqlMock.Unlock()

	db = sqlx.NewDb(sql.OpenDB(testConnector{}), "mysql")
	t.Cleanup(func() {
		_ = db.Close()
		db = origDB
		sqlMock.Lock()
		defer sqlMock.Unlock()
		if len(sqlMock.expectations) != 0 {
			t.Fatalf("unmet SQL expectations: %#v", sqlMock.expectations)
		}
	})
}

func popExpectation(query string, args []driver.NamedValue, exec bool) (sqlExpectation, error) {
	sqlMock.Lock()
	defer sqlMock.Unlock()
	if len(sqlMock.expectations) == 0 {
		return sqlExpectation{}, errors.New("unexpected SQL operation")
	}
	exp := sqlMock.expectations[0]
	sqlMock.expectations = sqlMock.expectations[1:]
	if exp.exec != exec {
		return exp, errors.New("unexpected SQL operation type")
	}
	if !strings.Contains(normalizeSQL(query), normalizeSQL(exp.queryContains)) {
		return exp, errors.New("unexpected SQL query: " + query)
	}
	gotArgs := make([]driver.Value, 0, len(args))
	for _, arg := range args {
		gotArgs = append(gotArgs, arg.Value)
	}
	wantArgs := exp.args
	if wantArgs == nil {
		wantArgs = []driver.Value{}
	}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		return exp, errors.New("unexpected SQL args")
	}
	return exp, nil
}

func normalizeSQL(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func TestEncryptPassword(t *testing.T) {
	got, err := encryptPassword("password123")
	if err != nil {
		t.Fatalf("encryptPassword error: %v", err)
	}
	if got != "70617373776f7264313233e373e2e4d68ff16ca66294b74a9be4e5" {
		t.Fatalf("encryptPassword = %s", got)
	}
}

func TestUserDAO(t *testing.T) {
	t.Run("CheckUserExist success", func(t *testing.T) {
		setMockDB(t, sqlExpectation{
			queryContains: "select count(user_id) from user where username=?",
			args:          []driver.Value{"alice"},
			columns:       []string{"count"},
			rows:          [][]driver.Value{{int64(0)}},
		})
		if err := CheckUserExist("alice"); err != nil {
			t.Fatalf("CheckUserExist error: %v", err)
		}
	})

	t.Run("CheckUserExist exists", func(t *testing.T) {
		setMockDB(t, sqlExpectation{
			queryContains: "select count(user_id) from user where username=?",
			args:          []driver.Value{"alice"},
			columns:       []string{"count"},
			rows:          [][]driver.Value{{int64(1)}},
		})
		if err := CheckUserExist("alice"); err == nil || err.Error() != "用户已存在" {
			t.Fatalf("CheckUserExist error = %v", err)
		}
	})

	t.Run("InsertUser hashes password", func(t *testing.T) {
		hashed, _ := encryptPassword("password123")
		setMockDB(t, sqlExpectation{
			queryContains: "insert into user (user_id,username,password,nickname) values (?,?,?,?)",
			args:          []driver.Value{int64(1), "alice", hashed, "alice"},
			exec:          true,
		})
		user := &models.User{UserID: 1, Username: "alice", Password: "password123"}
		if err := InsertUser(user); err != nil {
			t.Fatalf("InsertUser error: %v", err)
		}
		if user.Password != hashed {
			t.Fatalf("password = %s, want hash %s", user.Password, hashed)
		}
	})

	t.Run("Login success", func(t *testing.T) {
		hashed, _ := encryptPassword("password123")
		setMockDB(t, sqlExpectation{
			queryContains: "select user_id, username, password, coalesce(nickname, '') as nickname, coalesce(avatar_url, '') as avatar_url, coalesce(bio, '') as bio from user where username=? or nickname=? limit 1",
			args:          []driver.Value{"alice", "alice"},
			columns:       []string{"user_id", "username", "password", "nickname", "avatar_url", "bio"},
			rows:          [][]driver.Value{{int64(1), "alice", hashed, "Alice", "", "hello"}},
		})
		user := &models.User{Username: "alice", Password: "password123"}
		if err := Login(user); err != nil {
			t.Fatalf("Login error: %v", err)
		}
		if user.UserID != 1 || user.Username != "alice" {
			t.Fatalf("unexpected user: %#v", user)
		}
	})

	t.Run("Login wrong password", func(t *testing.T) {
		setMockDB(t, sqlExpectation{
			queryContains: "select user_id, username, password, coalesce(nickname, '') as nickname, coalesce(avatar_url, '') as avatar_url, coalesce(bio, '') as bio from user where username=? or nickname=? limit 1",
			args:          []driver.Value{"alice", "alice"},
			columns:       []string{"user_id", "username", "password", "nickname", "avatar_url", "bio"},
			rows:          [][]driver.Value{{int64(1), "alice", "bad-hash", "", "", ""}},
		})
		err := Login(&models.User{Username: "alice", Password: "password123"})
		if err == nil || err.Error() != "账号或密码错误" {
			t.Fatalf("Login error = %v", err)
		}
	})

	t.Run("GetUserById", func(t *testing.T) {
		setMockDB(t, sqlExpectation{
			queryContains: "select user_id, username, coalesce(nickname, '') as nickname, coalesce(avatar_url, '') as avatar_url, coalesce(bio, '') as bio, create_time from user where user_id=?",
			args:          []driver.Value{int64(1)},
			columns:       []string{"user_id", "username", "nickname", "avatar_url", "bio", "create_time"},
			rows:          [][]driver.Value{{int64(1), "alice", "Alice", "", "hello", time.Unix(100, 0)}},
		})
		user, err := GetUserById(1)
		if err != nil {
			t.Fatalf("GetUserById error: %v", err)
		}
		if user.Username != "alice" {
			t.Fatalf("username = %s", user.Username)
		}
	})
}

func TestCommunityDAO(t *testing.T) {
	now := time.Unix(100, 0)
	t.Run("GetCommunityList", func(t *testing.T) {
		setMockDB(t, sqlExpectation{
			queryContains: "SELECT community_id, community_name FROM community",
			columns:       []string{"community_id", "community_name"},
			rows:          [][]driver.Value{{int64(1), "go"}, {int64(2), "vue"}},
		})
		got, err := GetCommunityList()
		if err != nil {
			t.Fatalf("GetCommunityList error: %v", err)
		}
		if len(got) != 2 || got[0].Name != "go" {
			t.Fatalf("unexpected communities: %#v", got)
		}
	})

	t.Run("GetCommunityDetailByID success", func(t *testing.T) {
		setMockDB(t, sqlExpectation{
			queryContains: "SELECT community_id, community_name, introduction, create_time FROM community WHERE community_id=?",
			args:          []driver.Value{int64(1)},
			columns:       []string{"community_id", "community_name", "introduction", "create_time"},
			rows:          [][]driver.Value{{int64(1), "go", "golang", now}},
		})
		got, err := GetCommunityDetailByID(1)
		if err != nil {
			t.Fatalf("GetCommunityDetailByID error: %v", err)
		}
		if got.ID != 1 || got.Introduction != "golang" {
			t.Fatalf("unexpected detail: %#v", got)
		}
	})

	t.Run("GetCommunityDetailByID no rows", func(t *testing.T) {
		setMockDB(t, sqlExpectation{
			queryContains: "SELECT community_id, community_name, introduction, create_time FROM community WHERE community_id=?",
			args:          []driver.Value{int64(404)},
			columns:       []string{"community_id", "community_name", "introduction", "create_time"},
		})
		_, err := GetCommunityDetailByID(404)
		if err == nil || err.Error() != "无效的ID" {
			t.Fatalf("GetCommunityDetailByID error = %v", err)
		}
	})
}

func TestPostDAO(t *testing.T) {
	now := time.Unix(100, 0)
	t.Run("CreatePost", func(t *testing.T) {
		setMockDB(t, sqlExpectation{
			queryContains: "insert into post( post_id, title, content, author_id, community_id, status) values (?, ?, ?, ?, ?, ?)",
			args:          []driver.Value{int64(1), "title", "content", int64(2), int64(3), int64(models.PostStatusPublished)},
			exec:          true,
		})
		err := CreatePost(&models.Post{ID: 1, Title: "title", Content: "content", AuthorID: 2, CommunityID: 3, Status: models.PostStatusPublished})
		if err != nil {
			t.Fatalf("CreatePost error: %v", err)
		}
	})

	t.Run("GetPostById", func(t *testing.T) {
		setMockDB(t, sqlExpectation{
			queryContains: "SELECT post_id, title, content, author_id, community_id,create_time FROM post WHERE post_id=?",
			args:          []driver.Value{int64(1)},
			columns:       []string{"post_id", "title", "content", "author_id", "community_id", "create_time"},
			rows:          [][]driver.Value{{int64(1), "title", "content", int64(2), int64(3), now}},
		})
		got, err := GetPostById(1)
		if err != nil {
			t.Fatalf("GetPostById error: %v", err)
		}
		if got.ID != 1 || got.Title != "title" {
			t.Fatalf("unexpected post: %#v", got)
		}
	})

	t.Run("GetPostListByIDs empty", func(t *testing.T) {
		setMockDB(t)
		got, err := GetPostListByIDs(nil)
		if err != nil {
			t.Fatalf("GetPostListByIDs error: %v", err)
		}
		if got != nil {
			t.Fatalf("GetPostListByIDs = %#v, want nil", got)
		}
	})

	t.Run("GetPostListByIDs success", func(t *testing.T) {
		setMockDB(t, sqlExpectation{
			queryContains: "select post_id, title, content, author_id, community_id, create_time from post where post_id in (?, ?) order by FIND_IN_SET(post_id, ?)",
			args:          []driver.Value{int64(10), int64(11), "10,11"},
			columns:       []string{"post_id", "title", "content", "author_id", "community_id", "create_time"},
			rows: [][]driver.Value{
				{int64(10), "title10", "content10", int64(2), int64(3), now},
				{int64(11), "title11", "content11", int64(2), int64(3), now},
			},
		})
		got, err := GetPostListByIDs([]int64{10, 11})
		if err != nil {
			t.Fatalf("GetPostListByIDs error: %v", err)
		}
		if len(got) != 2 || got[0].ID != 10 || got[1].ID != 11 {
			t.Fatalf("unexpected posts: %#v", got)
		}
	})

	t.Run("GetPostList", func(t *testing.T) {
		setMockDB(t, sqlExpectation{
			queryContains: "SELECT post_id, title, content, author_id, community_id,create_time FROM post WHERE status=? ORDER BY create_time DESC LIMIT ?,?",
			args:          []driver.Value{int64(models.PostStatusPublished), int64(10), int64(10)},
			columns:       []string{"post_id", "title", "content", "author_id", "community_id", "create_time"},
			rows:          [][]driver.Value{{int64(10), "title", "content", int64(2), int64(3), now}},
		})
		got, err := GetPostList(2, 10)
		if err != nil {
			t.Fatalf("GetPostList error: %v", err)
		}
		if len(got) != 1 || got[0].ID != 10 {
			t.Fatalf("unexpected posts: %#v", got)
		}
	})
}
