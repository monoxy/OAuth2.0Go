package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"oauth2.0go/model"
	"oauth2.0go/utils"
)

const (
	protectedResource = "http://localhost:9002/resource"
)

var (
	state        = ""
	accessToken  = "null"
	refreshToken = "null"
	scope        = "null"
)

var authServer = &model.AuthSever{
	AuthorizationEndpoint: "http://localhost:9001/authorize",
	TokenEndpoint:         "http://localhost:9001/token",
}

var client = &model.Client{
	ClientID:     "oauth-client-1",
	ClientSecret: "oauth-client-secret-1",
	RedirectURIs: []string{"http://localhost:9000/callback"},
	Scope:        "",
}

func Start() {
	r := gin.Default()

	r.LoadHTMLGlob("static/client/*")
	r.Handle("GET", "/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"access_token":  accessToken,
			"scope":         scope,
			"refresh_token": refreshToken,
		})
	})

	r.Handle("GET", "/authorize", authorize)
	r.Handle("GET", "/callback", callback)
	r.Handle("GET", "/fetch_resource", fetchResource)
	go r.Run(":9000")
}

func authorize(c *gin.Context) {
	state = utils.RandString(8)
	urlParse, _ := url.Parse(authServer.AuthorizationEndpoint)
	vals := urlParse.Query()
	vals.Set("response_type", "code")
	vals.Set("scope", client.Scope)
	vals.Set("client_id", client.ClientID)
	vals.Set("redirect_uri", client.RedirectURIs[0])
	vals.Set("state", state)
	urlParse.RawQuery = vals.Encode()
	logrus.Info("redirect: ", urlParse.String())
	c.Redirect(302, urlParse.String())
}

func callback(c *gin.Context) {
	error := c.Query("error")
	if len(error) > 0 {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": error})
		return
	}

	resState := c.Query("state")
	if resState != state {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": "State value did not match"})
		return
	}

	code := c.Query("code")
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", code)
	formData.Set("redirect_uri", client.RedirectURIs[0])
	header := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"Authorization": fmt.Sprintf("Basic %s:%s",
			url.QueryEscape(client.ClientID),
			base64.StdEncoding.EncodeToString([]byte(client.ClientSecret))),
	}
	body, err := utils.Post(authServer.TokenEndpoint, formData.Encode(), header)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}
	rsp := model.AuthResponse{}
	err = json.Unmarshal(body, &rsp)
	if err != nil {
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}

	accessToken = rsp.AccessToken
	scope = rsp.Scope
	if rsp.RefreshToken != "" {
		refreshToken = rsp.RefreshToken
	}
	c.HTML(http.StatusOK, "index.html", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"scope":         scope,
	})
}

func fetchResource(c *gin.Context) {
	logrus.Info("Making request with access_token: ", accessToken)
	header := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
		"Content-Type":  "application/x-www-form-urlencoded",
	}
	body, err := utils.Post(protectedResource, "", header)
	if err != nil {
		accessToken = "null"
		c.HTML(http.StatusOK, "error.html", gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "data.html", gin.H{"data": string(body)})
}
