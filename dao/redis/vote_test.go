package redis

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"bluebell/models"
	miniredis "github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func setupRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	s := miniredis.RunT(t)
	origRDB := rdb
	rdb = goredis.NewClient(&goredis.Options{Addr: s.Addr()})
	t.Cleanup(func() {
		_ = rdb.Close()
		rdb = origRDB
		s.Close()
	})
	return s
}

func TestGetRediskey(t *testing.T) {
	if got := getRediskey(KeyPostTimeZset); got != "bluebell:post:time" {
		t.Fatalf("getRediskey = %s", got)
	}
}

func TestSavePostTimeAndScore(t *testing.T) {
	setupRedis(t)
	now := time.Unix(1000, 0)

	if err := SavePostTimeAndScore(10, 2, now); err != nil {
		t.Fatalf("SavePostTimeAndScore error: %v", err)
	}
	if got := GetPostScore(10); got != float64(now.Unix()) {
		t.Fatalf("GetPostScore = %f, want %d", got, now.Unix())
	}
	members, err := rdb.SMembers(context.Background(), getRediskey(KeyCommunitySetPF+"2")).Result()
	if err != nil {
		t.Fatalf("SMembers error: %v", err)
	}
	if !reflect.DeepEqual(members, []string{"10"}) {
		t.Fatalf("community members = %#v", members)
	}
}

func TestGetPostIDsInOrder(t *testing.T) {
	setupRedis(t)
	ctx := context.Background()
	_, err := rdb.ZAdd(ctx, getRediskey(KeyPostTimeZset),
		goredis.Z{Score: 10, Member: "10"},
		goredis.Z{Score: 30, Member: "30"},
		goredis.Z{Score: 20, Member: "20"},
	).Result()
	if err != nil {
		t.Fatalf("ZAdd time error: %v", err)
	}
	_, err = rdb.ZAdd(ctx, getRediskey(KeyPostScoreZset),
		goredis.Z{Score: 1, Member: "10"},
		goredis.Z{Score: 3, Member: "30"},
		goredis.Z{Score: 2, Member: "20"},
	).Result()
	if err != nil {
		t.Fatalf("ZAdd score error: %v", err)
	}

	tests := []struct {
		name string
		p    *models.ParamPostList
		want []string
	}{
		{
			name: "time order",
			p:    &models.ParamPostList{Page: 1, Size: 2, Order: "time"},
			want: []string{"30", "20"},
		},
		{
			name: "score order second page",
			p:    &models.ParamPostList{Page: 2, Size: 1, Order: "score"},
			want: []string{"20"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPostIDsInOrder(tt.p)
			if err != nil {
				t.Fatalf("GetPostIDsInOrder error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ids = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestGetCommunityPostIDsInOrder(t *testing.T) {
	setupRedis(t)
	ctx := context.Background()
	_, err := rdb.SAdd(ctx, getRediskey(KeyCommunitySetPF+"2"), "10", "30").Result()
	if err != nil {
		t.Fatalf("SAdd error: %v", err)
	}
	_, err = rdb.ZAdd(ctx, getRediskey(KeyPostScoreZset),
		goredis.Z{Score: 1, Member: "10"},
		goredis.Z{Score: 3, Member: "30"},
		goredis.Z{Score: 2, Member: "20"},
	).Result()
	if err != nil {
		t.Fatalf("ZAdd error: %v", err)
	}

	got, err := GetCommunityPostIDsInOrder(&models.ParamPostList{
		CommunityID: 2,
		Page:        1,
		Size:        10,
		Order:       "score",
	})
	if err != nil {
		t.Fatalf("GetCommunityPostIDsInOrder error: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"30", "10"}) {
		t.Fatalf("ids = %#v", got)
	}
}

func TestGetPostVoteData(t *testing.T) {
	setupRedis(t)
	ctx := context.Background()
	_, err := rdb.ZAdd(ctx, getRediskey(KeyPostVotedZsetPrefix+"10"),
		goredis.Z{Score: 1, Member: "1"},
		goredis.Z{Score: -1, Member: "2"},
		goredis.Z{Score: 1, Member: "3"},
	).Result()
	if err != nil {
		t.Fatalf("ZAdd error: %v", err)
	}
	_, err = rdb.ZAdd(ctx, getRediskey(KeyPostVotedZsetPrefix+"11"),
		goredis.Z{Score: 1, Member: "1"},
	).Result()
	if err != nil {
		t.Fatalf("ZAdd error: %v", err)
	}

	got, err := GetPostVoteData([]string{"10", "11"})
	if err != nil {
		t.Fatalf("GetPostVoteData error: %v", err)
	}
	if !reflect.DeepEqual(got, []int64{2, 1}) {
		t.Fatalf("vote data = %#v", got)
	}
}

func TestVoteForPost(t *testing.T) {
	t.Run("upvote repeat switch and cancel", func(t *testing.T) {
		setupRedis(t)
		now := time.Now()
		if err := SavePostTimeAndScore(10, 2, now); err != nil {
			t.Fatalf("SavePostTimeAndScore error: %v", err)
		}

		if err := VoteForPost(1, "10", 1); err != nil {
			t.Fatalf("upvote error: %v", err)
		}
		if got := GetPostScore(10); got != float64(now.Unix())+scorePerVote {
			t.Fatalf("score after upvote = %f", got)
		}
		if got := GetUserVoteScore(1, 10); got != 1 {
			t.Fatalf("user vote = %f, want 1", got)
		}

		if err := VoteForPost(1, "10", 1); !errors.Is(err, ErrVoteRepeated) {
			t.Fatalf("repeat error = %v", err)
		}

		if err := VoteForPost(1, "10", -1); err != nil {
			t.Fatalf("switch to downvote error: %v", err)
		}
		if got := GetUserVoteScore(1, 10); got != -1 {
			t.Fatalf("user vote = %f, want -1", got)
		}

		if err := VoteForPost(1, "10", 0); err != nil {
			t.Fatalf("cancel error: %v", err)
		}
		if got := GetUserVoteScore(1, 10); got != 0 {
			t.Fatalf("user vote = %f, want 0 after cancel", got)
		}
	})

	t.Run("cancel without vote", func(t *testing.T) {
		setupRedis(t)
		if err := SavePostTimeAndScore(10, 2, time.Now()); err != nil {
			t.Fatalf("SavePostTimeAndScore error: %v", err)
		}
		if err := VoteForPost(1, "10", 0); !errors.Is(err, ErrVoteRepeated) {
			t.Fatalf("cancel without vote error = %v", err)
		}
	})

	t.Run("expired post", func(t *testing.T) {
		setupRedis(t)
		expired := time.Now().Add(-(oneWeekInSeconds + 1) * time.Second)
		if err := SavePostTimeAndScore(10, 2, expired); err != nil {
			t.Fatalf("SavePostTimeAndScore error: %v", err)
		}
		if err := VoteForPost(1, "10", 1); !errors.Is(err, ErrVoteTimeExpire) {
			t.Fatalf("expired error = %v", err)
		}
	})
}
