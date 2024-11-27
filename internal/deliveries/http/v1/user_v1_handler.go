package v1

import (
	"github.com/labstack/echo/v4"
	"go-community/internal/config"
	"go-community/internal/deliveries/http/common/response"
	"go-community/internal/models"
	"go-community/internal/pkg/authorization"
	"go-community/internal/pkg/validator"
	"go-community/internal/usecases"
	"net/http"
	"strings"
	"time"
)

type UserHandler struct {
	usecase *usecases.Usecases
	auth    *authorization.Auth
	conf    *config.Configuration
}

func NewUserHandler(api *echo.Group, u *usecases.Usecases, c *config.Configuration) {
	handler := &UserHandler{usecase: u, conf: c}

	endpoint := api.Group("/users")
	endpoint.POST("/volunteers", handler.CreateVolunteer)
	endpoint.POST("/login", handler.Login)
	endpoint.GET("/check/:identifier", handler.Check)
	endpoint.GET("/:communityId", handler.GetByCommunityId)
	endpoint.PATCH("/:identifier/password", handler.UpdatePassword)
	endpoint.PUT("/logout", handler.Logout)

	userTypeEndpoint := endpoint.Group("/types")
	userTypeEndpoint.POST("", handler.CreateUserType)
	userTypeEndpoint.GET("", handler.GetAllUserTypes)
}

// CreateVolunteer godoc
// @Summary Create Volunteer User
// @Description Create user for volunteer
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.CreateVolunteerRequest true "User object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 201 {object} models.CreateVolunteerResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v1/users/volunteer [post]
func (uh *UserHandler) CreateVolunteer(ctx echo.Context) error {
	var request models.CreateVolunteerRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, err)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	user, err := uh.usecase.User.CreateVolunteer(ctx.Request().Context(), &request)
	if err != nil {
		return response.Error(ctx, err)
	}

	res := models.CreateVolunteerResponse{Type: models.TYPE_USER, CommunityId: user.CommunityID, Name: user.Name, PhoneNumber: user.PhoneNumber, Email: user.Email, CampusCode: user.CampusCode, PlaceOfBirth: user.PlaceOfBirth, DateOfBirth: user.DateOfBirth, Address: user.Address, Gender: user.Gender, DepartmentCode: user.Department, CoolID: user.CoolID, KKJNumber: user.KKJNumber, JemaatId: user.JemaatID, IsKOM100: user.IsKom100, IsBaptis: user.IsBaptized, MaritalStatus: user.MaritalStatus, Status: user.Status, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}
	return response.Success(ctx, http.StatusCreated, res.ToCreateVolunteer())
}

// Login godoc
// @Summary Login User
// @Description Login for all type of users
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.LoginUserRequest true "User object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 200 {object} models.LoginUserResponse{tokens=models.TokensResponse} "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v1/users/login [post]
func (uh *UserHandler) Login(ctx echo.Context) error {
	var request models.LoginUserRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	user, tokens, err := uh.usecase.User.Login(ctx.Request().Context(), &request)
	if err != nil {
		return response.Error(ctx, err)
	}

	ctx.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Expires:  tokens.RefreshExpiry,
		HttpOnly: true,                    // Prevent client-side JavaScript access
		Secure:   true,                    // Only send over HTTPS
		SameSite: http.SameSiteStrictMode, // Prevent CSRF
	})

	res := models.LoginUserResponse{Type: models.TYPE_USER, CommunityId: user.CommunityID, Name: user.Name, PhoneNumber: user.PhoneNumber, Email: user.Email, CampusCode: user.CampusCode, PlaceOfBirth: user.PlaceOfBirth, DateOfBirth: user.DateOfBirth, Address: user.Address, Gender: user.Gender, DepartmentCode: user.Department, CoolID: user.CoolID, KKJNumber: user.KKJNumber, JemaatId: user.JemaatID, IsKOM100: user.IsKom100, IsBaptized: user.IsBaptized, MaritalStatus: user.MaritalStatus, Status: user.Status, Token: tokens.ToGenerateTokens(), UserType: user.UserType, Roles: user.Roles}
	return response.Success(ctx, http.StatusCreated, res.ToLogin())
}

// Check godoc
// @Summary Check User Exist
// @Description To check whether user is exist or not by email or phone number
// @Tags users
// @Accept json
// @Produce json
// @Param identifier path string true "object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 200 {object} models.CheckUserExistResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v1/users/check/{identifier} [get]
func (uh *UserHandler) Check(ctx echo.Context) error {

	isExist, err := uh.usecase.User.Check(ctx.Request().Context(), strings.ToLower(ctx.Param("identifier")))
	if err != nil {
		return response.Error(ctx, err)
	}

	res := models.CheckUserExistResponse{Type: models.TYPE_USER, User: isExist, Identifier: strings.ToLower(ctx.Param("identifier"))}
	return response.Success(ctx, http.StatusOK, res.ToCheck())
}

