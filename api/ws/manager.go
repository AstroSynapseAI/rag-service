package ws

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/GoLangWebSDK/crud/database"
	"github.com/gorilla/websocket"
)

var websocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Manager struct {
	sync.RWMutex
	clients map[*Client]bool
	db      *database.Database
}

func NewManager(db *database.Database) *Manager {
	mng := &Manager{
		clients: make(map[*Client]bool),
		db:      db,
	}

	return mng
}

func (m *Manager) Handler(w http.ResponseWriter, r *http.Request) {
	websocketUpgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to initate RAG socket:", err)
		return
	}

	client := NewClient(conn, m)

	m.addClient(client)

	ctx := context.Background()

	go client.MaintainConnection(ctx)
	go client.ReadMsgs(ctx)
	go client.SendMsgs(ctx)
}

func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()
	m.clients[client] = true
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}
}

func (m *Manager) checkOrigin(allowed []string) {
	websocketUpgrader.CheckOrigin = func(r *http.Request) bool {
		for _, v := range allowed {
			if r.Header.Get("Origin") == v {
				return true
			}
		}

		return false
	}
}
