package config

import (
	"fmt"
	"net/http"
	"os"

	"github.com/AstroSynapseAI/rag-service/api/ws"
	"github.com/GoLangWebSDK/crud/database"
	"github.com/GoLangWebSDK/crud/database/adapters"
)

const (
	DefaultDevDSN = "postgresql://asai-admin:asai-password@asai-db:5432/asai-db"
)

type Config struct {
	DB  *database.Database
	ENV string
	DSN string
}

func New() *Config {
	config := &Config{
		ENV: os.Getenv("ENVIRONMENT"),
	}

	return config
}

func (cnf *Config) InitDB() {
	if cnf.DSN == "" {
		fmt.Println("Empty DSN, setting default.")
		cnf.DSN = DefaultDevDSN
	}

	adapter := adapters.NewPostgres(
		database.WithDSN(cnf.DSN),
	)

	cnf.DB = database.New(adapter)
}

func (cnf *Config) LoadEnvironment() {
	fmt.Println("Current Environment:", cnf.ENV)

	if cnf.ENV == "LOCAL DEV" {
		cnf.setupLocalDev()
		return
	}
	if cnf.ENV == "HEROKU DEV" {
		cnf.setupHeroku()
		return
	}

	if cnf.ENV == "AWS DEV" {
		cnf.setupAWS()
		return
	}

	if cnf.ENV == "AWS PROD" {
		cnf.setupAWSProd()
		return
	}

	fmt.Println("Unknown Environment")
}

func (cnf *Config) setupAWSProd() {
	username := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	database := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")

	cnf.DSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, database)
}

func (cnf *Config) setupAWS() {
	username := os.Getenv("RDS_USERNAME")
	password := os.Getenv("RDS_PASSWORD")
	database := os.Getenv("RDS_DBNAME")
	host := os.Getenv("RDS_HOST")
	port := os.Getenv("RDS_PORT")

	cnf.DSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", username, password, host, port, database)
}

func (cnf *Config) setupHeroku() {
	cnf.DSN = os.Getenv("DATABASE_URL")
}

func (cnf *Config) setupLocalDev() {
	cnf.DSN = DefaultDevDSN
}

func (cnf *Config) RunServer() error {

	router := http.NewServeMux()
	wsManager := ws.NewManager(cnf.DB)
	router.HandleFunc("/ws/chat", wsManager.Handler)

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	return http.ListenAndServe(":"+port, router)
}
