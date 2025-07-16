import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  RefreshCw,
  Activity,
  Info,
  BarChart3,
  FileText,
  Container as ContainerIcon,
  Server,
  Network,
  HardDrive,
  Search,
  CheckCircle,
  XCircle,
  Pause,
  RotateCcw
} from 'lucide-react'
import {
  containersApi,
  type ContainerDetail as ContainerDetailType,
  formatCPUUsage,
  formatMemoryUsageMB,
  formatBytes,
  getContainerStatusColor
} from '../services/api'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer
} from 'recharts'
import styles from './ContainerDetail.module.css'

type TabType = 'details' | 'resources' | 'logs'

export default function ContainerDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [data, setData] = useState<ContainerDetailType | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState('')
  const [activeTab, setActiveTab] = useState<TabType>('details')
  const [logSearch, setLogSearch] = useState('')
  const [autoScroll, setAutoScroll] = useState(false)

  const fetchData = async (isRefresh = false) => {
    if (!id) return
    
    try {
      if (isRefresh) {
        setRefreshing(true)
      } else {
        setLoading(true)
      }
      const response = await containersApi.getDetail(id)
      setData(response.data)
      setError('')
    } catch (error: any) {
      setError('Ошибка загрузки данных контейнера')
      console.error('Container detail fetch error:', error)
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
        <span className={styles.loadingText}>Загрузка данных контейнера...</span>
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
        <span className={styles.errorText}>Контейнер не найден</span>
      </div>
    )
  }

  const getStatusIcon = (status: string) => {
    if (status.toLowerCase().includes('running')) return <CheckCircle className={styles.statusIcon} />
    if (status.toLowerCase().includes('exited')) return <XCircle className={styles.statusIcon} />
    if (status.toLowerCase().includes('paused')) return <Pause className={styles.statusIcon} />
    if (status.toLowerCase().includes('restarting')) return <RotateCcw className={styles.statusIcon} />
    return <XCircle className={styles.statusIcon} />
  }

  const getStatusText = (status: string) => {
    if (status.toLowerCase().includes('running')) return 'Запущен'
    if (status.toLowerCase().includes('exited')) return 'Остановлен'
    if (status.toLowerCase().includes('paused')) return 'Приостановлен'
    if (status.toLowerCase().includes('restarting')) return 'Перезапускается'
    return 'Неизвестно'
  }

  const tabs = [
    { id: 'details', label: 'Детали', icon: Info },
    { id: 'resources', label: 'Ресурсы', icon: BarChart3 },
    { id: 'logs', label: 'Логи', icon: FileText },
  ]

  const renderDetailsTab = () => (
    <div className={styles.detailsContent}>
      {/* Container Info */}
      <div className={styles.infoGrid}>
        <div className={styles.infoCard}>
          <div className={styles.infoHeader}>
            <ContainerIcon className={styles.infoIcon} />
            <span className={styles.infoTitle}>Информация о контейнере</span>
          </div>
          <div className={styles.infoBody}>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>Container ID:</span>
              <span className={styles.infoValue}>{data.container_id}</span>
            </div>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>Образ:</span>
              <span className={styles.infoValue}>{data.image_id}</span>
            </div>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>Создан:</span>
              <span className={styles.infoValue}>
                {new Date(data.created_at).toLocaleString('ru-RU')}
              </span>
            </div>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>Перезапуски:</span>
              <span className={styles.infoValue}>{data.restart_count}</span>
            </div>
          </div>
        </div>

        <div className={styles.infoCard}>
          <div className={styles.infoHeader}>
            <Network className={styles.infoIcon} />
            <span className={styles.infoTitle}>Сетевая информация</span>
          </div>
          <div className={styles.infoBody}>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>IP адрес:</span>
              <span className={styles.infoValue}>{data.ip_address || 'N/A'}</span>
            </div>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>MAC адрес:</span>
              <span className={styles.infoValue}>{data.mac_address || 'N/A'}</span>
            </div>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>Отправлено:</span>
              <span className={styles.infoValue}>
                {formatBytes(data.network_sent_bytes || 0)}
              </span>
            </div>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>Получено:</span>
              <span className={styles.infoValue}>
                {formatBytes(data.network_received_bytes || 0)}
              </span>
            </div>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>Сети:</span>
              <span className={styles.infoValue}>
                {data.networks?.length ? data.networks.join(', ') : 'N/A'}
              </span>
            </div>
          </div>
        </div>

        <div className={styles.infoCard}>
          <div className={styles.infoHeader}>
            <HardDrive className={styles.infoIcon} />
            <span className={styles.infoTitle}>Хранилище</span>
          </div>
          <div className={styles.infoBody}>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>Тома:</span>
              <span className={styles.infoValue}>
                {data.volumes?.length ? data.volumes.join(', ') : 'N/A'}
              </span>
            </div>
          </div>
        </div>

        <div className={styles.infoCard}>
          <div className={styles.infoHeader}>
            <Server className={styles.infoIcon} />
            <span className={styles.infoTitle}>Агент</span>
          </div>
          <div className={styles.infoBody}>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>Имя агента:</span>
              <span className={styles.infoValue}>{data.agent.name}</span>
            </div>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>Статус агента:</span>
              <span className={`${styles.infoValue} ${getContainerStatusColor(data.agent.status)}`}>
                {data.agent.status === 'online' ? 'Онлайн' : 'Оффлайн'}
              </span>
            </div>
            <div className={styles.infoRow}>
              <span className={styles.infoLabel}>IP агента:</span>
              <span className={styles.infoValue}>{data.agent.public_ip || 'N/A'}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )

  const renderResourcesTab = () => (
    <div className={styles.resourcesContent}>
      {/* Current Metrics */}
      <div className={styles.currentMetrics}>
        <div className={styles.metricCard}>
          <div className={styles.metricHeader}>
            <Activity className={styles.metricIcon} />
            <span className={styles.metricTitle}>Текущие ресурсы</span>
          </div>
          <div className={styles.metricGrid}>
            <div className={styles.metricItem}>
              <span className={styles.metricLabel}>CPU:</span>
              <span className={styles.metricValue}>
                {formatCPUUsage(data.cpu_usage_percent)}
              </span>
            </div>
            <div className={styles.metricItem}>
              <span className={styles.metricLabel}>Память:</span>
              <span className={styles.metricValue}>
                {formatMemoryUsageMB(data.memory_usage_mb)}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* History Charts */}
      {data.history && data.history.length > 0 && (
        <div className={styles.chartsGrid}>
          <div className={styles.chartCard}>
            <h3 className={styles.chartTitle}>История CPU</h3>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={data.history.slice(-20)}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis 
                  dataKey="timestamp" 
                  tickFormatter={(time) => new Date(time).toLocaleTimeString('ru-RU', { 
                    hour: '2-digit', 
                    minute: '2-digit' 
                  })} 
                />
                <YAxis />
                <Tooltip 
                  formatter={(value) => [`${Number(value).toFixed(1)}%`, 'CPU']}
                  labelFormatter={(time) => new Date(time).toLocaleString('ru-RU')}
                />
                <Line type="monotone" dataKey="cpu_usage" stroke="#3b82f6" strokeWidth={2} />
              </LineChart>
            </ResponsiveContainer>
          </div>

          <div className={styles.chartCard}>
            <h3 className={styles.chartTitle}>История памяти</h3>
            <ResponsiveContainer width="100%" height={300}>
              <LineChart data={data.history.slice(-20)}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis 
                  dataKey="timestamp" 
                  tickFormatter={(time) => new Date(time).toLocaleTimeString('ru-RU', { 
                    hour: '2-digit', 
                    minute: '2-digit' 
                  })} 
                />
                <YAxis />
                <Tooltip 
                  formatter={(value) => [`${formatMemoryUsageMB(Number(value))}`, 'Memory']}
                  labelFormatter={(time) => new Date(time).toLocaleString('ru-RU')}
                />
                <Line type="monotone" dataKey="memory_usage" stroke="#10b981" strokeWidth={2} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}
    </div>
  )

  const renderLogsTab = () => (
    <div className={styles.logsContent}>
      <div className={styles.logsControls}>
        <div className={styles.logSearch}>
          <Search className={styles.searchIcon} />
          <input
            type="text"
            placeholder="Поиск в логах..."
            value={logSearch}
            onChange={(e) => setLogSearch(e.target.value)}
            className={styles.searchInput}
          />
        </div>
        <label className={styles.autoScrollLabel}>
          <input
            type="checkbox"
            checked={autoScroll}
            onChange={(e) => setAutoScroll(e.target.checked)}
          />
          Auto-scroll
        </label>
      </div>
      
      <div className={styles.logsContainer}>
        {data.logs && data.logs.length > 0 ? (
          data.logs
            .filter(log => log.log_line.toLowerCase().includes(logSearch.toLowerCase()))
            .map((log) => (
              <div key={log.id} className={styles.logLine}>
                <span className={styles.logTimestamp}>
                  {new Date(log.timestamp).toLocaleString('ru-RU')}
                </span>
                <span className={styles.logContent}>{log.log_line}</span>
              </div>
            ))
        ) : (
          <div className={styles.noLogs}>Нет логов для отображения</div>
        )}
      </div>
    </div>
  )

  return (
    <div className={styles.container}>
      {/* Header */}
      <div className={styles.header}>
        <div className={styles.headerMain}>
          <button onClick={() => navigate(-1)} className={styles.backButton}>
            <ArrowLeft className={styles.backIcon} />
          </button>
          <div className={styles.titleSection}>
            <h1 className={styles.title}>{data.name}</h1>
            <div className={`${styles.statusBadge} ${getContainerStatusColor(data.status)}`}>
              {getStatusIcon(data.status)}
              <span>{getStatusText(data.status)}</span>
            </div>
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

      {/* Tabs */}
      <div className={styles.tabs}>
        {tabs.map((tab) => {
          const Icon = tab.icon
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as TabType)}
              className={`${styles.tab} ${activeTab === tab.id ? styles.active : ''}`}
            >
              <Icon className={styles.tabIcon} />
              {tab.label}
            </button>
          )
        })}
      </div>

      {/* Content */}
      <div className={styles.content}>
        {activeTab === 'details' && renderDetailsTab()}
        {activeTab === 'resources' && renderResourcesTab()}
        {activeTab === 'logs' && renderLogsTab()}
      </div>
    </div>
  )
} 