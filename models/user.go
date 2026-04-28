package models

type User struct {
	UserID   int64  `db:"userid"`
	Username string `db:"username"`
	Password string `db:"password"`
}
