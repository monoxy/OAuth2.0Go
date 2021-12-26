package main

//
//import (
//	"net/http"
//
//	"github.com/gin-gonic/gin"
//
//	"oauth2.0go/model"
//)
//
//const (
//	accessToken  = "null"
//	refreshToken = "null"
//	scope        = "null"
//
//	protectedResource = "http://localhost:9002/resource"
//	state             = "null"
//)
//
//var client = &model.Client{
//	ClientID:     "oauth-client-1",
//	ClientSecret: "oauth-client-secret-1",
//	RedirectURIs: []string{"http://localhost:9000/callback"},
//}
//
//func startClient() {
//	r := gin.Default()
//
//	r.LoadHTMLGlob("static/client/*")
//	r.Handle("GET", "/", func(c *gin.Context) {
//		c.HTML(http.StatusOK, "index.html", gin.H{
//			"access_token":  accessToken,
//			"scope":         scope,
//			"refresh_token": refreshToken,
//		})
//	})
//	r.Handle("GET", "/authorize", func(c *gin.Context) {})
//	go r.Run(":9000")
//}
