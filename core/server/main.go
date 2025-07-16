// Package main Docker Monitoring System API
//
//	@title						Docker Monitoring System API
//	@version					1.0
//	@description				API для системы мониторинга Docker контейнеров
//	@termsOfService				http://swagger.io/terms/
//	@contact.name				API Support
//	@contact.url				http://www.swagger.io/support
//	@contact.email				support@swagger.io
//	@license.name				MIT
//	@license.url				https://opensource.org/licenses/MIT
//	@host						localhost:8000
//	@BasePath					/api
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "monitoring-system/core/server/docs" // Swagger документация
	"monitoring-system/core/server/internal/auth"
	"monitoring-system/core/server/internal/config"
	"monitoring-system/core/server/internal/database"
	"monitoring-system/core/server/internal/handlers"
)

func main() {
	// Инициализируем конфигурацию
	cfg := config.Load()

	// Подключаемся к базе данных
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Инициализируем сервисы
	authService := auth.NewService(cfg.JWTSecret)

	// Инициализируем обработчики
	h := handlers.New(db, authService)

	// Настраиваем роутер
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API маршруты
	r.Route("/api", func(r chi.Router) {
		// Swagger UI
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("http://localhost/api/swagger/doc.json"),
		))

		// Публичные маршруты
		r.Post("/login", h.Login)

		// Маршруты для агентов (с Bearer токеном)
		r.Post("/agent/ping", h.AgentPing)

		// Защищенные маршруты (требуют JWT аутентификации)
		r.Group(func(r chi.Router) {
			r.Use(authService.JWTMiddleware)

			// Управление агентами
			r.Get("/agents", h.GetAgents)
			r.Post("/agents", h.CreateAgent)
			r.Put("/agents/{id}", h.UpdateAgent)
			r.Delete("/agents/{id}", h.DeleteAgent)
			r.Get("/agents/{id}", h.GetAgentDetail)

			// Метрики и мониторинг
			r.Get("/agents/{id}/metrics", h.GetAgentMetrics)
			r.Get("/agents/{id}/containers", h.GetAgentContainers)
			r.Get("/dashboard", h.GetDashboardData)

			// Контейнеры
			r.Get("/containers", h.GetContainers)
			r.Get("/containers/{id}", h.GetContainerDetail)
			r.Get("/containers/{id}/logs", h.GetContainerLogs)

			// Образы
			r.Get("/images", h.GetImages)

			// Тома
			r.Get("/volumes", h.GetVolumes)

			// Сети
			r.Get("/networks", h.GetNetworks)
		})
	})

	// Запускаем сервер
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
