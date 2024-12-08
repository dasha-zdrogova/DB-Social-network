package main

import (
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
	// Инициализация конфига
	cfg := config.NewConfig()

	// Инициализация БД
	db, err := postgres.NewPostgresDB(postgres.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		Username: cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  "disable",
	})
	if err != nil {
		log.Fatal("Failed to initialize db: ", err)
	}
	defer db.Close()

	// Инициализация репозиториев
	repos := postgres.NewRepositories(db)

	// Инициализация сервисов
	services := service.NewServices(repos)

	// Инициализация хендлеров
	handlers := handler.NewHandler(services)

	// Создание роутера и регистрация обработчиков
	router := mux.NewRouter()
	handlers.Register(router)

	// Запуск сервера
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, router); err != nil {
		log.Fatal(err)
	}
}
