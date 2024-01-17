package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/coxlong/eureka/internal/model"
	"github.com/coxlong/eureka/internal/pkg/constants"
	"github.com/coxlong/eureka/internal/pkg/log"
	"github.com/coxlong/eureka/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

type ChatHandler interface {
	Completions(*gin.Context)
}

func NewChatHandler(service service.ConversationsService, baseURL string) ChatHandler {
	return &DefaultChatHandler{service, baseURL}
}

type ChatCompletionRequest struct {
	*openai.ChatCompletionRequest
	ID       string          `json:"id"`
	Messages []model.Message `json:"messages"`
	Save     bool            `json:"save"`
}

func (req *ChatCompletionRequest) UnmarshalJSON(data []byte) error {
	type Alias ChatCompletionRequest
	var aux Alias
	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}
	for _, item := range aux.Messages {
		aux.ChatCompletionRequest.Messages = append(aux.ChatCompletionRequest.Messages, openai.ChatCompletionMessage{
			Role:    item.Role,
			Content: item.Content,
		})
	}
	*req = ChatCompletionRequest(aux)
	return nil
}

type DefaultChatHandler struct {
	service service.ConversationsService
	baseURL string
}

func (h *DefaultChatHandler) Completions(c *gin.Context) {
	var req ChatCompletionRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(400, openai.ErrorResponse{
			Error: &openai.APIError{
				Message: err.Error(),
			},
		})
		return
	}

	authHeader := c.Request.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(400, openai.ErrorResponse{
			Error: &openai.APIError{
				Message: "Invalid Authorization",
			},
		})
		return
	}
	tokenString := authHeader[7:]

	config := openai.DefaultConfig(tokenString)
	if h.baseURL != "" {
		config.BaseURL = h.baseURL
	}
	client := openai.NewClientWithConfig(config)

	if !req.Stream {
		response, err := client.CreateChatCompletion(c, *req.ChatCompletionRequest)
		if err != nil {
			eResp := toOpenaiErrorResponse(err)
			c.JSON(eResp.Error.HTTPStatusCode, eResp)
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}

	stream, err := client.CreateChatCompletionStream(c, *req.ChatCompletionRequest)
	if err != nil {
		eResp := toOpenaiErrorResponse(err)
		c.JSON(eResp.Error.HTTPStatusCode, eResp)
		return
	}
	defer stream.Close()

	var answer string
	answerID := uuid.NewString()
	c.Header("Content-Type", "text/event-stream")
	c.Stream(func(w io.Writer) bool {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			w.Write([]byte("data: [DONE]\n\n"))
			return false
		}
		if err != nil {
			return false
		}
		answer += response.Choices[0].Delta.Content
		response.ID = answerID
		rByte, err := json.Marshal(response)

		if err != nil {
			return false
		}

		w.Write([]byte("data: "))
		w.Write(rByte)
		w.Write([]byte("\n\n"))
		return true
	})
	if req.Save {
		if err := h.save(c, &req, answer, answerID); err != nil {
			log.Error("save failed", zap.Error(err))
		}
	}
}

func toOpenaiErrorResponse(err error) openai.ErrorResponse {
	if e, ok := err.(*openai.APIError); ok {
		return openai.ErrorResponse{
			Error: e,
		}
	}
	return openai.ErrorResponse{
		Error: &openai.APIError{
			HTTPStatusCode: 500,
			Message:        err.Error(),
		},
	}
}

func (h *DefaultChatHandler) save(c *gin.Context, req *ChatCompletionRequest, answer, answerID string) error {
	meta := model.ConversationMeta{
		ID:            req.ID,
		Model:         req.Model,
		MaxTokens:     req.MaxTokens,
		Temperature:   req.Temperature,
		CurrentNodeID: answerID,
	}
	messages := []model.Message{}
	for _, item := range req.Messages {
		if item.ID != "" {
			messages = append(messages, item)
		}
	}
	parent := messages[len(messages)-1].ID
	messages = append(messages, model.Message{
		ID:      answerID,
		Parent:  parent,
		Role:    "assistant",
		Content: answer,
	})

	user := c.Value(constants.UserSessionKey).(model.User)
	if meta.ID == "" {
		meta.ID = answerID
		return h.service.CreateConversation(user.ID, &meta, messages)
	} else {
		return h.service.UpdateConversation(user.ID, &meta, messages)
	}
}
