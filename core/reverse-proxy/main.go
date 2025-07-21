package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type ReverseProxy struct {
	server        *http.Server
	router        *mux.Router
	domainAgents  map[string]string // domain -> agentIP
	domainAgentsM sync.RWMutex
}

func NewReverseProxy() *ReverseProxy {
	router := mux.NewRouter()

	proxy := &ReverseProxy{
		router:       router,
		domainAgents: make(map[string]string),
	}

	// Настраиваем маршруты
	proxy.setupRoutes()

	// Создаем HTTP сервер
	proxy.server = &http.Server{
		Addr:         ":80",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return proxy
}

func (rp *ReverseProxy) setupRoutes() {
	// Health check endpoint
	rp.router.HandleFunc("/health", rp.healthHandler).Methods("GET")

	// Динамические маршруты для доменов
	rp.router.HandleFunc("/{path:.*}", rp.handleDomainRequest)
}

func (rp *ReverseProxy) handleDomainRequest(w http.ResponseWriter, r *http.Request) {
	host := r.Host

	// Убираем порт из хоста если есть
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	// Проверяем, является ли это доменом приложения
	appDomain := os.Getenv("APP_DOMAIN")
	if appDomain != "" && (host == appDomain || host == "localhost" || host == "127.0.0.1") {
		// Если это домен приложения, проверяем, есть является ли это запрос на API
		if strings.HasPrefix(r.URL.Path, "/api") {
			rp.proxyToServer(w, r)
			return
		}

		// Если это домен приложения, проксируем на приложение
		rp.proxyToApp(w, r)
		return
	}

	// Получаем IP агента для домена
	rp.domainAgentsM.RLock()
	agentIP, exists := rp.domainAgents[host]
	rp.domainAgentsM.RUnlock()

	if !exists {
		// Если домен не найден, возвращаем 404
		http.Error(w, "Domain not found", http.StatusNotFound)
		return
	}

	// Проксируем запрос на IP агента с измененным Host
	rp.proxyToAgent(w, r, agentIP, host)
}

func (rp *ReverseProxy) proxyToAgent(w http.ResponseWriter, r *http.Request, agentIP, originalHost string) {
	// Формируем URL для агента (используем порт 80 по умолчанию)
	agentURL := fmt.Sprintf("http://%s:80", agentIP)
	target, err := url.Parse(agentURL)
	if err != nil {
		log.Printf("Error parsing agent URL: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Создаем прокси
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Настраиваем директор для модификации запросов
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Устанавливаем оригинальный Host заголовок
		req.Host = originalHost
		req.Header.Set("Host", originalHost)

		// Устанавливаем дополнительные заголовки
		req.Header.Set("X-Forwarded-Host", originalHost)
		req.Header.Set("X-Forwarded-Proto", "http")
		if r.Header.Get("X-Real-IP") != "" {
			req.Header.Set("X-Real-IP", r.Header.Get("X-Real-IP"))
		} else {
			req.Header.Set("X-Real-IP", r.RemoteAddr)
		}
	}

	// Настраиваем модификатор ответов
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Добавляем заголовки безопасности
		resp.Header.Set("X-Content-Type-Options", "nosniff")
		resp.Header.Set("X-Frame-Options", "DENY")
		resp.Header.Set("X-XSS-Protection", "1; mode=block")
		return nil
	}

	// Настраиваем обработчик ошибок
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error to agent %s: %v", agentIP, err)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}

	// Выполняем проксирование
	proxy.ServeHTTP(w, r)
}

func (rp *ReverseProxy) proxyToServer(w http.ResponseWriter, r *http.Request) {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://server:8000"
	}

	target, err := url.Parse(serverURL)
	if err != nil {
		log.Printf("Error parsing server URL: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rp.proxyRequest(w, r, target)
}

func (rp *ReverseProxy) proxyToApp(w http.ResponseWriter, r *http.Request) {
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://app:3000"
	}

	target, err := url.Parse(appURL)
	if err != nil {
		log.Printf("Error parsing app URL: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rp.proxyRequest(w, r, target)
}

func (rp *ReverseProxy) proxyRequest(w http.ResponseWriter, r *http.Request, target *url.URL) {
	// Создаем прокси
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Настраиваем директор для модификации запросов
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Устанавливаем заголовки
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Forwarded-Proto", "http")
		if r.Header.Get("X-Real-IP") != "" {
			req.Header.Set("X-Real-IP", r.Header.Get("X-Real-IP"))
		} else {
			req.Header.Set("X-Real-IP", r.RemoteAddr)
		}
	}

	// Настраиваем модификатор ответов
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Добавляем заголовки безопасности
		resp.Header.Set("X-Content-Type-Options", "nosniff")
		resp.Header.Set("X-Frame-Options", "DENY")
		resp.Header.Set("X-XSS-Protection", "1; mode=block")
		return nil
	}

	// Настраиваем обработчик ошибок
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error: %v", err)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}

	// Выполняем проксирование
	proxy.ServeHTTP(w, r)
}

func (rp *ReverseProxy) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateDomainAgents обновляет маппинг доменов на IP агентов
func (rp *ReverseProxy) UpdateDomainAgents(domainAgents map[string]string) {
	rp.domainAgentsM.Lock()
	defer rp.domainAgentsM.Unlock()

	rp.domainAgents = domainAgents
	log.Printf("Updated domain agents: %d domains", len(domainAgents))
}

func (rp *ReverseProxy) Start() error {
	log.Printf("Starting reverse proxy on :80")

	// Запускаем сервер в горутине
	go func() {
		if err := rp.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Ожидаем сигнал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down reverse proxy...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := rp.server.Shutdown(ctx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
		return err
	}

	log.Println("Reverse proxy stopped")
	return nil
}

func main() {
	// Настраиваем логирование
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Получаем URL сервера из переменной окружения
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://server:8000"
	}

	// Создаем reverse proxy
	proxy := NewReverseProxy()

	// Создаем сервис синхронизации
	syncService := NewSyncService(serverURL, proxy)

	// Запускаем синхронизацию доменов в фоне
	go syncService.StartSync()

	// Запускаем сервер
	if err := proxy.Start(); err != nil {
		log.Fatalf("Failed to start reverse proxy: %v", err)
	}
}
