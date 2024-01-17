package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/coxlong/eureka/internal/model"
	"github.com/coxlong/eureka/internal/pkg/constants"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type AuthHandler interface {
	GetUserInfo(*gin.Context)
	Login(c *gin.Context)
	Logout(c *gin.Context)
	Callback(c *gin.Context)
}

var (
	stateCache *expirable.LRU[string, any]
)

func init() {
	stateCache = expirable.NewLRU[string, any](5, nil, time.Minute)
}

func NewDefaultAuthHandler(clientID, clientSecret, frontendAddr string) *DefaultAuthHandler {
	return &DefaultAuthHandler{
		githubOauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"user:email"},
			Endpoint:     github.Endpoint,
		},
		frontendAddr: frontendAddr,
	}
}

type DefaultAuthHandler struct {
	frontendAddr      string
	githubOauthConfig *oauth2.Config
}

func (a *DefaultAuthHandler) GetUserInfo(c *gin.Context) {
	session := sessions.Default(c)
	if value := session.Get(constants.UserSessionKey); value != nil {
		c.JSON(200, value)
		return
	}
	c.String(401, http.StatusText(http.StatusUnauthorized))
}

// 用户登录
func (a *DefaultAuthHandler) Login(c *gin.Context) {
	state, err := generateRandomToken(32)
	if err != nil {
		c.String(500, err.Error())
		return
	}
	stateCache.Add(state, nil)
	url := a.githubOauthConfig.AuthCodeURL(state)
	c.String(200, url)
}

func (a *DefaultAuthHandler) Callback(c *gin.Context) {
	// TODO
	// provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	_, ok := stateCache.Get(state)
	if !ok {
		c.String(401, "invalid request")
		return
	}
	stateCache.Remove(state)

	user, err := a.getGithubUserInfo(c, code)
	if err != nil {
		c.String(401, err.Error())
		return
	}
	session := sessions.Default(c)
	session.Set(constants.UserSessionKey, user)
	session.Save()
	c.Redirect(302, a.frontendAddr)
}

// 用户登出
func (a *DefaultAuthHandler) Logout(c *gin.Context) {
	// 登出逻辑
}

func (a *DefaultAuthHandler) getGithubUserInfo(ctx context.Context, code string) (*model.User, error) {
	// 从 code 参数中获取访问令牌
	token, err := a.githubOauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	// 使用访问令牌访问 GitHub API
	client := a.githubOauthConfig.Client(context.Background(), token)
	userResp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer userResp.Body.Close()
	var user struct {
		ID     int64  `json:"id"`
		Login  string `json:"login"`
		Avatar string `json:"avatar_url"`
	}
	err = json.NewDecoder(userResp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}
	emailResp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return nil, err
	}
	defer emailResp.Body.Close()
	var email []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	err = json.NewDecoder(emailResp.Body).Decode(&email)
	if err != nil {
		return nil, err
	}
	var primaryEmail string
	for _, item := range email {
		if item.Primary {
			primaryEmail = item.Email
			break
		}
	}
	return &model.User{
		ID:        fmt.Sprintf("github_%d", user.ID),
		OAuthType: "github",
		OAuthID:   strconv.FormatInt(user.ID, 10),
		Username:  user.Login,
		Email:     primaryEmail,
		Avatar:    user.Avatar,
	}, nil
}

func generateRandomToken(n int) (string, error) {
	// 创建足够的随机字节
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// 将字节编码为字符串，这里使用 Base64 编码
	return base64.URLEncoding.EncodeToString(b), nil
}
