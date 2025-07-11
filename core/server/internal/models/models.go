package models

import (
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя системы
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Email        *string    `json:"email" db:"email"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	Role         string     `json:"role" db:"role"`
	Created      time.Time  `json:"created" db:"created"`
	LastLogin    *time.Time `json:"last_login" db:"last_login"`
}

// Agent представляет агент мониторинга
type Agent struct {
	ID       uuid.UUID  `json:"id" db:"id"`
	Name     string     `json:"name" db:"name"`
	Token    string     `json:"token" db:"token"`
	IsActive bool       `json:"is_active" db:"is_active"`
	Created  time.Time  `json:"created" db:"created"`
	LastPing *time.Time `json:"last_ping" db:"last_ping"`
	PublicIP *string    `json:"public_ip,omitempty"`
	Status   string     `json:"status"` // online, offline, unknown
}

// AgentPing представляет пинг от агента
type AgentPing struct {
	ID      uuid.UUID `json:"id" db:"id"`
	AgentID uuid.UUID `json:"agent_id" db:"agent_id"`
	Created time.Time `json:"created" db:"created"`
}

// CPUMetric представляет метрику CPU
type CPUMetric struct {
	ID           uuid.UUID `json:"id" db:"id"`
	PingID       uuid.UUID `json:"ping_id" db:"ping_id"`
	CPUName      string    `json:"cpu_name" db:"cpu_name"`
	UsagePercent float64   `json:"usage_percent" db:"usage_percent"`
}

// MemoryMetric представляет метрику памяти
type MemoryMetric struct {
	ID          uuid.UUID `json:"id" db:"id"`
	PingID      uuid.UUID `json:"ping_id" db:"ping_id"`
	RAMTotalMB  int64     `json:"ram_total_mb" db:"ram_total_mb"`
	RAMUsageMB  int64     `json:"ram_usage_mb" db:"ram_usage_mb"`
	SwapTotalMB int64     `json:"swap_total_mb" db:"swap_total_mb"`
	SwapUsageMB int64     `json:"swap_usage_mb" db:"swap_usage_mb"`
}

// DiskMetric представляет метрику диска
type DiskMetric struct {
	ID         uuid.UUID `json:"id" db:"id"`
	PingID     uuid.UUID `json:"ping_id" db:"ping_id"`
	DiskName   string    `json:"disk_name" db:"disk_name"`
	ReadBytes  int64     `json:"read_bytes" db:"read_bytes"`
	WriteBytes int64     `json:"write_bytes" db:"write_bytes"`
	Reads      int64     `json:"reads" db:"reads"`
	Writes     int64     `json:"writes" db:"writes"`
}

// NetworkMetric представляет метрику сети
type NetworkMetric struct {
	ID            uuid.UUID `json:"id" db:"id"`
	PingID        uuid.UUID `json:"ping_id" db:"ping_id"`
	PublicIP      string    `json:"public_ip" db:"public_ip"`
	SentBytes     int64     `json:"sent_bytes" db:"sent_bytes"`
	ReceivedBytes int64     `json:"received_bytes" db:"received_bytes"`
}

// Container представляет контейнер Docker
type Container struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	PingID               uuid.UUID `json:"ping_id" db:"ping_id"`
	ContainerID          string    `json:"container_id" db:"container_id"`
	Name                 string    `json:"name" db:"name"`
	ImageID              string    `json:"image_id" db:"image_id"`
	Status               string    `json:"status" db:"status"`
	RestartCount         int       `json:"restart_count" db:"restart_count"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	IPAddress            *string   `json:"ip_address" db:"ip_address"`
	MACAddress           *string   `json:"mac_address" db:"mac_address"`
	CPUUsagePercent      *float64  `json:"cpu_usage_percent" db:"cpu_usage_percent"`
	MemoryUsageMB        *int64    `json:"memory_usage_mb" db:"memory_usage_mb"`
	NetworkSentBytes     *int64    `json:"network_sent_bytes" db:"network_sent_bytes"`
	NetworkReceivedBytes *int64    `json:"network_received_bytes" db:"network_received_bytes"`
}

