package middleware

import (
	"net/http"

	"github.com/coxlong/eureka/internal/pkg/constants"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func CheckLogin(c *gin.Context) {
	session := sessions.Default(c)
	userInfo := session.Get(constants.UserSessionKey)
	if userInfo == nil {
		c.String(401, http.StatusText(http.StatusUnauthorized))
		c.Abort()
		return
	}
	c.Set(constants.UserSessionKey, userInfo)
	c.Next()
}
