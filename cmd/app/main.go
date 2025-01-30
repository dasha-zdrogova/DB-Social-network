package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"social-network/internal/config"
	"social-network/internal/handler"
	"social-network/internal/repository/postgres"
	"social-network/internal/service"
)

func main() {
	cfg := config.NewConfig()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	repos := postgres.NewRepositories(db)

	services := service.NewServices(repos)

	handlers := handler.NewHandler(services)

	router := mux.NewRouter()
	handlers.Register(router)

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, router); err != nil {
		log.Fatal(err)
	}
}