// AgentData представляет данные от агента (соответствует JSON от агента)
type AgentData struct {
	Metrics Metrics    `json:"metrics"`
	Docker  DockerInfo `json:"docker"`
}

type Metrics struct {
	CPU     []CPUInfo   `json:"cpu"`
	Memory  MemoryInfo  `json:"memory"`
	Disk    []DiskInfo  `json:"disk"`
	Network NetworkInfo `json:"network"`
}

type CPUInfo struct {
	Name  string  `json:"name"`
	Usage float64 `json:"usage"`
}

type MemoryInfo struct {
	RAM  RAMInfo  `json:"ram"`
	Swap SwapInfo `json:"swap"`
}

type RAMInfo struct {
	Total uint64 `json:"total"`
	Usage uint64 `json:"usage"`
}

type SwapInfo struct {
	Total uint64 `json:"total"`
	Usage uint64 `json:"usage"`
}

type DiskInfo struct {
	Name       string `json:"name"`
	ReadBytes  uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
	Reads      uint64 `json:"reads"`
	Writes     uint64 `json:"writes"`
}

type NetworkInfo struct {
	PublicIP string `json:"public_ip"`
	Sent     uint64 `json:"sent"`
	Received uint64 `json:"received"`
}

type DockerInfo struct {
	Containers []ContainerInfo `json:"containers"`
	Images     []ImageInfo     `json:"images"`
	Volumes    []VolumeInfo    `json:"volumes"`
	Networks   []NetworkDocker `json:"networks"`
}

type ContainerInfo struct {
	ID           string               `json:"id"`
	Created      string               `json:"created"`
	Status       string               `json:"status"`
	RestartCount int                  `json:"restart_count"`
	Image        string               `json:"image"`
	Name         string               `json:"name"`
	IP           *string              `json:"ip"`
	MAC          *string              `json:"mac"`
	CPU          *float64             `json:"cpu"`
	Memory       *uint64              `json:"memory"`
	Network      ContainerNetworkInfo `json:"network"`
	Volumes      []string             `json:"volumes"`
	Logs         []string             `json:"logs"`
}

type ContainerNetworkInfo struct {
	Sent     *uint64  `json:"sent"`
	Received *uint64  `json:"received"`
	Networks []string `json:"networks"`
}

type ImageInfo struct {
	ID           string   `json:"id"`
	Created      string   `json:"created"`
	Size         int64    `json:"size"`
	Tags         []string `json:"tags"`
	Architecture string   `json:"architecture"`
}

type VolumeInfo struct {
	Name       string `json:"name"`
	Created    string `json:"created"`
	Driver     string `json:"driver"`
	Mountpoint string `json:"mountpoint"`
}

type NetworkDocker struct {
	ID      string  `json:"id"`
	Created string  `json:"created"`
	Name    string  `json:"name"`
	Driver  string  `json:"driver"`
	Scope   string  `json:"scope"`
	Subnet  *string `json:"subnet"`
	Gateway *string `json:"gateway"`
}

// LoginRequest представляет запрос на вход
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse представляет ответ на вход
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// CreateAgentRequest представляет запрос на создание агента
type CreateAgentRequest struct {
	Name string `json:"name"`
}

// DashboardData представляет данные для дашборда
type DashboardData struct {
	Agents         []Agent        `json:"agents"`
	TotalAgents    int            `json:"total_agents"`
	OnlineAgents   int            `json:"online_agents"`
	OfflineAgents  int            `json:"offline_agents"`
	RecentMetrics  []RecentMetric `json:"recent_metrics"`
	SystemOverview SystemOverview `json:"system_overview"`
}

type RecentMetric struct {
	AgentID   uuid.UUID `json:"agent_id"`
	AgentName string    `json:"agent_name"`
	Timestamp time.Time `json:"timestamp"`
	CPUUsage  float64   `json:"cpu_usage"`
	RAMUsage  float64   `json:"ram_usage"`
	PublicIP  string    `json:"public_ip"`
}

type SystemOverview struct {
	TotalCPUCores     int   `json:"total_cpu_cores"`
	TotalRAMMB        int64 `json:"total_ram_mb"`
	TotalContainers   int   `json:"total_containers"`
	RunningContainers int   `json:"running_containers"`
}
