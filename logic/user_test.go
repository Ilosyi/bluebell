package logic

import (
	"errors"
	"testing"

	"bluebell/models"
	"bluebell/pkg/jwt"
)

func resetUserDeps(t *testing.T) {
	t.Helper()
	origCheckUserExist := checkUserExist
	origInsertUser := insertUser
	origLoginUser := loginUser
	origGetUserProfile := getUserProfile
	origUpdateProfile := updateProfile
	origGenUserID := genUserID
	origGenToken := genToken
	t.Cleanup(func() {
		checkUserExist = origCheckUserExist
		insertUser = origInsertUser
		loginUser = origLoginUser
		getUserProfile = origGetUserProfile
		updateProfile = origUpdateProfile
		genUserID = origGenUserID
		genToken = origGenToken
	})
}

func TestSignUp(t *testing.T) {
	tests := []struct {
		name           string
		checkErr       error
		insertErr      error
		wantErr        error
		wantInsertCall bool
	}{
		{
			name:           "success",
			wantInsertCall: true,
		},
		{
			name:     "user exists",
			checkErr: errors.New("用户已存在"),
			wantErr:  errors.New("用户已存在"),
		},
		{
			name:           "insert fails",
			insertErr:      errors.New("insert failed"),
			wantErr:        errors.New("insert failed"),
			wantInsertCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetUserDeps(t)
			const wantUserID int64 = 12345
			insertCalled := false

			checkUserExist = func(username string) error {
				if username != "alice" {
					t.Fatalf("username = %q, want alice", username)
				}
				return tt.checkErr
			}
			genUserID = func() int64 { return wantUserID }
			insertUser = func(user *models.User) error {
				insertCalled = true
				if user.UserID != wantUserID {
					t.Fatalf("user id = %d, want %d", user.UserID, wantUserID)
				}
				if user.Username != "alice" || user.Password != "password123" {
					t.Fatalf("unexpected user: %#v", user)
				}
				return tt.insertErr
			}

			err := SignUp(&models.SignUpParam{
				Username:   "alice",
				Password:   "password123",
				RePassword: "password123",
			})
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("SignUp error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("SignUp error = %v, want %v", err, tt.wantErr)
			}
			if insertCalled != tt.wantInsertCall {
				t.Fatalf("insert called = %v, want %v", insertCalled, tt.wantInsertCall)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name      string
		loginErr  error
		tokenErr  error
		wantErr   error
		wantToken string
	}{
		{
			name:      "success",
			wantToken: "jwt-token",
		},
		{
			name:     "mysql login fails",
			loginErr: errors.New("用户名或密码错误"),
			wantErr:  errors.New("用户名或密码错误"),
		},
		{
			name:     "token generation fails",
			tokenErr: errors.New("token failed"),
			wantErr:  errors.New("token failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetUserDeps(t)

			loginUser = func(user *models.User) error {
				if user.Username != "bob" || user.Password != "password123" {
					t.Fatalf("unexpected login user: %#v", user)
				}
				if tt.loginErr != nil {
					return tt.loginErr
				}
				user.UserID = 99
				return nil
			}
			genToken = func(claims *jwt.Myclaims) (string, error) {
				if claims.UserID != 99 || claims.Username != "bob" {
					t.Fatalf("unexpected claims: %#v", claims)
				}
				return tt.wantToken, tt.tokenErr
			}

			user, err := Login(&models.LoginParam{Username: "bob", Password: "password123"})
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("Login error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if err.Error() != tt.wantErr.Error() {
					t.Fatalf("Login error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if user.UserID != 99 || user.Username != "bob" || user.Token != tt.wantToken {
				t.Fatalf("unexpected user: %#v", user)
			}
		})
	}
}
