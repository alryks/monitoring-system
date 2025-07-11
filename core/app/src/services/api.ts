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

export interface DashboardData {
  agents: Agent[]
  total_agents: number
  online_agents: number
  offline_agents: number
  recent_metrics: RecentMetric[]
  system_overview: SystemOverview
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
}

// API методы
export const agentsApi = {
  getAll: () => api.get<Agent[]>('/api/agents'),
  create: (name: string) => api.post<Agent>('/api/agents', { name }),
  update: (id: string, data: { name?: string; is_active?: boolean }) => 
    api.put(`/api/agents/${id}`, data),
  delete: (id: string) => api.delete(`/api/agents/${id}`),
  getMetrics: (id: string, limit?: number) => 
    api.get<MetricPoint[]>(`/api/agents/${id}/metrics`, { 
      params: { limit } 
    }),
  getContainers: (id: string) => 
    api.get<Container[]>(`/api/agents/${id}/containers`),
}

export const dashboardApi = {
  getData: () => api.get<DashboardData>('/api/dashboard'),
} 