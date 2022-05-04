package model

type Client struct {
	ClientID     string
	ClientSecret string
	ClientName   string
	RedirectURIs []string
	Scope        string
}

type AuthSever struct {
	AuthorizationEndpoint string
	TokenEndpoint         string
}

type CodeInfo struct {
	AuthorizationEndpointRequest *ClientQuery
	Scopes                       []string
	User                         string
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}
