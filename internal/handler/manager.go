package handler

import (
	"github.com/coxlong/eureka/internal/pkg/config"
	"github.com/coxlong/eureka/internal/service"
)

type Manager struct {
	Auth          AuthHandler
	Chat          ChatHandler
	Conversations ConversationsHandler
}

func NewManager(cfg *config.Config, conversationsService service.ConversationsService) *Manager {
	return &Manager{
		Auth:          NewDefaultAuthHandler(cfg.Authorization.GithubClient, cfg.Authorization.GithubClientSecret, cfg.Env.FrontendAddr),
		Chat:          NewChatHandler(conversationsService, cfg.OpenAI.BaseURL),
		Conversations: NewConversationHandler(conversationsService),
	}
}
