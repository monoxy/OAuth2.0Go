package main

//
//import (
//	"net/http"
//	"strings"
//
//	"github.com/gin-gonic/gin"
//	"github.com/sirupsen/logrus"
//
//	"oauth2.0go/db"
//)
//
//var resource = map[string]string{
//	"name":        "Protected Resource",
//	"description": "This data has been protected by OAuth 2.0",
//}
//
//func startProtectedResource() {
//	r := gin.Default()
//
//	r.LoadHTMLGlob("static/protectedResource/*")
//	r.Handle("GET", "/", func(c *gin.Context) {
//		c.HTML(http.StatusOK, "index.html", gin.H{})
//	})
//
//	r.OPTIONS("/resource", cors())
//	r.POST("/resource", cors(), getAccessToken(), func(c *gin.Context) {
//		if _, ok := c.Get("access_token"); ok {
//			c.JSON(http.StatusOK, resource)
//		} else {
//			c.Status(401)
//		}
//	})
//	go r.Run(":9002")
//}
//
//func cors() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		method := c.Request.Method
//		origin := c.Request.Header.Get("Origin") //请求头部
//		if origin != "" {
//			//接收客户端发送的origin （重要！）
//			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
//			//服务器支持的所有跨域请求的方法
//			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
//			//允许跨域设置可以返回其他子段，可以自定义字段
//			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session")
//			// 允许浏览器（客户端）可以解析的头部 （重要）
//			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
//			//设置缓存时间
//			c.Header("Access-Control-Max-Age", "172800")
//			//允许客户端传递校验信息比如 cookie (重要)
//			c.Header("Access-Control-Allow-Credentials", "true")
//		}
//
//		//允许类型校验
//		if method == "OPTIONS" {
//			c.JSON(http.StatusOK, "ok!")
//		}
//
//		defer func() {
//			if err := recover(); err != nil {
//				logrus.Printf("Panic info is: %v", err)
//			}
//		}()
//
//		c.Next()
//	}
//}
//
//func getAccessToken() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		auth := c.GetHeader("authorization")
//		var inToken string
//		if len(auth) > 0 && strings.HasPrefix(strings.ToLower(auth), "bearer ") {
//			inToken = auth[len("bearer "):]
//		} else if t := c.PostForm("access_token"); len(t) > 0 {
//			// 不在header就去body中查询
//			inToken = t
//		} else if t := c.Query("access_token"); len(t) > 0 {
//			// 最后去url参数中查询
//			inToken = t
//		}
//
//		logrus.Info("request token: ", inToken)
//		db.Insert(nil)
//
//		c.Set("access_token", inToken)
//		c.Next()
//	}
//}
