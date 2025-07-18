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
	upstreams     map[string]*Upstream
	upstreamsM    sync.RWMutex
	domainRoutes  map[string]*DomainRoute
	domainRoutesM sync.RWMutex
}

type Upstream struct {
	Name      string
	URL       *url.URL
	Health    bool
	LastCheck time.Time
}

type DomainRoute struct {
	Domain     string
	AgentIP    string
	Routes     []Route
	SSLEnabled bool
}

type Route struct {
	Path          string
	ContainerName string
	Port          string
}

func NewReverseProxy() *ReverseProxy {
	router := mux.NewRouter()

	proxy := &ReverseProxy{
		router:       router,
		upstreams:    make(map[string]*Upstream),
		domainRoutes: make(map[string]*DomainRoute),
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

	// API маршруты проксируются на сервер
	rp.router.PathPrefix("/api/").HandlerFunc(rp.proxyToServer)

	// Динамические маршруты для доменов
	rp.router.HandleFunc("/", rp.handleDomainRequest)
}

func (rp *ReverseProxy) handleDomainRequest(w http.ResponseWriter, r *http.Request) {
	host := r.Host

	// Убираем порт из хоста если есть
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	rp.domainRoutesM.RLock()
	domainRoute, exists := rp.domainRoutes[host]
	rp.domainRoutesM.RUnlock()

	if !exists {
		// Если домен не найден, проксируем на приложение
		rp.proxyToApp(w, r)
		return
	}

	// Ищем подходящий маршрут
	var targetRoute *Route
	requestPath := r.URL.Path

	for _, route := range domainRoute.Routes {
		if strings.HasPrefix(requestPath, route.Path) {
			targetRoute = &route
			break
		}
	}

	if targetRoute == nil {
		// Если маршрут не найден, возвращаем 404
		http.Error(w, "Route not found", http.StatusNotFound)
		return
	}

	// Формируем URL для контейнера
	containerURL := fmt.Sprintf("http://%s:%s", targetRoute.ContainerName, targetRoute.Port)
	target, err := url.Parse(containerURL)
	if err != nil {
		log.Printf("Error parsing container URL: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Проксируем запрос к контейнеру
	rp.proxyRequest(w, r, target)
}

func (rp *ReverseProxy) proxyToServer(w http.ResponseWriter, r *http.Request) {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://server:8080"
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

// UpdateDomainRoutes обновляет маршруты доменов
func (rp *ReverseProxy) UpdateDomainRoutes(routes map[string]*DomainRoute) {
	rp.domainRoutesM.Lock()
	defer rp.domainRoutesM.Unlock()

	rp.domainRoutes = routes
	log.Printf("Updated domain routes: %d domains", len(routes))
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

func (rp *ReverseProxy) healthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rp.checkUpstreams()
		}
	}
}

func (rp *ReverseProxy) checkUpstreams() {
	rp.upstreamsM.Lock()
	defer rp.upstreamsM.Unlock()

	for name, upstream := range rp.upstreams {
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		resp, err := client.Get(upstream.URL.String() + "/health")
		if err != nil {
			log.Printf("Health check failed for %s: %v", name, err)
			upstream.Health = false
		} else {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				upstream.Health = true
			} else {
				upstream.Health = false
			}
		}
		upstream.LastCheck = time.Now()
	}
}

func main() {
	// Настраиваем логирование
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Получаем URL сервера из переменной окружения
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://server:8080"
	}

	// Создаем reverse proxy
	proxy := NewReverseProxy()

	// Создаем сервис синхронизации
	syncService := NewSyncService(serverURL, proxy)

	// Запускаем health check в фоне
	go proxy.healthCheck()

	// Запускаем синхронизацию доменов в фоне
	go syncService.StartSync()

	// Запускаем сервер
	if err := proxy.Start(); err != nil {
		log.Fatalf("Failed to start reverse proxy: %v", err)
	}
}