// CreateUserType godoc
// @Summary Create User Type
// @Description User Type is something like volunteer, lead or etc.
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.CreateUserTypeRequest true "User object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 201 {object} models.UserTypeResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v1/users/types [post]
func (uh *UserHandler) CreateUserType(ctx echo.Context) error {
	var request models.CreateUserTypeRequest
	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	// Validate inputs based on requirement
	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	// Usage of the usecase
	new, err := uh.usecase.UserType.Create(ctx.Request().Context(), &request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusCreated, new.ToResponse())
}

// GetAllUserTypes godoc
// @Summary Get All User Types
// @Description User Type is something like volunteer, lead or etc.
// @Tags users
// @Accept json
// @Produce json
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 200 {object} models.UserTypeResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v1/users/types [get]
func (uh *UserHandler) GetAllUserTypes(ctx echo.Context) error {
	data, err := uh.usecase.UserType.GetAll(ctx.Request().Context())
	if err != nil {
		return response.Error(ctx, err)
	}

	var res []models.UserTypeResponse
	for _, v := range data {
		res = append(res, *v.ToResponse())
	}

	return response.SuccessList(ctx, http.StatusOK, len(res), res)
}

// UpdatePassword godoc
// @Summary Update User Password
// @Description Update user Password
// @Tags users
// @Accept json
// @Produce json
// @Param identifier path string true "object that needs to be added"
// @Param user body models.UpdateUserPasswordRequest true "User object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 201 {object} models.UpdateUserPasswordResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v1/users/{identifier}/password [patch]
func (uh *UserHandler) UpdatePassword(ctx echo.Context) error {
	var request models.UpdateUserPasswordRequest
	parameter := models.UpdateUserPasswordParam{
		Identifier: strings.ToLower(ctx.Param("identifier")),
	}

	if err := ctx.Bind(&request); err != nil {
		return response.Error(ctx, models.ErrorInvalidInput)
	}

	if err := validator.Validate(parameter); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	if err := validator.Validate(request); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	user, err := uh.usecase.User.UpdatePassword(ctx.Request().Context(), &parameter, &request)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusOK, user.ToUpdatePassword())
}

// GetByCommunityId godoc
// @Summary Get User By Community ID
// @Description Get all information needed about user by community Id
// @Tags users
// @Accept json
// @Produce json
// @Param communityId path int true "object that needs to be added"
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 200 {object} models.GetOneByCommunityIdResponse "Response indicates that the request succeeded and the resources has been fetched and transmitted in the message body"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v1/users/{communityId} [get]
func (uh *UserHandler) GetByCommunityId(ctx echo.Context) error {
	parameter := models.GetOneByCommunityIdParameter{
		CommunityId: strings.ToLower(ctx.Param("communityId")),
	}

	if err := validator.Validate(parameter); err != nil {
		return response.ErrorValidation(ctx, err)
	}

	data, err := uh.usecase.User.GetByCommunityId(ctx.Request().Context(), parameter)
	if err != nil {
		return response.Error(ctx, err)
	}

	return response.Success(ctx, http.StatusOK, data.ToGetOneByCommunityId())

}

// Logout godoc
// @Summary Logout User
// @Description Logout user for all kinds of user
// @Tags users
// @Accept json
// @Produce json
// @Param X-API-Key header string true "mandatory header to access endpoint"
// @Success 204 "No Content"
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 422 {object} models.ErrorValidationResponse{errors=validator.ErrorValidateResponse} "Validation error. This can happen if there is an error validation while create account"
// @Router /v1/users/logout [put]
func (uh *UserHandler) Logout(ctx echo.Context) error {
	ctx.SetCookie(&http.Cookie{
		Name:     "refresh_token",                // Name of the cookie holding the refresh token
		Value:    "",                             // Set value to empty
		Expires:  time.Now().Add(-1 * time.Hour), // Expire the cookie immediately
		HttpOnly: true,                           // Prevent client-side access to the cookie
		Secure:   true,                           // Set Secure flag if using HTTPS
		SameSite: http.SameSiteStrictMode,        // Set SameSite attribute
	})

	return ctx.NoContent(http.StatusNoContent)
}

//
//func (uh *UserHandler) CreateUser(ctx echo.Context) error {
//	var request models.CreateUserRequest
//	if err := ctx.Bind(&request); err != nil {
//		return response.Error(ctx, models.ErrorInvalidInput)
//	}
//
//	if err := validator.Validate(request); err != nil {
//		return response.ErrorValidation(ctx, err)
//	}
//
//	new, err := uh.usecase.User.CreateUser(ctx.Request().Context(), &request)
//	if err != nil {
//		return response.Error(ctx, err)
//	}
//
//	return response.Success(ctx, http.StatusCreated, new.ToCreateUser())
//}
//
//func (uh *UserHandler) Check(ctx echo.Context) error {
//	request := ctx.QueryParam("email")
//
//	isExist, data, err := uh.usecase.User.CheckByEmail(ctx.Request().Context(), request)
//	if err != nil {
//		return response.Error(ctx, err)
//	}
//
//	res := models.CheckUserEmailResponse{IsExist: isExist, UserType: data.UserType, Email: request}
//	return response.Success(ctx, http.StatusOK, res.ToCheck())
//}
//
