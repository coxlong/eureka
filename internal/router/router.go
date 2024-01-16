package router

import (
	"time"

	"github.com/coxlong/eureka/internal/handler"
	"github.com/coxlong/eureka/internal/middleware"
	"github.com/coxlong/eureka/internal/pkg/config"
	"github.com/coxlong/eureka/internal/pkg/constants"
	"github.com/coxlong/eureka/internal/pkg/log"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
)

func Setup(env *config.Env, store sessions.Store, handlerManager *handler.Manager) (*gin.Engine, error) {
	engine := gin.New()
	engine.Use(ginzap.Ginzap(log.GetLogger(), time.RFC3339, true))
	engine.Use(gin.Recovery())
	engine.Use(sessions.Sessions(constants.UserSessionName, store))

	// 允许跨域
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{env.FrontendAddr}
	config.AllowHeaders = append(config.AllowHeaders, "Authorization")
	config.AllowCredentials = true

	engine.Use(cors.New(config))

	router := engine.Group("/api")

	// 注册鉴权路由
	setupAuthRouter(router.Group("/auth"), handlerManager.Auth)

	// 注册鉴权中间件
	router.Use(middleware.CheckLogin)

	// 注册/chat/completions接口
	router.POST("/chat/completions", handlerManager.Chat.Completions)

	// 注册conversations接口
	setupConversationsRouter(router.Group("/conversations"), handlerManager.Conversations)

	return engine, nil
}

func setupAuthRouter(router *gin.RouterGroup, handle handler.AuthHandler) {
	router.GET("/user", handle.GetUserInfo)
	router.GET("/login/:provider", handle.Login)
	router.GET("/callback/:provider", handle.Callback)
}

func setupConversationsRouter(router *gin.RouterGroup, handle handler.ConversationsHandler) {
	router.GET("/:id", handle.GetConversation)
	router.GET("/", handle.GetConversations)
}
