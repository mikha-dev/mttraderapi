package controller

import (
	"mttraderapi/model"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
)

// Auth godoc
// @Summary Auth user
// @Description get user info
// @Tags users
// @Accept  json
// @Produce  json
// @Param user body model.LoginUser true "credentials"
// @Success 200 {object} model.UserLoginResponse
// @Failure 400 {object} httputil.HTTPError
// @Failure 401 {object} httputil.HTTPError
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /auth/login [post]
func (c *Controller) UserAuth(ctx *gin.Context) (interface{}, error) {
	dealer := getDealer(ctx)

	var loginVals model.LoginUser
	if err := ctx.ShouldBind(&loginVals); err != nil {
		return "", jwt.ErrMissingLoginValues
	}

	if ret, err := dealer.CheckPassword(loginVals.Login, loginVals.Password); ret {
		return &model.User{
			Login: loginVals.Login,
		}, nil
	} else {
		return nil, err
	}

	//return nil, jwt.ErrFailedAuthentication
}
