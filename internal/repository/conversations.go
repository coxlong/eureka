package repository

import "github.com/coxlong/eureka/internal/model"

type ConversationsRepo interface {
	CreateConversation(uid string, meta *model.ConversationMeta) error
	UpdateConversation(uid string, meta *model.ConversationMeta) error
	GetConversationByID(id string, uid string) (*model.ConversationMeta, []model.Message, error)
	GetConversations(uid string) ([]model.ConversationMeta, error)
	CreateMessages(conversationID string, messages []model.Message) error
	Transaction(txFunc func(r ConversationsRepo) error) error
}
