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
	// Дополнительные поля для совместимости с frontend
	AgentID   *uuid.UUID `json:"agent_id"`
	AgentName *string    `json:"agent_name"`
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
// @Description Запрос на аутентификацию пользователя
type LoginRequest struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"admin123"`
}

// LoginResponse представляет ответ на вход
// @Description Ответ с JWT токеном и информацией о пользователе
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  User   `json:"user"`
}

// CreateAgentRequest представляет запрос на создание агента
// @Description Запрос на создание нового агента мониторинга
type CreateAgentRequest struct {
	Name string `json:"name" example:"Production Server 1"`
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

// Image представляет Docker образ в БД
type Image struct {
	ID           uuid.UUID `json:"id" db:"id"`
	PingID       uuid.UUID `json:"ping_id" db:"ping_id"`
	ImageID      string    `json:"image_id" db:"image_id"`
	Created      time.Time `json:"created" db:"created"`
	Size         int64     `json:"size" db:"size"`
	Architecture string    `json:"architecture" db:"architecture"`
}

// ImageTag представляет тег Docker образа
type ImageTag struct {
	ID      uuid.UUID `json:"id" db:"id"`
	ImageID uuid.UUID `json:"image_id" db:"image_id"`
	Tag     string    `json:"tag" db:"tag"`
}

// Volume представляет Docker том
type Volume struct {
	ID         uuid.UUID `json:"id" db:"id"`
	PingID     uuid.UUID `json:"ping_id" db:"ping_id"`
	Name       string    `json:"name" db:"name"`
	Created    time.Time `json:"created" db:"created"`
	Driver     string    `json:"driver" db:"driver"`
	Mountpoint string    `json:"mountpoint" db:"mountpoint"`
}

// Network представляет Docker сеть в БД
type Network struct {
	ID      uuid.UUID `json:"id" db:"id"`
	PingID  uuid.UUID `json:"ping_id" db:"ping_id"`
	NetID   string    `json:"network_id" db:"network_id"`
	Created time.Time `json:"created" db:"created"`
	Name    string    `json:"name" db:"name"`
	Driver  string    `json:"driver" db:"driver"`
	Scope   string    `json:"scope" db:"scope"`
	Subnet  *string   `json:"subnet" db:"subnet"`
	Gateway *string   `json:"gateway" db:"gateway"`
}

// ContainerVolume представляет связь контейнера с томом
type ContainerVolume struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ContainerID uuid.UUID `json:"container_id" db:"container_id"`
	VolumeName  string    `json:"volume_name" db:"volume_name"`
}

// ContainerNetwork представляет связь контейнера с сетью
type ContainerNetwork struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ContainerID uuid.UUID `json:"container_id" db:"container_id"`
	NetworkName string    `json:"network_name" db:"network_name"`
}

// ContainerLog представляет лог контейнера
type ContainerLog struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ContainerID uuid.UUID `json:"container_id" db:"container_id"`
	LogLine     string    `json:"log_line" db:"log_line"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
}

// Расширенные модели для API ответов

// ContainerDetail представляет детальную информацию о контейнере
type ContainerDetail struct {
	Container
	Agent    Agent             `json:"agent"`
	Volumes  []string          `json:"volumes"`
	Networks []string          `json:"networks"`
	Logs     []ContainerLog    `json:"logs"`
	History  []ContainerMetric `json:"history"`
}

// ContainerMetric представляет историю метрик контейнера
type ContainerMetric struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUUsage    *float64  `json:"cpu_usage"`
	MemoryUsage *int64    `json:"memory_usage"`
}

// ImageDetail представляет детальную информацию об образе
type ImageDetail struct {
	Image
	Tags  []string `json:"tags"`
	Agent Agent    `json:"agent"`
}

// VolumeDetail представляет детальную информацию о томе
type VolumeDetail struct {
	Volume
	Agent Agent `json:"agent"`
}

// NetworkDetail представляет детальную информацию о сети
type NetworkDetail struct {
	Network
	Agent Agent `json:"agent"`
}

// AgentDetail представляет детальную информацию об агенте
type AgentDetail struct {
	Agent
	Metrics       AgentMetrics      `json:"metrics"`
	Containers    []ContainerDetail `json:"containers"`
	Images        []ImageDetail     `json:"images"`
	Volumes       []VolumeDetail    `json:"volumes"`
	Networks      []NetworkDetail   `json:"networks"`
	SystemMetrics []SystemMetric    `json:"system_metrics"`
}

// AgentMetrics представляет текущие метрики агента
type AgentMetrics struct {
	CPU     []CPUMetricCurrent   `json:"cpu"`
	Memory  MemoryMetricCurrent  `json:"memory"`
	Disk    []DiskMetricCurrent  `json:"disk"`
	Network NetworkMetricCurrent `json:"network"`
}

// CPUMetricCurrent представляет текущую метрику CPU
type CPUMetricCurrent struct {
	Name  string  `json:"name"`
	Usage float64 `json:"usage"`
}

// MemoryMetricCurrent представляет текущую метрику памяти
type MemoryMetricCurrent struct {
	RAMTotal    int64   `json:"ram_total"`
	RAMUsage    int64   `json:"ram_usage"`
	RAMPercent  float64 `json:"ram_percent"`
	SwapTotal   int64   `json:"swap_total"`
	SwapUsage   int64   `json:"swap_usage"`
	SwapPercent float64 `json:"swap_percent"`
}

