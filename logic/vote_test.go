package logic

import (
	"errors"
	"testing"

	"bluebell/models"
)

func resetVoteDeps(t *testing.T) {
	t.Helper()
	origVoteForPostInRedis := voteForPostInRedis
	t.Cleanup(func() {
		voteForPostInRedis = origVoteForPostInRedis
	})
}

func TestVoteForPost(t *testing.T) {
	tests := []struct {
		name     string
		redisErr error
		wantErr  bool
	}{
		{name: "success"},
		{name: "redis error", redisErr: errors.New("redis failed"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetVoteDeps(t)
			voteForPostInRedis = func(userID int64, postID string, value float64) error {
				if userID != 7 || postID != "100" || value != 1 {
					t.Fatalf("unexpected args: userID=%d postID=%s value=%f", userID, postID, value)
				}
				return tt.redisErr
			}

			err := VoteForPost(7, &models.ParamVoteData{PostId: "100", Direction: 1})
			if (err != nil) != tt.wantErr {
				t.Fatalf("VoteForPost error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
