package memory

import (
	"context"
	"errors"
	"fmt"

	"github.com/AstroSynapseAI/app-service/engine"
	"github.com/AstroSynapseAI/app-service/models"
	"github.com/tmc/langchaingo/schema"
	"gorm.io/gorm"
)

var (
	ErrDBConnection     = errors.New("can't connect to database")
	ErrDBMigration      = errors.New("can't migrate database")
	ErrMissingSessionID = errors.New("session id can not be empty")
	InitiativePrompt    = "New user, has connected."
)

type PersistentChatHistory struct {
	db        *gorm.DB
	records   *models.ChatHistory
	messages  []schema.ChatMessage
	sessionID string
}

var _ schema.ChatMessageHistory = &PersistentChatHistory{}

func NewPersistentChatHistory(config engine.AvatarConfig) *PersistentChatHistory {
	history := &PersistentChatHistory{}
	history.db = config.GetDB().Adapter.Gorm()

	err := history.db.AutoMigrate(models.ChatHistory{})
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return history
}

func (history *PersistentChatHistory) GetSessionID() string {
	return history.sessionID
}

func (history *PersistentChatHistory) SetSessionID(id string) {
	history.sessionID = id
}

func (history *PersistentChatHistory) Messages(context.Context) ([]schema.ChatMessage, error) {
	if history.sessionID == "" {
		return []schema.ChatMessage{}, ErrMissingSessionID
	}

	err := history.db.Where(models.ChatHistory{SessionID: history.sessionID}).Find(&history.records).Error
	if err != nil {
		return nil, err
	}

	history.messages = []schema.ChatMessage{}

	if history.records.ChatHistory != nil {
		for i := range *history.records.ChatHistory {
			msg := (*history.records.ChatHistory)[i]

			if msg.Type == "human" {
				history.messages = append(history.messages, schema.HumanChatMessage{Content: msg.Content})
			}

			if msg.Type == "ai" {
				history.messages = append(history.messages, schema.AIChatMessage{Content: msg.Content})
			}
		}
	}

	return history.messages, nil
}

func (history *PersistentChatHistory) AddMessage(ctx context.Context, message schema.ChatMessage) error {
	if history.sessionID == "" {
		return ErrMissingSessionID
	}

	if message.GetContent() == InitiativePrompt {
		return nil
	}

	history.messages = append(history.messages, message)
	bufferString, err := schema.GetBufferString(history.messages, "Human", "AI")
	if err != nil {
		return err
	}

	history.records.SessionID = history.sessionID
	history.records.ChatHistory = history.loadNewMessages()
	history.records.BufferString = bufferString

	err = history.db.Save(&history.records).Error
	if err != nil {
		return err
	}

	return nil
}

func (history *PersistentChatHistory) AddAIMessage(ctx context.Context, message string) error {
	return history.AddMessage(ctx, schema.AIChatMessage{Content: message})
}

func (history *PersistentChatHistory) AddUserMessage(ctx context.Context, message string) error {
	return history.AddMessage(ctx, schema.HumanChatMessage{Content: message})
}

func (history *PersistentChatHistory) SetMessages(ctx context.Context, messages []schema.ChatMessage) error {
	if history.sessionID == "" {
		return ErrMissingSessionID
	}

	history.messages = messages
	bufferString, err := schema.GetBufferString(history.messages, "Human", "AI")
	if err != nil {
		return err
	}

	history.records.SessionID = history.sessionID
	history.records.ChatHistory = history.loadNewMessages()
	history.records.BufferString = bufferString

	err = history.db.Save(&history.records).Error
	if err != nil {
		return err
	}

	return nil
}

func (history *PersistentChatHistory) Clear(context.Context) error {
	history.messages = []schema.ChatMessage{}

	err := history.db.Where(models.ChatHistory{SessionID: history.sessionID}).Delete(&history.records).Error
	if err != nil {
		return err
	}

	return nil
}

func (history *PersistentChatHistory) loadNewMessages() *models.Messages {
	newMsgs := models.Messages{}
	for _, msg := range history.messages {
		newMsgs = append(newMsgs, models.Message{
			Type:    string(msg.GetType()),
			Content: msg.GetContent(),
		})
	}

	return &newMsgs
}
