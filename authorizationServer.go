package main

//
//import (
//	"fmt"
//	"math/rand"
//	"net/http"
//	"net/url"
//	"strings"
//	"sync"
//
//	"github.com/gin-gonic/gin"
//
//	"oauth2.0go/db"
//	"oauth2.0go/model"
//)
//
//var authServer = &model.AuthSever{
//	AuthorizationEndpoint: "http://localhost:9001/authorize",
//	TokenEndpoint:         "http://localhost:9001/token",
//}
//
//var clients = []*model.Client{
//	{
//		ClientID:     "oauth-client-1",
//		ClientSecret: "oauth-client-secret-1",
//		RedirectURIs: []string{"http://localhost:9000/callback"},
//		Scope:        "foo bar",
//	},
//}
//
//var codes = map[string]*model.CodeInfo{}
//var codeMutex sync.RWMutex
//
//var getClient = func(clientID string) *model.Client {
//	for _, client := range clients {
//		if client.ClientID == clientID {
//			return client
//		}
//	}
//	return nil
//}
//
//var requests = map[string]*gin.Context{} // 记录客户端请求
//var mutex sync.RWMutex
//
//func startAuthorizationServer() {
//	r := gin.Default()
//
//	r.LoadHTMLGlob("static/authorizationServer/*")
//	r.Handle("GET", "/", func(c *gin.Context) {
//		c.HTML(http.StatusOK, "index.html", gin.H{
//			"clients":    clients,
//			"authServer": authServer,
//		})
//	})
//
//	r.GET("/authorize", authorize)
//	r.GET("/approve", approve)
//	r.GET("/token", token)
//	go r.Run(":9001")
//}
//
//func authorize(c *gin.Context) {
//	clientID, redirectURL := c.Query("client_id"), c.Query("redirect_uri")
//	client := getClient(clientID)
//	// 必须已注册的客户端
//	if client == nil {
//		c.HTML(http.StatusOK, "error.html", gin.H{
//			"error": "Unknown client",
//		})
//		return
//	} else if !contain(redirectURL, client.RedirectURIs) {
//		c.HTML(http.StatusOK, "error.html", gin.H{
//			"error": "Invalid redirect URI",
//		})
//		return
//	}
//
//	// 检查客户端请求scope范围
//	var rscope, cscope []string
//	if reqScope := c.Query("scope"); len(reqScope) > 0 {
//		rscope = strings.Split(reqScope, " ")
//	}
//	cscope = strings.Split(client.Scope, " ")
//	if !checkScope(rscope, cscope) {
//		redirect(c, map[string]string{"error": "invalid_scope"})
//		return
//	}
//
//	// 保存客户端请求
//	var reqID = randString(8)
//	mutex.Lock()
//	requests[reqID] = c
//	mutex.Unlock()
//	c.HTML(http.StatusOK, "approve.html", gin.H{
//		"client": client,
//		"reqid":  reqID,
//		"scope":  rscope,
//	})
//}
//
//func approve(c *gin.Context) {
//	reqID := c.PostForm("reqid")
//	mutex.RLock()
//	query := requests[reqID]
//	delete(requests, reqID) // 使用后要删除reqID对应的request信息
//	mutex.RUnlock()
//	if query == nil {
//		c.HTML(http.StatusOK, "error.html", gin.H{
//			"error": "No matching authorization request",
//		})
//		return
//	}
//
//	approve := c.PostForm("approve")
//	if len(approve) > 0 {
//		// 如果是授权码模式，则生成8字符的授权码
//		if query.Query("response_type") == "code" {
//			code := randString(8)
//			user := c.PostForm("user")
//			var getScopes = func(ctx *gin.Context) []string {
//				var cscope []string
//				c.Request.ParseForm()
//				for k := range c.Request.PostForm {
//					if strings.HasPrefix(k, "scope_") {
//						cscope = append(cscope, k[len("scope_"):])
//					}
//				}
//				return cscope
//			}
//			rscope := getScopes(c)
//			client := getClient(query.Query("client_id"))
//			cscope := strings.Split(client.Scope, " ")
//			if !checkScope(rscope, cscope) {
//				redirect(c, map[string]string{"error": "invalid_scope"})
//				return
//			}
//
//			// 将code保存起来，以待后续校验用
//			codeMutex.Lock()
//			codes[code] = &model.CodeInfo{
//				AuthorizationEndpointRequest: c,
//				Scopes:                       rscope,
//				User:                         user,
//			}
//			codeMutex.Unlock()
//			redirect(c, map[string]string{"code": code, "state": query.Query("state")})
//			return
//		} else {
//			// 非授权码类型暂时不支持
//			redirect(c, map[string]string{"error": "unsupported_response_type"})
//			return
//		}
//	} else {
//		redirect(c, map[string]string{"error": "invalid_scope"})
//		return
//	}
//}
//
//func token(c *gin.Context) {
//	var auth = c.GetHeader("authorization")
//	var clientID, clientSecret string
//	if len(auth) > 0 {
//		clientID, clientSecret = decodeAuth(auth)
//	}
//
//	// 如果header中传递了auth，但同时body也传递clientId，则是不合规范的
//	if len(c.PostForm("client_id")) > 0 {
//		if len(clientID) > 0 {
//			c.JSON(401, map[string]string{"error": "invalid_client"})
//			return
//		}
//		clientID, clientSecret = c.PostForm("client_id"), c.PostForm("client_secret")
//	}
//
//	client := getClient(clientID)
//	if client == nil {
//		c.JSON(401, map[string]string{"error": "invalid_client"})
//		return
//	}
//	if client.ClientSecret != clientSecret {
//		c.JSON(401, map[string]string{"error": "invalid_client"})
//		return
//	}
//
//	if c.PostForm("grant_type") == "authorization_code" {
//		codeMutex.RLock()
//		var code = codes[c.PostForm("code")]
//		codeMutex.RUnlock()
//
//		if code != nil {
//			// code仅能使用一次
//			codeMutex.Lock()
//			delete(codes, c.PostForm("code"))
//			codeMutex.Unlock()
//			if code.AuthorizationEndpointRequest.Query("client_id") == clientID {
//				accessToken := randString(8)
//				cscope := strings.Join(code.Scopes, " ")
//				db.Insert(map[string]interface{}{"access_token": accessToken, "client_id": clientID, "scope": cscope})
//
//				tokenResponse := map[string]interface{}{
//					"access_token": accessToken,
//					"token_type":   "Bearer",
//					"scope":        cscope,
//				}
//				c.JSON(200, tokenResponse)
//			} else {
//				c.JSON(400, map[string]string{"error": "invalid_grant"})
//				return
//			}
//		} else {
//			c.JSON(400, map[string]string{"error": "invalid_grant"})
//			return
//		}
//	} else {
//		c.JSON(400, map[string]string{"error": "unsupported_grant_type"})
//		return
//	}
//}
//
//func contain(s string, list []string) bool {
//	for _, v := range list {
//		if s == v {
//			return true
//		}
//	}
//	return false
//}
//
//func checkScope(reqScopes, clientScopes []string) bool {
//	for _, v := range reqScopes {
//		if !contain(v, clientScopes) {
//			return false
//		}
//	}
//	return true
//}
//
//func randString(n int) string {
//	key := make([]byte, n)
//	_, err := rand.Read(key)
//	if err != nil {
//		panic(err)
//	}
//
//	return fmt.Sprintf("%x", key)
//}
//
//func redirect(c *gin.Context, appendKV map[string]string) {
//	urlParse, _ := url.Parse(c.Query("redirect_uri"))
//	vals := urlParse.Query()
//	// 在重定向url附加参数
//	for k, v := range appendKV {
//		vals.Set(k, v)
//	}
//	urlParse.RawQuery = vals.Encode()
//	c.Redirect(302, urlParse.String())
//}
//
//func decodeAuth(auth string) (clientID, clientSecret string) {
//	if strings.HasPrefix(auth, "basic ") {
//		s := auth[len("basic "):]
//		list := strings.Split(s, ":")
//		if len(list) > 2 {
//			clientID, clientSecret = list[0], list[1]
//		}
//	}
//	return
//}
