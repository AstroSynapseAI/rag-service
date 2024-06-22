package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/AstroSynapseAI/rag-service/config"
)

const (
	DefaultServerURL = "http://localhost:8080/api"
)

func TestMain(m *testing.M) {
	// Setup: start the server
	go func() {
		app := config.New()

		app.LoadEnvironment()

		app.InitDB()

		err := app.RunServer()
		if err != nil {
			fmt.Println("Failed to run server:", err)
			return
		}
	}()

	// give the server some time to start
	time.Sleep(time.Second * 2)

	// Run the tests
	code := m.Run()
	os.Exit(code)
}
