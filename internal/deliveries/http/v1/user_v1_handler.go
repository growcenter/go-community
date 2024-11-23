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
	//endpoint.POST("")
	endpoint.GET("/check/:identifier", handler.Check)
	//endpoint.GET("/:accountNumber", handler.GetByAccountNumber)

	userTypeEndpoint := endpoint.Group("/types")
	userTypeEndpoint.POST("", handler.CreateUserType)
	userTypeEndpoint.GET("", handler.GetAllUserTypes)
}

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

	tokenRes := models.UserTokens{AccessToken: models.UserAccessTokenResponse{Type: models.TYPE_ACCESS_TOKEN, AccessToken: tokens.AccessToken, ExpiresAt: tokens.AccessExpiry}, RefreshToken: models.UserRefreshTokenResponse{Type: models.TYPE_REFRESH_TOKEN, RefreshToken: tokens.RefreshToken, ExpiresAt: tokens.RefreshExpiry}}
	res := models.LoginUserResponse{Type: models.TYPE_USER, CommunityId: user.CommunityID, Name: user.Name, PhoneNumber: user.PhoneNumber, Email: user.Email, CampusCode: user.CampusCode, PlaceOfBirth: user.PlaceOfBirth, DateOfBirth: user.DateOfBirth, Address: user.Address, Gender: user.Gender, DepartmentCode: user.Department, CoolID: user.CoolID, KKJNumber: user.KKJNumber, JemaatId: user.JemaatID, IsKOM100: user.IsKom100, IsBaptized: user.IsBaptized, MaritalStatus: user.MaritalStatus, Status: user.Status, Token: tokenRes, UserType: user.UserType, Roles: user.Roles}
	return response.Success(ctx, http.StatusCreated, res.ToLogin())
}

func (uh *UserHandler) Check(ctx echo.Context) error {

	isExist, err := uh.usecase.User.Check(ctx.Request().Context(), strings.ToLower(ctx.Param("identifier")))
	if err != nil {
		return response.Error(ctx, err)
	}

	res := models.CheckUserExistResponse{Type: models.TYPE_USER, User: isExist, Identifier: strings.ToLower(ctx.Param("identifier"))}
	return response.Success(ctx, http.StatusOK, res.ToCheck())
}

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
//func (uh *UserHandler) GetByAccountNumber(ctx echo.Context) error {
//	data, err := uh.usecase.User.GetByAccountNumber(ctx.Request().Context(), ctx.Param("accountNumber"))
//	if err != nil {
//		return response.Error(ctx, err)
//	}
//
//	return response.Success(ctx, http.StatusOK, data.ToGetUserByAccountNumber())
//
//}
