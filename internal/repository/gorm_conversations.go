package repository

import (
	"time"

	"github.com/coxlong/eureka/internal/model"
	"gorm.io/gorm"
)

type Message struct {
	ID             string `gorm:"primarykey;type:char(36)"`
	ConversationID string `gorm:"primarykey;type:char(36)"`
	Parent         string `gorm:"type:char(36)"`
	Role           string `gorm:"type:char(9);NOT NULL"`
	Content        string
	CreatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

type Conversation struct {
	ID            string    `gorm:"primarykey;type:char(36)"`
	UID           string    `gorm:"index"`
	Title         string    `gorm:"type:varchar(64)"`
	Model         string    `gorm:"type:char(16)"`
	MaxTokens     int       `gorm:"type:INT"`
	Temperature   float32   `gorm:"type:FLOAT"`
	CurrentNodeID string    `gorm:"type:char(36)"`
	Messages      []Message `gorm:"foreignKey:ConversationID"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func NewGormConversationRepository(db *gorm.DB) (ConversationsRepo, error) {
	err := db.AutoMigrate(&Conversation{}, &Message{})
	if err != nil {
		return nil, err
	}
	return &GormConversationRepository{db}, nil
}

type GormConversationRepository struct {
	db *gorm.DB
}

func (r *GormConversationRepository) CreateConversation(uid string, meta *model.ConversationMeta) error {
	params := Conversation{
		ID:            meta.ID,
		UID:           uid,
		Title:         meta.Title,
		Model:         meta.Model,
		MaxTokens:     meta.MaxTokens,
		Temperature:   meta.Temperature,
		CurrentNodeID: meta.CurrentNodeID,
	}
	return r.db.Create(&params).Error
}

func (r *GormConversationRepository) UpdateConversation(uid string, meta *model.ConversationMeta) error {
	params := Conversation{
		ID:            meta.ID,
		UID:           uid,
		Title:         meta.Title,
		Model:         meta.Model,
		MaxTokens:     meta.MaxTokens,
		Temperature:   meta.Temperature,
		CurrentNodeID: meta.CurrentNodeID,
	}
	return r.db.Updates(&params).Error
}

func (r *GormConversationRepository) GetConversationByID(id string, uid string) (*model.ConversationMeta, []model.Message, error) {
	var conversation Conversation
	tx := r.db.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at")
	}).Where(Conversation{ID: id, UID: uid}).First(&conversation)
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	result := model.ConversationMeta{
		ID:            conversation.ID,
		Title:         conversation.Title,
		Model:         conversation.Model,
		MaxTokens:     conversation.MaxTokens,
		Temperature:   conversation.Temperature,
		CurrentNodeID: conversation.CurrentNodeID,
		CreatedAt:     conversation.CreatedAt,
		UpdatedAt:     conversation.UpdatedAt,
	}
	messages := []model.Message{}
	for _, item := range conversation.Messages {
		messages = append(messages, model.Message{
			ID:        item.ID,
			Parent:    item.Parent,
			Role:      item.Role,
			Content:   item.Content,
			CreatedAt: item.CreatedAt,
		})
	}
	return &result, messages, nil
}

func (r *GormConversationRepository) GetConversations(uid string) ([]model.ConversationMeta, error) {
	var conversations []Conversation
	tx := r.db.Where(Conversation{UID: uid}).Order("updated_at DESC").Find(&conversations)
	if tx.Error != nil {
		return nil, tx.Error
	}
	result := []model.ConversationMeta{}
	for _, item := range conversations {
		result = append(result, model.ConversationMeta{
			ID:            item.ID,
			Title:         item.Title,
			Model:         item.Model,
			MaxTokens:     item.MaxTokens,
			Temperature:   item.Temperature,
			CurrentNodeID: item.CurrentNodeID,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
		})
	}
	return result, nil
}

func (r *GormConversationRepository) CreateMessages(conversationID string, messages []model.Message) error {
	var params []Message
	for _, item := range messages {
		params = append(params, Message{
			ID:             item.ID,
			Parent:         item.Parent,
			ConversationID: conversationID,
			Role:           item.Role,
			Content:        item.Content,
		})
	}
	return r.db.CreateInBatches(params, 100).Error
}

func (r *GormConversationRepository) Transaction(txFunc func(r ConversationsRepo) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return txFunc(&GormConversationRepository{tx})
	})
}
