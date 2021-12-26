package model

import (
	"github.com/gin-gonic/gin"
)

type Client struct {
	ClientID     string
	ClientSecret string
	RedirectURIs []string
	Scope        string
}

type AuthSever struct {
	AuthorizationEndpoint string
	TokenEndpoint         string
}

type CodeInfo struct {
	AuthorizationEndpointRequest *gin.Context
	Scopes                       []string
	User                         string
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}
