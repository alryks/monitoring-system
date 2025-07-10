package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Post("/agent", handleAgentData)
	})

	log.Println("Server started on :8000")
	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func handleAgentData(w http.ResponseWriter, r *http.Request) {
	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// // Проверяем токен авторизации
	// authHeader := r.Header.Get("Authorization")
	// if authHeader == "" {
	// 	log.Println("Missing Authorization header")
	// 	http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
	// 	return
	// }

	// // Простая проверка токена (Bearer token)
	// expectedToken := "Bearer my-token"
	// if authHeader != expectedToken {
	// 	log.Printf("Invalid token: %s", authHeader)
	// 	http.Error(w, "Invalid token", http.StatusUnauthorized)
	// 	return
	// }

	// Проверяем, что это валидный JSON
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		log.Printf("Invalid JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Форматируем JSON для красивого вывода
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		fmt.Println(string(body))
	} else {
		fmt.Println(prettyJSON.String())
	}

	fmt.Println("-------------")
	fmt.Println()

	// Отправляем ответ агенту
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok", "message": "Data received successfully"}`))
}
