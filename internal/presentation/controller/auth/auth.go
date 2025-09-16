package authcontroller

import (
	"astral/internal/domain/dto"
	controllererrors "astral/internal/presentation/controller/errors"

	"github.com/gin-gonic/gin"
)

// @Summary Register
// @Description Registration user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body registerRequest true "user data"
// @Success 200 {object} registerResponse
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Not Found"
// @Failure 409 {object} response.ErrorResponse "Conflict"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /api/register [post]
func (c *Controller) Register(ctx *gin.Context) {
	var regRequestData registerRequest
	err := ctx.ShouldBindJSON(&regRequestData)
	if err != nil {
		c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("invalid json"))
		return
	}

	if regRequestData.Login == "" {
		c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("login field is required"))
		return
	}

	if regRequestData.Pass == "" {
		c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("pswd field is required"))
		return
	}

	if regRequestData.Token == "" {
		c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("token field is required"))
		return
	}

	user := dto.UserData{
		Login: regRequestData.Login,
		Password: regRequestData.Pass,
	}

	res, err := c.authService.Registration(regRequestData.Token, user)
	if err != nil {
		c.responseBuilder.Error(ctx, err)
		return
	}

	login := map[string]string{
		"login": res.Login,
	}

	c.responseBuilder.Ok(ctx, login, nil)
}

// @Tags auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "user data"
// @Success 200 {object} loginResponse
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Not Found"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Router /api/auth [post]
func (c *Controller) Login(ctx *gin.Context) {
	var loginRequestData loginRequest
	err := ctx.ShouldBindJSON(&loginRequestData)
	if err != nil {
		c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("invalid json"))
		return
	}

	if loginRequestData.Login == "" {
		c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("login field is required"))
		return
	}

	if loginRequestData.Pass == "" {
		c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("pswd field is required"))
		return
	}

	user := dto.UserData{
		Login: loginRequestData.Login,
		Password: loginRequestData.Pass,
	}

	res, err := c.authService.Login(user)
	if err != nil {
		c.responseBuilder.Error(ctx, err)
		return
	}

	token := map[string]string{
		"token": res.Token,
	}

	c.responseBuilder.Ok(ctx, token, nil)
}

// @Tags auth
// @Accept json
// @Produce json
// @Param request body deleteResponse true "user data"
// @Success 200 {object} loginResponse
// @Failure 400 {object} response.ErrorResponse "Bad Request"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Failure 403 {object} response.ErrorResponse "Forbidden"
// @Failure 404 {object} response.ErrorResponse "Not Found"
// @Failure 500 {object} response.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /api/auth/:token_id [delete]
func (c *Controller) DeleteSession(ctx *gin.Context) {
	token := c.utils.GetTokenFromHeader(ctx)
	if token == nil {
		return
	}

	tokenStr := ctx.Param("token_id")
	if tokenStr == "" {
		c.responseBuilder.Error(ctx, controllererrors.NewErrInvalidInputData("token param is required"))
		return
	}

	tokenData := dto.TokenData{
		Login: token.Login,
		Token: tokenStr,
	}

	res, err := c.authService.CloseSession(tokenData)
	if err != nil {
		c.responseBuilder.Error(ctx, err)
		return
	}

	isClosedSession := map[string]bool{
		res.Token: true,
	}
	
	c.responseBuilder.Ok(ctx, isClosedSession, nil)
}