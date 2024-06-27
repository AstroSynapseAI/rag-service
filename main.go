package main

import (
	"fmt"
	"github.com/AstroSynapseAI/rag-service/config"
)

func main() {

	app := config.New()

	app.LoadEnvironment()

	app.InitDB()

	err := app.RunServer()
	if err != nil {
		fmt.Println("Failed to run server:", err)
		return
	}
}
