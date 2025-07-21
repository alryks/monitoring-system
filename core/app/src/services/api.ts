import axios from 'axios'

export const api = axios.create({
  baseURL: '',
  timeout: 10000,
})

// Добавляем токен к каждому запросу, если он есть
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Обрабатываем ответы и ошибки
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Автоматически выходим из системы при ошибке аутентификации
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Типы данных

// Базовые типы
export interface Agent {
  id: string
  name: string
  token: string
  is_active: boolean
  created: string
  last_ping?: string
  public_ip?: string
  status: 'online' | 'offline' | 'unknown'
}

export interface Container {
  id: string
  ping_id: string
  container_id: string
  name: string
  image_id: string
  status: string
  restart_count: number
  created_at: string
  ip_address?: string
  mac_address?: string
  cpu_usage_percent?: number
  memory_usage_mb?: number
  network_sent_bytes?: number
  network_received_bytes?: number
  agent_id?: string
  agent_name?: string
}

export interface ContainerDetail extends Container {
  agent: Agent
  logs: ContainerLog[]
  history: ContainerMetric[]
}

export interface ContainerLog {
  id: string
  container_id: string
  log_line: string
  timestamp: string
}

export interface ContainerMetric {
  timestamp: string
  cpu_usage?: number
  memory_usage?: number
}

// Response interfaces для API
export interface ContainerListResponse {
  containers: Container[]
  total: number
}

export interface ImageListResponse {
  images: Image[]
  total: number
}



export interface Image {
  id: string
  ping_id: string
  image_id: string
  created: string
  size: number
  architecture: string
  tags: string[]
  agent_id?: string
  agent_name?: string
}

export interface ImageDetail extends Image {
  agent: Agent
}



// Dashboard и агенты
export interface DashboardData {
  kpis: KPIMetrics
  resource_usage: ResourceUsagePoint[]
  network_activity: NetworkActivityPoint[]
  top_containers_cpu: TopContainer[]
  top_containers_memory: TopContainer[]
  agents_summary: AgentSummary[]
}

export interface KPIMetrics {
  agents_online: number
  agents_total: number
  containers_running: number
  containers_stopped: number
  containers_total: number
  avg_cpu_usage: number
  avg_memory_usage: number
}

export interface ResourceUsagePoint {
  timestamp: string
  cpu: number
  memory: number
}

export interface NetworkActivityPoint {
  timestamp: string
  sent: number
  received: number
}

export interface TopContainer {
  name: string
  agent_name: string
  cpu_usage?: number
  memory_usage: number
  status: string
}

export interface AgentSummary {
  id: string
  name: string
  status: string
  last_ping: string | null
  containers: number
  cpu_usage: number
  memory_usage: number
}

export interface RecentMetric {
  agent_id: string
  agent_name: string
  timestamp: string
  cpu_usage: number
  ram_usage: number
  public_ip: string
}

export interface SystemOverview {
  total_cpu_cores: number
  total_ram_mb: number
  total_containers: number
  running_containers: number
}

export interface MetricPoint {
  timestamp: string
  public_ip: string
  ram_total_mb: number
  ram_usage_mb: number
  swap_total_mb: number
  swap_usage_mb: number
}

// Детальная информация об агенте
export interface AgentDetail {
  id: string
  name: string
  token: string
  is_active: boolean
  created: string
  last_ping?: string
  public_ip?: string
  status: 'online' | 'offline' | 'unknown'
  metrics: AgentMetrics
  containers: ContainerDetail[]
  images: ImageDetail[]
  system_metrics: SystemMetric[]
}

export interface AgentMetrics {
  cpu: CPUMetricCurrent[]
  memory: MemoryMetricCurrent
  disk: DiskMetricCurrent[]
  network: NetworkMetricCurrent
}

export interface CPUMetricCurrent {
  name: string
  usage: number
}

export interface MemoryMetricCurrent {
  ram_total: number
  ram_usage: number
  ram_percent: number
  swap_total: number
  swap_usage: number
  swap_percent: number
}

export interface DiskMetricCurrent {
  name: string
  read_bytes: number
  write_bytes: number
  read_speed: number
  write_speed: number
}

export interface NetworkMetricCurrent {
  public_ip: string
  sent_bytes: number
  received_bytes: number
  sent_speed: number
  received_speed: number
}

export interface SystemMetric {
  timestamp: string
  cpu_usage: number
  ram_usage: number
  disk_read: number
  disk_write: number
  network_sent: number
  network_received: number
}

