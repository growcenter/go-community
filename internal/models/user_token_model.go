package models

import "time"

var (
	TYPE_REFRESH_TOKEN = "refreshToken"
	TYPE_ACCESS_TOKEN  = "accessToken"
)

type UserToken struct {
	AccessToken   string    `json:"accessToken"`
	AccessExpiry  time.Time `json:"accessTokenExpiry"`
	RefreshToken  string    `json:"refreshToken"`
	RefreshExpiry time.Time `json:"refreshTokenExpiry"`
}

func (ut UserToken) ToGenerateTokens() []interface{} {
	access := TokensResponse{
		Type:      TYPE_ACCESS_TOKEN,
		Token:     ut.AccessToken,
		ExpiresAt: ut.AccessExpiry,
	}

	refresh := TokensResponse{
		Type:      TYPE_REFRESH_TOKEN,
		Token:     ut.RefreshToken,
		ExpiresAt: ut.RefreshExpiry,
	}

	return []interface{}{access, refresh}
}

func (at *GenerateAccessTokenResponse) ToCreateAccessToken() GenerateAccessTokenResponse {
	return GenerateAccessTokenResponse{
		Type:        TYPE_ACCESS_TOKEN,
		AccessToken: at.AccessToken,
	}
}

type GenerateAccessTokenResponse struct {
	Type        string `json:"type"`
	AccessToken string `json:"accessToken"`
}

type (
	TokensResponse struct {
		Type      string    `json:"type" example:"accessToken"`
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expiresAt"`
	}
	UserAccessTokenResponse struct {
		Type        string    `json:"type" example:"accessToken"`
		AccessToken string    `json:"accessToken"`
		ExpiresAt   time.Time `json:"expiresAt"`
	}
	UserRefreshTokenResponse struct {
		Type         string    `json:"type" example:"accessToken"`
		RefreshToken string    `json:"refreshToken"`
		ExpiresAt    time.Time `json:"expiresAt"`
	}
	UserTokens struct {
		AccessToken  UserAccessTokenResponse  `json:"accessToken"`
		RefreshToken UserRefreshTokenResponse `json:"refreshToken"`
	}
)

type UserTokenResponse struct {
	Type      string    `json:"type"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}
