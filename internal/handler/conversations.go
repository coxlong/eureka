package handler

import (
	"github.com/coxlong/eureka/internal/model"
	"github.com/coxlong/eureka/internal/pkg/constants"
	"github.com/coxlong/eureka/internal/service"
	"github.com/gin-gonic/gin"
)

type ConversationsHandler interface {
	GetConversation(*gin.Context)
	GetConversations(*gin.Context)
}

func NewConversationHandler(service service.ConversationsService) ConversationsHandler {
	return &DefaultConversationsHandler{service}
}

type DefaultConversationsHandler struct {
	service service.ConversationsService
}

func (h *DefaultConversationsHandler) GetConversation(c *gin.Context) {
	cid := c.Param("id")
	user := c.Value(constants.UserSessionKey).(model.User)
	meta, messages, err := h.service.GetConversation(cid, user.ID)
	if err != nil {
		c.String(400, err.Error())
		return
	}
	messagesMap := map[string]model.Message{}
	for _, item := range messages {
		messagesMap[item.ID] = item
	}
	c.JSON(200, map[string]any{
		"meta":     meta,
		"messages": messagesMap,
	})
}
func (h *DefaultConversationsHandler) GetConversations(c *gin.Context) {
	user := c.Value(constants.UserSessionKey).(model.User)
	conversations, err := h.service.GetConversations(user.ID)
	if err != nil {
		c.String(400, err.Error())
		return
	}
	c.JSON(200, conversations)
}
