import { useState, useEffect } from 'react'
import { 
  Server, 
  HardDrive, 
  CheckCircle,
  XCircle,
  RefreshCw
} from 'lucide-react'
import { dashboardApi, type DashboardData } from '../services/api'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import styles from './Dashboard.module.css'

export default function Dashboard() {
  const [data, setData] = useState<DashboardData | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState('')

  const fetchData = async (isRefresh = false) => {
    try {
      if (isRefresh) {
        setRefreshing(true)
      } else {
        setLoading(true)
      }
      const response = await dashboardApi.getData()
      setData(response.data)
      setError('')
    } catch (error: any) {
      setError('Ошибка загрузки данных')
      console.error('Dashboard data fetch error:', error)
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    fetchData()
    const interval = setInterval(() => fetchData(true), 30000)
    return () => clearInterval(interval)
  }, [])

  if (loading && !data) {
    return (
      <div className={styles.loading}>
        <RefreshCw className={`${styles.refreshIcon} ${styles.spinning}`} />
        <span className={styles.loadingText}>Загрузка...</span>
      </div>
    )
  }

  if (error && !data) {
    return (
      <div className={styles.error}>
        <span className={styles.errorText}>{error}</span>
      </div>
    )
  }

  const stats = [
    {
      name: 'Всего агентов',
      value: data?.total_agents || 0,
      icon: Server,
    },
    {
      name: 'Онлайн',
      value: data?.online_agents || 0,
      icon: CheckCircle,
    },
    {
      name: 'Оффлайн',
      value: data?.offline_agents || 0,
      icon: XCircle,
    },
    {
      name: 'Контейнеры',
      value: data?.system_overview?.total_containers || 0,
      icon: HardDrive,
    },
  ]

  const chartData = data?.recent_metrics && Array.isArray(data.recent_metrics) 
    ? data.recent_metrics.slice(0, 10).reverse().map(metric => ({
        time: new Date(metric.timestamp).toLocaleTimeString('ru-RU', { 
          hour: '2-digit', 
          minute: '2-digit' 
        }),
        cpu: Math.round(metric.cpu_usage * 100),
        ram: Math.round(metric.ram_usage),
        agent: metric.agent_name,
      }))
    : []

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1 className={styles.title}>Дашборд</h1>
        <div className={styles.subtitle}>
          <button
            onClick={() => fetchData(true)}
            disabled={refreshing}
            className={styles.refreshButton}
          >
            <RefreshCw className={`${styles.refreshIcon} ${refreshing ? styles.spinning : ''}`} />
            Обновить
          </button>
        </div>
      </div>

      {error && (
        <div className={styles.error}>
          <span className={styles.errorText}>{error}</span>
        </div>
      )}

      <div className={styles.grid}>
        {stats.map((stat) => {
          const Icon = stat.icon
          return (
            <div key={stat.name} className={styles.card}>
              <div className={styles.cardHeader}>
                <span className={styles.cardTitle}>{stat.name}</span>
                <Icon className={styles.cardIcon} />
              </div>
              <div className={styles.cardValue}>{stat.value}</div>
            </div>
          )
        })}
      </div>

      <div className={styles.section}>
        <div className={styles.chartContainer}>
          <h3 className={styles.chartTitle}>Использование ресурсов</h3>
          {chartData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="time" />
                <YAxis />
                <Tooltip 
                  formatter={(value, name) => [
                    `${value}%`, 
                    name === 'cpu' ? 'CPU' : 'RAM'
                  ]}
                  labelFormatter={(label) => `Время: ${label}`}
                />
                <Line 
                  type="monotone" 
                  dataKey="cpu" 
                  stroke="#3b82f6" 
                  strokeWidth={2}
                  name="cpu"
                />
                <Line 
                  type="monotone" 
                  dataKey="ram" 
                  stroke="#10b981" 
                  strokeWidth={2}
                  name="ram"
                />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <div className={styles.loading}>
              <span>Нет данных для отображения</span>
            </div>
          )}
        </div>
      </div>

      <div className={styles.section}>
        <h2 className={styles.sectionTitle}>Активность агентов</h2>
        {data?.agents && Array.isArray(data.agents) && data.agents.length > 0 ? (
          <div className={styles.agentsList}>
            {data.agents.slice(0, 6).map((agent) => (
              <div key={agent.id} className={styles.agentCard}>
                <div className={styles.agentHeader}>
                  <span className={styles.agentName}>{agent.name}</span>
                  <div className={styles.agentStatus}>
                    {agent.status === 'online' ? (
                      <CheckCircle className={`${styles.statusIcon} ${styles.online}`} />
                    ) : (
                      <XCircle className={`${styles.statusIcon} ${styles.offline}`} />
                    )}
                    <span>{agent.status === 'online' ? 'Онлайн' : 'Оффлайн'}</span>
                  </div>
                </div>
                <div className={styles.agentInfo}>
                  <div className={styles.infoItem}>
                    <span className={styles.infoLabel}>IP адрес</span>
                    <span className={styles.infoValue}>{agent.public_ip || 'Неизвестен'}</span>
                  </div>
                  <div className={styles.infoItem}>
                    <span className={styles.infoLabel}>Последний пинг</span>
                    <span className={styles.infoValue}>
                      {agent.last_ping ? 
                        new Date(agent.last_ping).toLocaleString('ru-RU') : 
                        'Нет пингов'
                      }
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className={styles.loading}>
            <span>Нет агентов</span>
          </div>
        )}
      </div>
    </div>
  )
} 