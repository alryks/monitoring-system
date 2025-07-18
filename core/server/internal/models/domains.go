package models

import (
	"time"

	"github.com/google/uuid"
)

// Domain представляет домен в системе
type Domain struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`               // например: "dashboard.domain.net"
	AgentID    uuid.UUID `json:"agent_id" db:"agent_id"`       // ID агента, на котором размещен домен
	AgentIP    string    `json:"agent_ip" db:"agent_ip"`       // IP адрес агента
	IsActive   bool      `json:"is_active" db:"is_active"`     // активен ли домен
	SSLEnabled bool      `json:"ssl_enabled" db:"ssl_enabled"` // включен ли SSL
	Created    time.Time `json:"created" db:"created"`
	Updated    time.Time `json:"updated" db:"updated"`
	// Дополнительные поля для совместимости с frontend
	AgentName *string `json:"agent_name"`
}

// DomainRoute представляет маршрут домена к контейнеру
type DomainRoute struct {
	ID            uuid.UUID `json:"id" db:"id"`
	DomainID      uuid.UUID `json:"domain_id" db:"domain_id"`
	ContainerName string    `json:"container_name" db:"container_name"` // имя контейнера
	Port          string    `json:"port" db:"port"`                     // порт контейнера
	Path          string    `json:"path" db:"path"`                     // путь (опционально, для /api/*)
	IsActive      bool      `json:"is_active" db:"is_active"`
	Created       time.Time `json:"created" db:"created"`
	Updated       time.Time `json:"updated" db:"updated"`
}

// DomainDetail представляет детальную информацию о домене
type DomainDetail struct {
	Domain
	Agent  Agent         `json:"agent"`
	Routes []DomainRoute `json:"routes"`
}

// CreateDomainRequest представляет запрос на создание домена
type CreateDomainRequest struct {
	Name       string    `json:"name" example:"dashboard.domain.net"`
	AgentID    uuid.UUID `json:"agent_id"`
	SSLEnabled bool      `json:"ssl_enabled" example:"false"`
}

// UpdateDomainRequest представляет запрос на обновление домена
type UpdateDomainRequest struct {
	Name       *string    `json:"name,omitempty"`
	AgentID    *uuid.UUID `json:"agent_id,omitempty"`
	IsActive   *bool      `json:"is_active,omitempty"`
	SSLEnabled *bool      `json:"ssl_enabled,omitempty"`
}

// CreateDomainRouteRequest представляет запрос на создание маршрута
type CreateDomainRouteRequest struct {
	DomainID      uuid.UUID `json:"domain_id"`
	ContainerName string    `json:"container_name" example:"my-container"`
	Port          string    `json:"port" example:"3000"`
	Path          string    `json:"path,omitempty" example:"/api"`
}

// UpdateDomainRouteRequest представляет запрос на обновление маршрута
type UpdateDomainRouteRequest struct {
	ContainerName *string `json:"container_name,omitempty"`
	Port          *string `json:"port,omitempty"`
	Path          *string `json:"path,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

// DomainListResponse представляет ответ со списком доменов
type DomainListResponse struct {
	Domains []DomainDetail `json:"domains"`
	Total   int            `json:"total"`
}

// DomainRouteListResponse представляет ответ со списком маршрутов
type DomainRouteListResponse struct {
	Routes []DomainRoute `json:"routes"`
	Total  int           `json:"total"`
}

// NginxConfig представляет конфигурацию nginx для домена
type NginxConfig struct {
	Domain     string          `json:"domain"`
	AgentIP    string          `json:"agent_ip"`
	Routes     []NginxRoute    `json:"routes"`
	SSLEnabled bool            `json:"ssl_enabled"`
	SSLCert    *SSLCertificate `json:"ssl_cert,omitempty"`
}

// NginxRoute представляет маршрут в конфигурации nginx
type NginxRoute struct {
	Path          string `json:"path"`           // например: "/" или "/api"
	ContainerName string `json:"container_name"` // имя контейнера
	Port          string `json:"port"`           // порт контейнера
}

// SSLCertificate представляет SSL сертификат
type SSLCertificate struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

// AgentNginxConfig представляет конфигурацию nginx для агента
type AgentNginxConfig struct {
	AgentID uuid.UUID     `json:"agent_id"`
	Domains []NginxConfig `json:"domains"`
}

// DomainStatus представляет статус домена
type DomainStatus struct {
	DomainID   uuid.UUID     `json:"domain_id"`
	DomainName string        `json:"domain_name"`
	AgentID    uuid.UUID     `json:"agent_id"`
	AgentName  string        `json:"agent_name"`
	AgentIP    string        `json:"agent_ip"`
	IsActive   bool          `json:"is_active"`
	SSLEnabled bool          `json:"ssl_enabled"`
	Routes     []RouteStatus `json:"routes"`
}

// RouteStatus представляет статус маршрута
type RouteStatus struct {
	RouteID         uuid.UUID `json:"route_id"`
	ContainerName   string    `json:"container_name"`
	Port            string    `json:"port"`
	Path            string    `json:"path"`
	IsActive        bool      `json:"is_active"`
	ContainerStatus string    `json:"container_status"` // running, stopped, etc.
}
