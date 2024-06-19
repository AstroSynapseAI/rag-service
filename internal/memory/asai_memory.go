package memory

import (
	"context"
	"fmt"

	"github.com/AstroSynapseAI/rag-service/internal"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
)

type AsaiMemory struct {
	buffer      schema.Memory
	chatHistory *PersistentChatHistory
}

// NewMemory creates a new instance of AsaiMemory.
//
// It takes a dsn postgred string as a parameter and returns a pointer to AsaiMemory.
func NewMemory(config internal.AvatarConfig) *AsaiMemory {
	chatHistory := NewPersistentChatHistory(config)

	buffer := memory.NewConversationTokenBuffer(
		config.GetAvatarLLM(),
		config.GetAvatarMemorySize(),
		memory.WithChatHistory(chatHistory),
	)

	return &AsaiMemory{
		buffer:      buffer,
		chatHistory: chatHistory,
	}
}

// GetSessionID returns the session ID of the AsaiMemory.
//
// No parameters.
// Returns a string representing the session ID.
func (m *AsaiMemory) GetSessionID() string {
	return m.chatHistory.GetSessionID()
}

// SetSessionID sets the session ID for the AsaiMemory instance.
//
// Parameters:
//
//	id (string): The session ID to set.
//
// Return:
//
//	None.
func (m *AsaiMemory) SetSessionID(id string) {
	m.chatHistory.SetSessionID(id)
}

// Buffer returns the memory buffer of the AsaiMemory.
//
// It does not take any parameters.
// It returns a schema.Memory object.
func (m *AsaiMemory) Buffer() schema.Memory {
	return m.buffer
}

func (m *AsaiMemory) Messages() []llms.ChatMessage {
	msgs, err := m.chatHistory.Messages(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return msgs
}
