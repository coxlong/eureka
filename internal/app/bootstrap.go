package app

import (
	"encoding/gob"
	"errors"

	"github.com/coxlong/eureka/internal/handler"
	"github.com/coxlong/eureka/internal/model"
	"github.com/coxlong/eureka/internal/pkg/config"
	"github.com/coxlong/eureka/internal/pkg/log"
	"github.com/coxlong/eureka/internal/repository"
	"github.com/coxlong/eureka/internal/router"
	"github.com/coxlong/eureka/internal/service"
	gormsessions "github.com/gin-contrib/sessions/gorm"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Bootstrap(cfg *config.Config) (*gin.Engine, error) {
	gob.Register(model.User{})
	if cfg.Env.Mode == config.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化日志服务
	_, err := log.InitLogger(cfg.Env.Mode, &cfg.Logger)
	if err != nil {
		return nil, err
	}

	db, err := initDB(&cfg.Database)
	if err != nil {
		return nil, err
	}
	sessionStore := gormsessions.NewStore(db, true, []byte(cfg.Authorization.SessionKey))

	conversationsRepo, err := repository.NewGormConversationRepository(db)
	if err != nil {
		return nil, err
	}
	conversationsService := service.NewConversationService(conversationsRepo)

	return router.Setup(&cfg.Env, sessionStore, handler.NewManager(cfg, conversationsService))
}

func initDB(cfg *config.Database) (*gorm.DB, error) {
	switch cfg.Type {
	case "sqlite":
		return gorm.Open(sqlite.Open(cfg.SQLitePath), &gorm.Config{})
	case "mysql":
		return gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	default:
		return nil, errors.New("无效的数据库类型")
	}
}
