package authorizationServer

import (
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"oauth2.0go/db"
	"oauth2.0go/model"
	"oauth2.0go/utils"
)

var authServer = &model.AuthSever{
	AuthorizationEndpoint: "http://localhost:9001/authorize",
	TokenEndpoint:         "http://localhost:9001/token",
}

var clients = []*model.Client{
	{
		ClientID:     "oauth-client-1",
		ClientSecret: "oauth-client-secret-1",
		RedirectURIs: []string{"http://localhost:9000/callback"},
		Scope:        "foo bar",
	},
}

var codes = map[string]*model.CodeInfo{}
var codeMutex sync.RWMutex

var getClient = func(clientID string) *model.Client {
	for _, client := range clients {
		if client.ClientID == clientID {
			return client
		}
	}
	return nil
}

var requests = map[string]*model.ClientQuery{} // 记录客户端请求
var mutex sync.Mutex

func Start() {
	r := gin.Default()

	r.LoadHTMLGlob("static/authorizationServer/*")
	r.Handle("GET", "/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"clients":    clients,
			"authServer": authServer,
		})
	})

	r.GET("/authorize", authorize)
	r.POST("/approve", approve)
	r.POST("/token", token)
	go r.Run(":9001")
}

func authorize(c *gin.Context) {
	clientID, redirectURL := c.Query("client_id"), c.Query("redirect_uri")
	client := getClient(clientID)
	// 必须已注册的客户端
	if client == nil {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"error": "Unknown client",
		})
		return
	} else if !contain(client.RedirectURIs, redirectURL) {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"error": "Invalid redirect URI",
		})
		return
	}

	// 检查客户端请求scope范围
	var rscope, cscope []string
	if reqScope := c.Query("scope"); len(reqScope) > 0 {
		rscope = strings.Split(reqScope, " ")
	}
	cscope = strings.Split(client.Scope, " ")
	if !checkScope(rscope, cscope) {
		redirect(c, redirectURL, map[string]string{"error": "invalid_scope"})
		return
	}

	// 保存客户端请求
	var reqID = utils.RandString(8)
	clientQuery := model.NewClientQuery(c.Request.URL.Query())
	mutex.Lock()
	requests[reqID] = clientQuery
	mutex.Unlock()
	logrus.Info("requests len: ", len(requests))
	logrus.Info("save request url: ", clientQuery.Vals, ", reqId: ", reqID)
	c.HTML(http.StatusOK, "approve.html", gin.H{
		"client": client,
		"reqid":  reqID,
		"scope":  rscope,
	})
}

func approve(c *gin.Context) {
	reqID := c.PostForm("reqid")
	mutex.Lock()
	clientQuery := requests[reqID]
	delete(requests, reqID) // 使用后要删除reqID对应的request信息
	mutex.Unlock()
	if clientQuery == nil {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"error": "No matching authorization request",
		})
		return
	}

	logrus.Info("get request url: ", clientQuery.Vals, ", reqId: ", reqID)
	redirectURL := clientQuery.Get("redirect_uri") // 上次请求中保存的重定向url

	approve := c.PostForm("approve")
	if approve != "" {
		// 如果是授权码模式，则生成8字符的授权码
		if clientQuery.Get("response_type") == "code" {
			code := utils.RandString(8)
			user := c.PostForm("user")
			var getScopes = func(ctx *gin.Context) []string {
				var cscope []string
				c.Request.ParseForm()
				for k := range c.Request.PostForm {
					if strings.HasPrefix(k, "scope_") {
						cscope = append(cscope, k[len("scope_"):])
					}
				}
				return cscope
			}
			rscope := getScopes(c)
			client := getClient(clientQuery.Get("client_id"))
			cscope := strings.Split(client.Scope, " ")
			if !checkScope(rscope, cscope) {
				redirect(c, redirectURL, map[string]string{"error": "invalid_scope"})
				return
			}

			// 将code保存起来，以待后续校验用
			codeMutex.Lock()
			codes[code] = &model.CodeInfo{
				AuthorizationEndpointRequest: clientQuery,
				Scopes:                       rscope,
				User:                         user,
			}
			logrus.Info("save code: ", codes[code])
			codeMutex.Unlock()
			redirect(c, redirectURL, map[string]string{"code": code, "state": clientQuery.Get("state")})
			return
		} else {
			// 非授权码类型暂时不支持
			redirect(c, redirectURL, map[string]string{"error": "unsupported_response_type"})
			return
		}
	} else {
		redirect(c, redirectURL, map[string]string{"error": "invalid_scope"})
		return
	}
}

func token(c *gin.Context) {
	var auth = c.GetHeader("authorization")
	logrus.Info("auth: ", auth)
	var clientID, clientSecret string
	if len(auth) > 0 {
		clientID, clientSecret = utils.DecodeAuth(auth)
	}
	logrus.Info("clientID: ", clientID, ", clientSecret: ", clientSecret)

	// 如果header中传递了auth，但同时body也传递clientId，则是不合规范的
	if len(c.PostForm("client_id")) > 0 {
		if len(clientID) > 0 {
			c.JSON(401, map[string]string{"error": "invalid_client"})
			return
		}
		clientID, clientSecret = c.PostForm("client_id"), c.PostForm("client_secret")
	}

	client := getClient(clientID)
	if client == nil {
		c.JSON(401, map[string]string{"error": "invalid_client"})
		return
	}
	if client.ClientSecret != clientSecret {
		c.JSON(401, map[string]string{"error": "invalid_client"})
		return
	}

	if c.PostForm("grant_type") == "authorization_code" {
		codeMutex.RLock()
		var code = codes[c.PostForm("code")]
		codeMutex.RUnlock()

		if code != nil {
			// code仅能使用一次
			codeMutex.Lock()
			delete(codes, c.PostForm("code"))
			codeMutex.Unlock()
			if code.AuthorizationEndpointRequest.Get("client_id") == clientID {
				accessToken := utils.RandString(8)
				cscope := strings.Join(code.Scopes, " ")
				db.Insert(map[string]interface{}{"access_token": accessToken, "client_id": clientID, "scope": cscope})

				tokenResponse := map[string]interface{}{
					"access_token": accessToken,
					"token_type":   "Bearer",
					"scope":        cscope,
				}
				c.JSON(200, tokenResponse)
			} else {
				logrus.Error("clientId not found")
				c.JSON(400, map[string]string{"error": "invalid_grant"})
				return
			}
		} else {
			logrus.Error("code not found")
			c.JSON(400, map[string]string{"error": "invalid_grant"})
			return
		}
	} else {
		c.JSON(400, map[string]string{"error": "unsupported_grant_type"})
		return
	}
}

func contain(list []string, s string) bool {
	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
}

func checkScope(reqScopes, clientScopes []string) bool {
	for _, v := range reqScopes {
		if !contain(clientScopes, v) {
			return false
		}
	}
	return true
}

func redirect(c *gin.Context, redirectURL string, appendKV map[string]string) {
	urlParse, _ := url.Parse(redirectURL)
	vals := urlParse.Query()
	// 在重定向url附加参数
	for k, v := range appendKV {
		vals.Set(k, v)
	}
	urlParse.RawQuery = vals.Encode()
	c.Redirect(302, urlParse.String())
}

