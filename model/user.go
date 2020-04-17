package model

type LoginUser struct {
	Login    int    `form:"login" json:"login" example:"1" format:"int64" binding:"required"`
	Password string `form:"password" json:"password" example:"1234Aa" binding:"required"`
}

type User struct {
	Login int `json:"login" example:"1" format:"int64"`
}

type UserLoginResponse struct {
	Code   int
	Token  string
	Expire string
}