// DiskMetricCurrent представляет текущую метрику диска
type DiskMetricCurrent struct {
	Name       string `json:"name"`
	ReadBytes  int64  `json:"read_bytes"`
	WriteBytes int64  `json:"write_bytes"`
	ReadSpeed  int64  `json:"read_speed"`
	WriteSpeed int64  `json:"write_speed"`
}

// NetworkMetricCurrent представляет текущую метрику сети
type NetworkMetricCurrent struct {
	PublicIP      string `json:"public_ip"`
	SentBytes     int64  `json:"sent_bytes"`
	ReceivedBytes int64  `json:"received_bytes"`
	SentSpeed     int64  `json:"sent_speed"`
	ReceivedSpeed int64  `json:"received_speed"`
}

// SystemMetric представляет историческую метрику системы
type SystemMetric struct {
	Timestamp time.Time `json:"timestamp"`
	CPUUsage  float64   `json:"cpu_usage"`
	RAMUsage  float64   `json:"ram_usage"`
	PublicIP  string    `json:"public_ip"`
}

// ContainerListResponse представляет ответ со списком контейнеров
type ContainerListResponse struct {
	Containers []ContainerDetail `json:"containers"`
	Total      int               `json:"total"`
}

// ImageListResponse представляет ответ со списком образов
type ImageListResponse struct {
	Images []ImageDetail `json:"images"`
	Total  int           `json:"total"`
}

// VolumeListResponse представляет ответ со списком томов
type VolumeListResponse struct {
	Volumes []VolumeDetail `json:"volumes"`
	Total   int            `json:"total"`
}

// NetworkListResponse представляет ответ со списком сетей
type NetworkListResponse struct {
	Networks []NetworkDetail `json:"networks"`
	Total    int             `json:"total"`
}

// TopContainer представляет контейнер в топе по ресурсам
type TopContainer struct {
	Name        string  `json:"name"`
	AgentName   string  `json:"agent_name"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage int64   `json:"memory_usage"`
	Status      string  `json:"status"`
}

// DashboardMetrics представляет расширенные метрики для дашборда
type DashboardMetrics struct {
	TotalContainers   int            `json:"total_containers"`
	RunningContainers int            `json:"running_containers"`
	StoppedContainers int            `json:"stopped_containers"`
	AverageCPUUsage   float64        `json:"average_cpu_usage"`
	AverageRAMUsage   float64        `json:"average_ram_usage"`
	NetworkTraffic    NetworkTraffic `json:"network_traffic"`
	TopContainersCPU  []TopContainer `json:"top_containers_cpu"`
	TopContainersRAM  []TopContainer `json:"top_containers_ram"`
}

// NetworkTraffic представляет сетевой трафик
type NetworkTraffic struct {
	SentSpeed     int64 `json:"sent_speed"`
	ReceivedSpeed int64 `json:"received_speed"`
}

// KPIMetrics представляет основные KPI метрики
type KPIMetrics struct {
	AgentsOnline      int     `json:"agents_online"`
	AgentsTotal       int     `json:"agents_total"`
	ContainersRunning int     `json:"containers_running"`
	ContainersStopped int     `json:"containers_stopped"`
	ContainersTotal   int     `json:"containers_total"`
	AvgCPUUsage       float64 `json:"avg_cpu_usage"`
	AvgMemoryUsage    float64 `json:"avg_memory_usage"`
}

// ResourceUsagePoint представляет точку использования ресурсов
type ResourceUsagePoint struct {
	Timestamp time.Time `json:"timestamp"`
	CPU       float64   `json:"cpu"`
	Memory    float64   `json:"memory"`
}

// NetworkActivityPoint представляет точку сетевой активности
type NetworkActivityPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Sent      int64     `json:"sent"`
	Received  int64     `json:"received"`
}

// AgentSummary представляет сводку по агенту
type AgentSummary struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	LastPing    *time.Time `json:"last_ping"`
	Containers  int        `json:"containers"`
	CPUUsage    float64    `json:"cpu_usage"`
	MemoryUsage float64    `json:"memory_usage"`
}

// DashboardExtended представляет расширенные данные дашборда
type DashboardExtended struct {
	Kpis                KPIMetrics             `json:"kpis"`
	ResourceUsage       []ResourceUsagePoint   `json:"resource_usage"`
	NetworkActivity     []NetworkActivityPoint `json:"network_activity"`
	TopContainersCPU    []TopContainer         `json:"top_containers_cpu"`
	TopContainersMemory []TopContainer         `json:"top_containers_memory"`
	AgentsSummary       []AgentSummary         `json:"agents_summary"`
}

// TimeRange представляет временной диапазон
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// FilterOptions представляет опции фильтрации
type FilterOptions struct {
	AgentID   *uuid.UUID `json:"agent_id"`
	TimeRange *TimeRange `json:"time_range"`
	Status    *string    `json:"status"`
	Search    *string    `json:"search"`
}

// PaginationOptions представляет опции пагинации
type PaginationOptions struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// ListRequest представляет запрос на получение списка с фильтрацией
type ListRequest struct {
	Filter     FilterOptions     `json:"filter"`
	Pagination PaginationOptions `json:"pagination"`
}

// UpdateAgentRequest представляет запрос на обновление агента
type UpdateAgentRequest struct {
	Name *string `json:"name"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse представляет успешный ответ
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
