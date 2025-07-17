import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  RefreshCw,
  Cpu,
  MemoryStick,
  CheckCircle,
  XCircle,
  HardDrive,
  Network,
} from 'lucide-react'
import {
  agentsApi,
  type AgentDetail as AgentDetailType,
  formatCPUUsage,
  formatBytes,
  getAgentStatusColor
} from '../services/api'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell
} from 'recharts'
import styles from './AgentDetail.module.css'



export default function AgentDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [data, setData] = useState<AgentDetailType | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState('')

  const fetchData = async (isRefresh = false) => {
    if (!id) return
    
    try {
      if (isRefresh) {
        setRefreshing(true)
      } else {
        setLoading(true)
      }
      const response = await agentsApi.getDetail(id)
      setData(response.data)
      setError('')
    } catch (error: any) {
      setError('Ошибка загрузки данных агента')
      console.error('Agent detail fetch error:', error)
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    fetchData()
    const interval = setInterval(() => fetchData(true), 30000)
    return () => clearInterval(interval)
  }, [id])

  if (loading && !data) {
    return (
      <div className={styles.loading}>
        <RefreshCw className={`${styles.refreshIcon} ${styles.spinning}`} />
        <span className={styles.loadingText}>Загрузка данных агента...</span>
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

  if (!data) {
    return (
      <div className={styles.error}>
        <span className={styles.errorText}>Агент не найден</span>
      </div>
    )
  }

  console.log(data.system_metrics)

  const cpuChartData = data.system_metrics?.slice(-20).map(metric => ({
    time: new Date(metric.timestamp).toLocaleTimeString('ru-RU', { 
      hour: '2-digit', 
      minute: '2-digit' 
    }),
    cpu: metric.cpu_usage * 100,
  })) || []

  const memoryChartData = data.system_metrics?.slice(-20).map(metric => ({
    time: new Date(metric.timestamp).toLocaleTimeString('ru-RU', { 
      hour: '2-digit', 
      minute: '2-digit' 
    }),
    memory: metric.ram_usage,
  })) || []

  const diskChartData = data.system_metrics?.slice(-20).map(metric => ({
    time: new Date(metric.timestamp).toLocaleTimeString('ru-RU', { 
      hour: '2-digit', 
      minute: '2-digit' 
    }),
    read: (metric.disk_read || 0) / 1024 / 1024, // MB/s
    write: (metric.disk_write || 0) / 1024 / 1024, // MB/s
  })) || []

  const networkChartData = data.system_metrics?.slice(-20).map(metric => ({
    time: new Date(metric.timestamp).toLocaleTimeString('ru-RU', { 
      hour: '2-digit', 
      minute: '2-digit' 
    }),
    sent: (metric.network_sent || 0) / 1024 / 1024, // MB/s
    received: (metric.network_received || 0) / 1024 / 1024, // MB/s
  })) || []

  const ramData = [
    { name: 'Использовано', value: data.metrics?.memory?.ram_usage || 0, color: '#3b82f6' },
    { name: 'Свободно', value: (data.metrics?.memory?.ram_total || 0) - (data.metrics?.memory?.ram_usage || 0), color: '#e5e7eb' },
  ]

  const swapData = [
    { name: 'Использовано', value: data.metrics?.memory?.swap_usage || 0, color: '#f59e0b' },
    { name: 'Свободно', value: (data.metrics?.memory?.swap_total || 0) - (data.metrics?.memory?.swap_usage || 0), color: '#e5e7eb' },
  ]

  const renderOverviewTab = () => (
    <div className={styles.overviewContent}>
      {/* System Status */}
      <div className={styles.statusGrid}>
        <div className={`${styles.statusCard} ${data.status === 'online' ? styles.online : styles.offline}`}>
          <div className={styles.statusHeader}>
            {data.status === 'online' ? (
              <CheckCircle className={styles.statusIcon} />
            ) : (
              <XCircle className={styles.statusIcon} />
            )}
            <span className={styles.statusLabel}>
              {data.status === 'online' ? 'Онлайн' : 'Оффлайн'}
            </span>
          </div>
          <div className={styles.statusInfo}>
            <span>IP: {data.public_ip || 'N/A'}</span>
            <span>Последний пинг: {data.last_ping ? new Date(data.last_ping).toLocaleString('ru-RU') : 'N/A'}</span>
          </div>
        </div>
      </div>

      {/* CPU Metrics */}
      <div className={styles.metricsGrid}>
        <div className={styles.metricCard}>
          <h3 className={styles.metricTitle}>
            <Cpu className={styles.metricIcon} />
            Загрузка CPU
          </h3>
          <div className={styles.cpuCores}>
            {data.metrics.cpu.map((cpu) => (
              <div key={cpu.name} className={styles.cpuCore}>
                <span className={styles.cpuName}>{cpu.name}</span>
                <div className={styles.cpuBar}>
                  <div 
                    className={styles.cpuFill} 
                    style={{ width: `${cpu.usage * 100}%` }}
                  />
                </div>
                <span className={styles.cpuValue}>{formatCPUUsage(cpu.usage)}</span>
              </div>
            ))}
          </div>
          {cpuChartData.length > 0 && (
            <div className={styles.chartContainer}>
              <ResponsiveContainer width="100%" height={200}>
                <LineChart data={cpuChartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip formatter={(value) => [`${Number(value).toFixed(1)}%`, 'CPU']} />
                  <Line type="monotone" dataKey="cpu" stroke="#3b82f6" strokeWidth={2} />
                </LineChart>
              </ResponsiveContainer>
            </div>
          )}
        </div>

        {/* Memory Metrics */}
        <div className={styles.metricCard}>
          <h3 className={styles.metricTitle}>
            <MemoryStick className={styles.metricIcon} />
            Использование памяти
          </h3>
          <div className={styles.memoryGrid}>
            <div className={styles.memorySection}>
              <h4>RAM</h4>
              <div className={styles.memoryInfo}>
                <span>{formatBytes(data.metrics.memory.ram_usage * 1024 * 1024)} / {formatBytes(data.metrics.memory.ram_total * 1024 * 1024)}</span>
                <span>{formatCPUUsage((data.metrics.memory.ram_usage / data.metrics.memory.ram_total))}</span>
              </div>
              <ResponsiveContainer width="100%" height={120}>
                <PieChart>
                  <Pie
                    data={ramData}
                    cx="50%"
                    cy="50%"
                    innerRadius={30}
                    outerRadius={50}
                    paddingAngle={5}
                    dataKey="value"
                  >
                    {ramData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip formatter={(value) => [formatBytes(Number(value) * 1024 * 1024), 'RAM']} />
                </PieChart>
              </ResponsiveContainer>
            </div>
            <div className={styles.memorySection}>
              <h4>Swap</h4>
              <div className={styles.memoryInfo}>
                <span>{formatBytes(data.metrics.memory.swap_usage * 1024 * 1024)} / {formatBytes(data.metrics.memory.swap_total * 1024 * 1024)}</span>
                <span>{formatCPUUsage((data.metrics.memory.swap_usage / data.metrics.memory.swap_total) * 100)}</span>
              </div>
              <ResponsiveContainer width="100%" height={120}>
                <PieChart>
                  <Pie
                    data={swapData}
                    cx="50%"
                    cy="50%"
                    innerRadius={30}
                    outerRadius={50}
                    paddingAngle={5}
                    dataKey="value"
                  >
                    {swapData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip formatter={(value) => [formatBytes(Number(value) * 1024 * 1024), 'Swap']} />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </div>
          {memoryChartData.length > 0 && (
            <div className={styles.chartContainer}>
              <ResponsiveContainer width="100%" height={200}>
                <LineChart data={memoryChartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip formatter={(value) => [`${Number(value).toFixed(1)}%`, 'Memory']} />
                  <Line type="monotone" dataKey="memory" stroke="#10b981" strokeWidth={2} />
                </LineChart>
              </ResponsiveContainer>
            </div>
          )}
        </div>

        {/* Disk I/O */}
        <div className={styles.metricCard}>
          <h3 className={styles.metricTitle}>
            <HardDrive className={styles.metricIcon} />
            Дисковая активность
          </h3>
          <div className={styles.diskList}>
            {data.metrics.disk.map((disk) => (
              <div key={disk.name} className={styles.diskItem}>
                <span className={styles.diskName}>{disk.name}</span>
                <div className={styles.diskMetrics}>
                  <span>R: {formatBytes(disk.read_bytes)}</span>
                  <span>W: {formatBytes(disk.write_bytes)}</span>
                </div>
              </div>
            ))}
          </div>
          {diskChartData.length > 0 && (
            <div className={styles.chartContainer}>
              <ResponsiveContainer width="100%" height={200}>
                <LineChart data={diskChartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip formatter={(value, name) => [`${Number(value).toFixed(2)} MB/s`, name === 'read' ? 'Чтение' : 'Запись']} />
                  <Line type="monotone" dataKey="read" stroke="#3b82f6" strokeWidth={2} />
                  <Line type="monotone" dataKey="write" stroke="#ef4444" strokeWidth={2} />
                </LineChart>
              </ResponsiveContainer>
            </div>
          )}
        </div>

        {/* Network I/O */}
        <div className={styles.metricCard}>
          <h3 className={styles.metricTitle}>
            <Network className={styles.metricIcon} />
            Сетевая активность
          </h3>
          <div className={styles.networkInfo}>
            <div className={styles.networkItem}>
              <span className={styles.networkLabel}>Public IP:</span>
              <span className={styles.networkValue}>{data.metrics.network.public_ip}</span>
            </div>
            <div className={styles.networkItem}>
              <span className={styles.networkLabel}>Отправлено:</span>
              <span className={styles.networkValue}>{formatBytes(data.metrics.network.sent_bytes)}</span>
            </div>
            <div className={styles.networkItem}>
              <span className={styles.networkLabel}>Получено:</span>
              <span className={styles.networkValue}>{formatBytes(data.metrics.network.received_bytes)}</span>
            </div>
          </div>
          {networkChartData.length > 0 && (
            <div className={styles.chartContainer}>
              <ResponsiveContainer width="100%" height={200}>
                <LineChart data={networkChartData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip formatter={(value, name) => [`${Number(value).toFixed(2)} MB/s`, name === 'sent' ? 'Отправлено' : 'Получено']} />
                  <Line type="monotone" dataKey="sent" stroke="#f59e0b" strokeWidth={2} />
                  <Line type="monotone" dataKey="received" stroke="#06b6d4" strokeWidth={2} />
                </LineChart>
              </ResponsiveContainer>
            </div>
          )}
        </div>
      </div>
    </div>
  )

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <button 
          className={styles.backButton}
          onClick={() => navigate('/agents')}
        >
          <ArrowLeft className={styles.backIcon} />
          Назад к агентам
        </button>
        <div className={styles.titleSection}>
          <h1 className={styles.title}>{data.name}</h1>
          <div className={styles.agentMeta}>
            <span className={`${styles.agentStatus} ${getAgentStatusColor(data.status)}`}>
              {data.status === 'online' ? <CheckCircle /> : <XCircle />}
              {data.status === 'online' ? 'Онлайн' : 'Оффлайн'}
            </span>
            <span className={styles.agentIP}>{data.public_ip || 'N/A'}</span>
          </div>
        </div>
        <button
          onClick={() => fetchData(true)}
          disabled={refreshing}
          className={styles.refreshButton}
        >
          <RefreshCw className={`${styles.refreshIcon} ${refreshing ? styles.spinning : ''}`} />
          Обновить
        </button>
      </div>

      {error && (
        <div className={styles.error}>
          <span className={styles.errorText}>{error}</span>
        </div>
      )}

      <div className={styles.tabsContainer}>
        <div className={styles.tabsList}>
          {/* Удалить все переменные, функции, JSX-блоки, связанные с tabs, containers, images, activeTab, setActiveTab, renderContainersTab, renderImagesTab, renderTabContent и т.д. */}
          {/* Оставить только рендеринг renderOverviewTab() в основном return. */}
        </div>

        <div className={styles.tabContent}>
          {renderOverviewTab()}
        </div>
      </div>
    </div>
  )
}