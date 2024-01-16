package service

import (
	"github.com/coxlong/eureka/internal/model"
	"github.com/coxlong/eureka/internal/repository"
)

type ConversationsService interface {
	CreateConversation(uid string, meta *model.ConversationMeta, messages []model.Message) error
	UpdateConversation(uid string, meta *model.ConversationMeta, messages []model.Message) error
	GetConversation(cid string, uid string) (*model.ConversationMeta, []model.Message, error)
	GetConversations(uid string) ([]model.ConversationMeta, error)
}

func NewConversationService(r repository.ConversationsRepo) ConversationsService {
	return &DefaultConversationService{r}
}

type DefaultConversationService struct {
	repo repository.ConversationsRepo
}

func (s *DefaultConversationService) CreateConversation(uid string, meta *model.ConversationMeta, messages []model.Message) error {
	return s.repo.Transaction(func(r repository.ConversationsRepo) error {
		if err := r.CreateConversation(uid, meta); err != nil {
			return err
		}
		if err := r.CreateMessages(meta.ID, messages); err != nil {
			return err
		}
		return nil
	})
}

func (s *DefaultConversationService) UpdateConversation(uid string, meta *model.ConversationMeta, messages []model.Message) error {
	return s.repo.Transaction(func(r repository.ConversationsRepo) error {
		if err := r.CreateMessages(meta.ID, messages); err != nil {
			return err
		}
		if err := r.UpdateConversation(uid, meta); err != nil {
			return err
		}
		return nil
	})
}

func (s *DefaultConversationService) GetConversation(cid string, uid string) (*model.ConversationMeta, []model.Message, error) {
	return s.repo.GetConversationByID(cid, uid)
}

func (s *DefaultConversationService) GetConversations(uid string) ([]model.ConversationMeta, error) {
	return s.repo.GetConversations(uid)
}
