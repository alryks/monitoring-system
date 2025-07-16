import { useState, useEffect } from 'react'
import { 
  LineChart, 
  Line, 
  BarChart, 
  Bar, 
  PieChart, 
  Pie, 
  Cell, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer 
} from 'recharts'
import { 
  Server, 
  Container, 
  Cpu, 
  MemoryStick, 
  TrendingUp, 
  TrendingDown, 
  RefreshCw 
} from 'lucide-react'
import { 
  dashboardApi, 
  type DashboardData,
  formatCPUUsage,
  formatMemoryUsage,
} from '../services/api'
import styles from './Dashboard.module.css'

export default function Dashboard() {
  const [data, setData] = useState<DashboardData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchData = async () => {
    try {
      setLoading(true)
      const response = await dashboardApi.getData()
      setData(response.data)
      setError(null)
    } catch (err) {
      setError('Ошибка загрузки данных дашборда')
      console.error('Dashboard error:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
    const interval = setInterval(fetchData, 30000) // Обновляем каждые 30 секунд
    return () => clearInterval(interval)
  }, [])

  if (loading && !data) {
    return (
      <div className={styles.loading}>
        <RefreshCw className={styles.spinner} />
        <span>Загрузка данных...</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className={styles.error}>
        <span>{error}</span>
        <button onClick={fetchData} className={styles.retryButton}>
          <RefreshCw size={16} />
          Повторить
        </button>
      </div>
    )
  }

  if (!data) {
    return <div className={styles.error}>Нет данных для отображения</div>
  }

  const kpiCards = [
    {
      name: 'Агенты онлайн',
      value: `${data?.kpis?.agents_online || 0} / ${data?.kpis?.agents_total || 0}`,
      icon: Server,
      color: 'green',
      trend: data?.kpis?.agents_online === data?.kpis?.agents_total ? 'positive' : 'neutral'
    },
    {
      name: 'Всего контейнеров',
      value: data?.kpis?.containers_total || 0,
      subValue: `${data?.kpis?.containers_running || 0} запущено, ${data?.kpis?.containers_stopped || 0} остановлено`,
      icon: Container,
      color: 'blue',
      trend: 'neutral'
    },
    {
      name: 'Средняя нагрузка CPU',
      value: formatCPUUsage(data?.kpis?.avg_cpu_usage  || 0),
      icon: Cpu,
      color: data?.kpis?.avg_cpu_usage > 0.8 ? 'red' : data?.kpis?.avg_cpu_usage > 0.6 ? 'yellow' : 'green',
      trend: data?.kpis?.avg_cpu_usage > 0.7 ? 'negative' : 'neutral'
    },
    {
      name: 'Средняя нагрузка RAM',
      value: formatMemoryUsage(data?.kpis?.avg_memory_usage || 0),
      icon: MemoryStick,
      color: data?.kpis?.avg_memory_usage > 80 ? 'red' : data?.kpis?.avg_memory_usage > 60 ? 'yellow' : 'green',
      trend: data?.kpis?.avg_memory_usage > 70 ? 'negative' : 'neutral'
    }
  ]

  // Форматируем данные для графиков
  const resourceData = (data.resource_usage || []).map(point => ({
    time: new Date(point.timestamp).toLocaleTimeString('ru-RU', { 
      hour: '2-digit', 
      minute: '2-digit' 
    }),
    cpu: point.cpu,
    memory: point.memory
  }))

  const networkData = (data.network_activity || []).map(point => ({
    time: new Date(point.timestamp).toLocaleTimeString('ru-RU', { 
      hour: '2-digit', 
      minute: '2-digit' 
    }),
    sent: point.sent / (1024 * 1024), // Конвертируем в MB
    received: point.received / (1024 * 1024)
  }))

  // Данные для круговой диаграммы контейнеров
  const containerDistribution = [
    { name: 'Запущено', value: data.kpis.containers_running, color: '#22c55e' },
    { name: 'Остановлено', value: data.kpis.containers_stopped, color: '#ef4444' }
  ]

  // Безопасные данные для графиков с проверкой на null
  const topCPUData = (data.top_containers_cpu || []).map(container => ({
    ...container,
    cpu_usage: container.cpu_usage ? container.cpu_usage * 100 : 0 // Конвертируем в проценты
  }))

  return (
    <div className={styles.dashboard}>
      <div className={styles.header}>
        <h1>Мониторинг системы</h1>
        <button onClick={fetchData} className={styles.refreshButton} disabled={loading}>
          <RefreshCw size={16} className={loading ? styles.spinning : ''} />
          Обновить
        </button>
      </div>

      {/* KPI Cards */}
      <div className={styles.kpiGrid}>
        {kpiCards.map((kpi, index) => (
          <div key={index} className={`${styles.kpiCard} ${styles[kpi.color]}`}>
            <div className={styles.kpiIcon}>
              <kpi.icon size={24} />
            </div>
            <div className={styles.kpiContent}>
              <div className={styles.kpiValue}>{kpi.value}</div>
              <div className={styles.kpiName}>{kpi.name}</div>
              {kpi.subValue && <div className={styles.kpiSubValue}>{kpi.subValue}</div>}
            </div>
            <div className={styles.kpiTrend}>
              {kpi.trend === 'positive' && <TrendingUp size={16} className={styles.trendUp} />}
              {kpi.trend === 'negative' && <TrendingDown size={16} className={styles.trendDown} />}
            </div>
          </div>
        ))}
      </div>

      {/* Charts Grid */}
      <div className={styles.chartsGrid}>
        {/* Использование ресурсов */}
        <div className={styles.chartCard}>
          <h3>Использование ресурсов</h3>
          {resourceData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={resourceData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="time" />
                <YAxis />
                <Tooltip />
                <Line 
                  type="monotone" 
                  dataKey="cpu" 
                  stroke="#3b82f6" 
                  strokeWidth={2}
                  name="CPU (%)"
                />
                <Line 
                  type="monotone" 
                  dataKey="memory" 
                  stroke="#10b981" 
                  strokeWidth={2}
                  name="Память (%)"
                />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <div className={styles.noData}>Нет данных для отображения</div>
          )}
        </div>

        {/* Сетевая активность */}
        <div className={styles.chartCard}>
          <h3>Сетевая активность</h3>
          {networkData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={networkData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="time" />
                <YAxis />
                <Tooltip formatter={(value: number) => [`${value.toFixed(2)} MB`, '']} />
                <Line 
                  type="monotone" 
                  dataKey="sent" 
                  stroke="#f59e0b" 
                  strokeWidth={2}
                  name="Отправлено (MB)"
                />
                <Line 
                  type="monotone" 
                  dataKey="received" 
                  stroke="#8b5cf6" 
                  strokeWidth={2}
                  name="Получено (MB)"
                />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <div className={styles.noData}>Нет данных для отображения</div>
          )}
        </div>

        {/* Топ контейнеров по CPU */}
        <div className={styles.chartCard}>
          <h3>Топ контейнеров по CPU</h3>
          {topCPUData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={topCPUData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis 
                  dataKey="name" 
                  angle={-45}
                  textAnchor="end"
                  height={80}
                  interval={0}
                />
                <YAxis 
                  label={{ value: 'CPU (%)', angle: -90, position: 'insideLeft' }}
                />
                <Tooltip 
                  formatter={(value: number) => [`${value.toFixed(3)}%`, 'CPU Usage']}
                  labelFormatter={(label) => `Контейнер: ${label}`}
                />
                <Bar dataKey="cpu_usage" fill="#ef4444" />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <div className={styles.noData}>Нет данных для отображения</div>
          )}
        </div>

        {/* Распределение контейнеров */}
        <div className={styles.chartCard}>
          <h3>Статус контейнеров</h3>
          {containerDistribution.some(item => item.value > 0) ? (
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={containerDistribution}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, value }) => `${name}: ${value}`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {containerDistribution.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <div className={styles.noData}>Нет контейнеров для отображения</div>
          )}
        </div>
      </div>

      {/* Agents Summary Table */}
      <div className={styles.agentsSection}>
        <h2>Сводка по агентам</h2>
        <div className={styles.agentsTable}>
          <div className={styles.tableHeader}>
            <div>Агент</div>
            <div>Статус</div>
            <div>Контейнеры</div>
            <div>CPU</div>
            <div>Память</div>
            <div>Последний пинг</div>
          </div>
          {data.agents_summary.map((agent) => (
            <div key={agent.id} className={styles.tableRow}>
              <div className={styles.agentName}>{agent.name}</div>
              <div className={styles.agentStatus}>
                <span className={`${styles.statusDot} ${styles[agent.status]}`}></span>
                {agent.status === 'online' ? 'Онлайн' : 
                 agent.status === 'offline' ? 'Оффлайн' : 'Неизвестно'}
              </div>
              <div>{agent.containers}</div>
              <div>{formatCPUUsage(agent.cpu_usage)}</div>
              <div>{formatMemoryUsage(agent.memory_usage)}</div>
              <div>
                {agent.last_ping 
                  ? new Date(agent.last_ping).toLocaleString('ru-RU')
                  : 'Никогда'
                }
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
} 