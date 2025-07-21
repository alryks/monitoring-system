package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// SyncService представляет сервис синхронизации с сервером
type SyncService struct {
	serverURL string
	client    *http.Client
	proxy     *ReverseProxy
}

// NewSyncService создает новый сервис синхронизации
func NewSyncService(serverURL string, proxy *ReverseProxy) *SyncService {
	return &SyncService{
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		proxy: proxy,
	}
}

// StartSync запускает периодическую синхронизацию
func (s *SyncService) StartSync() {
	ticker := time.NewTicker(30 * time.Second) // Синхронизируем каждые 30 секунд
	defer ticker.Stop()

	log.Println("Starting domain sync service")

	for {
		select {
		case <-ticker.C:
			if err := s.syncDomains(); err != nil {
				log.Printf("Error syncing domains: %v", err)
			}
		}
	}
}

// syncDomains синхронизирует домены с сервером
func (s *SyncService) syncDomains() error {
	// Получаем все домены с сервера
	domains, err := s.fetchDomains()
	if err != nil {
		return fmt.Errorf("failed to fetch domains: %v", err)
	}

	// Преобразуем в простой маппинг домен -> IP агента
	domainAgents := make(map[string]string)

	for _, domain := range domains {
		domainAgents[domain.Name] = domain.AgentIP
	}

	// Обновляем маппинг в reverse-proxy
	s.proxy.UpdateDomainAgents(domainAgents)

	log.Printf("Synced %d domains", len(domainAgents))
	return nil
}

// fetchDomains получает домены с сервера
func (s *SyncService) fetchDomains() ([]DomainDetail, error) {
	url := fmt.Sprintf("%s/api/domains/public", s.serverURL)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var response struct {
		Domains []DomainDetail `json:"domains"`
		Total   int            `json:"total"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Domains, nil
}

// DomainDetail представляет детальную информацию о домене (упрощенная версия)
type DomainDetail struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	AgentIP    string `json:"agent_ip"`
	SSLEnabled bool   `json:"ssl_enabled"`
}
