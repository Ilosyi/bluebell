package models

type SignUpParam struct {
	Username   string `json:"username" binding:"required,min=2,max=20"`
	Password   string `json:"password" binding:"required,min=6,max=20"`
	RePassword string `json:"re_password" binding:"required,eqfield=Password"`
}