// API методы
export const agentsApi = {
  getAll: () => api.get<Agent[]>('/api/agents'),
  create: (name: string) => api.post<Agent>('/api/agents', { name }),
  update: (id: string, data: { name?: string; is_active?: boolean }) => 
    api.put(`/api/agents/${id}`, data),
  delete: (id: string) => api.delete(`/api/agents/${id}`),
  getDetail: (id: string) => api.get<AgentDetail>(`/api/agents/${id}`),
  getMetrics: (id: string, limit?: number) => 
    api.get<MetricPoint[]>(`/api/agents/${id}/metrics`, { 
      params: { limit } 
    }),
  getContainers: (id: string) => 
    api.get<Container[]>(`/api/agents/${id}/containers`),
  getImages: (id: string) => 
    api.get<Image[]>(`/api/agents/${id}/images`),
}

export const dashboardApi = {
  getData: () => api.get<DashboardData>('/api/dashboard'),
}

export const containersApi = {
  getAll: async (params?: { agent_id?: string; status?: string; search?: string }) => {
    const response = await api.get<ContainerListResponse>('/api/containers', { params })
    return {
      ...response,
      data: response.data.containers
    }
  },
  getDetail: (id: string) => api.get<ContainerDetail>(`/api/containers/${id}`),
  getLogs: (id: string) => api.get<ContainerLog[]>(`/api/containers/${id}/logs`),
}

export const imagesApi = {
  getAll: async (params?: { agent_id?: string; search?: string }) => {
    const response = await api.get<ImageListResponse>('/api/images', { params })
    return {
      ...response,
      data: response.data.images
    }
  },
  getDetail: (id: string) => api.get<ImageDetail>(`/api/images/${id}`),
}



// Утилитарные функции
export const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

export const formatCPUUsage = (usage?: number): string => {
  if (usage === undefined || usage === null) return '0%'
  return `${(usage * 100).toFixed(1)}%`
}

export const formatMemoryUsage = (usage?: number): string => {
  if (usage === undefined || usage === null) return '0%'
  return `${Math.round(usage)}%`
}

export const formatMemoryUsageMB = (usage?: number): string => {
  if (usage === undefined || usage === null) return '0 MB'
  return `${Math.round(usage)} MB`
}

export const getContainerStatusColor = (status: string): string => {
  switch (status.toLowerCase()) {
    case 'running':
      return 'text-green-600'
    case 'exited':
      return 'text-red-600'
    case 'paused':
      return 'text-yellow-600'
    case 'restarting':
      return 'text-blue-600'
    default:
      return 'text-gray-600'
  }
}

export const getAgentStatusColor = (status: string): string => {
  switch (status.toLowerCase()) {
    case 'online':
      return 'text-green-600'
    case 'offline':
      return 'text-red-600'
    default:
      return 'text-gray-600'
  }
}

// Action types
export interface Action {
  id: string
  agent_id: string
  type: string
  payload: Record<string, any>
  status: 'pending' | 'completed' | 'failed'
  created: string
  completed?: string
  response?: string
  error?: string
}

export interface CreateActionRequest {
  agent_id: string
  type: string
  payload: Record<string, any>
}

export interface ActionListResponse {
  actions: Action[]
  total: number
}

// API functions for actions
export const actionsApi = {
  create: (data: CreateActionRequest) => api.post<Action>('/api/actions', data),
  list: (params?: { agent_id?: string; status?: string }) => 
    api.get<ActionListResponse>('/api/actions', { params }),
}

// Notification types
export interface EmailSettings {
  enabled: boolean
  smtp_host: string
  smtp_port: number
  username: string
  password: string
  from_email: string
  from_name: string
  to_emails: string
  use_tls: boolean
  use_start_tls: boolean
}

export interface NotificationSettings {
  telegram_bot_token: string
  telegram_chat_id: string
  email_settings: EmailSettings
  notifications: {
    agent_offline: {
      enabled: boolean
      message: string
    }
    container_stopped: {
      enabled: boolean
      message: string
    }
    cpu_threshold: {
      enabled: boolean
      threshold: number
      message: string
    }
    ram_threshold: {
      enabled: boolean
      threshold: number
      message: string
    }
  }
}

// API functions for notifications
export const notificationsApi = {
  getSettings: () => api.get<NotificationSettings>('/api/notifications/settings'),
  updateSettings: (settings: NotificationSettings) => 
    api.post<NotificationSettings>('/api/notifications/settings', settings),
  sendTest: () => api.post<{ message: string }>('/api/notifications/test'),
}