import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  ArrowLeft,
  RefreshCw,
  Activity,
  Container,
  Image,
  HardDrive,
  Network,
  Cpu,
  MemoryStick,
  ChevronDown,
  ChevronUp,
  Search,
  CheckCircle,
  XCircle,
} from 'lucide-react'
import {
  agentsApi,
  type AgentDetail as AgentDetailType,
  formatCPUUsage,
  formatMemoryUsage,
  formatBytes,
  getContainerStatusColor,
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

type TabType = 'overview' | 'containers' | 'images' | 'volumes' | 'networks'

export default function AgentDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [data, setData] = useState<AgentDetailType | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState('')
  const [activeTab, setActiveTab] = useState<TabType>('overview')
  const [expandedContainer, setExpandedContainer] = useState<string | null>(null)
  const [containerSubTab, setContainerSubTab] = useState<'details' | 'resources' | 'logs'>('details')
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

  const handleContainerToggle = (containerId: string) => {
    setExpandedContainer(expandedContainer === containerId ? null : containerId)
    setContainerSubTab('details')
  }

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

  const tabs = [
    { id: 'overview', label: 'Обзор метрик', icon: Activity },
    { id: 'containers', label: 'Контейнеры', icon: Container },
    { id: 'images', label: 'Образы', icon: Image },
    { id: 'volumes', label: 'Тома', icon: HardDrive },
    { id: 'networks', label: 'Сети', icon: Network },
  ]

  const cpuChartData = data.system_metrics?.slice(-20).map(metric => ({
    time: new Date(metric.timestamp).toLocaleTimeString('ru-RU', { 
      hour: '2-digit', 
      minute: '2-digit' 
    }),
    cpu: metric.cpu_usage,
  })) || []

  const memoryChartData = data.system_metrics?.slice(-20).map(metric => ({
    time: new Date(metric.timestamp).toLocaleTimeString('ru-RU', { 
      hour: '2-digit', 
      minute: '2-digit' 
    }),
    memory: metric.memory_usage,
  })) || []

  const diskChartData = data.system_metrics?.slice(-20).map(metric => ({
    time: new Date(metric.timestamp).toLocaleTimeString('ru-RU', { 
      hour: '2-digit', 
      minute: '2-digit' 
    }),
    read: metric.disk_read / 1024 / 1024, // MB/s
    write: metric.disk_write / 1024 / 1024, // MB/s
  })) || []

  const networkChartData = data.system_metrics?.slice(-20).map(metric => ({
    time: new Date(metric.timestamp).toLocaleTimeString('ru-RU', { 
      hour: '2-digit', 
      minute: '2-digit' 
    }),
    sent: metric.network_sent / 1024 / 1024, // MB/s
    received: metric.network_received / 1024 / 1024, // MB/s
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
                    style={{ width: `${cpu.usage}%` }}
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
                <span>{formatCPUUsage((data.metrics.memory.ram_usage / data.metrics.memory.ram_total) * 100)}</span>
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

  const renderContainersTab = () => (
    <div className={styles.containersContent}>
      <div className={styles.containersList}>
        {data.containers.map((container) => (
          <div key={container.id} className={styles.containerCard}>
            <div 
              className={styles.containerHeader}
              onClick={() => handleContainerToggle(container.id)}
            >
              <div className={styles.containerInfo}>
                <div className={`${styles.containerStatus} ${getContainerStatusColor(container.status)}`}>
                  {container.status === 'running' ? <CheckCircle /> : <XCircle />}
                </div>
                <div className={styles.containerDetails}>
                  <h3 className={styles.containerName}>{container.name}</h3>
                  <span className={styles.containerImage}>{container.image_id}</span>
                </div>
              </div>
              <div className={styles.containerMetrics}>
                <span>CPU: {formatCPUUsage(container.cpu_usage_percent)}</span>
                <span>RAM: {formatMemoryUsage(container.memory_usage_mb)}</span>
                <span>Restarts: {container.restart_count}</span>
              </div>
              <div className={styles.containerToggle}>
                {expandedContainer === container.id ? <ChevronUp /> : <ChevronDown />}
              </div>
            </div>

            {expandedContainer === container.id && (
              <div className={styles.containerExpanded}>
                <div className={styles.containerSubTabs}>
                  <button 
                    className={`${styles.subTab} ${containerSubTab === 'details' ? styles.active : ''}`}
                    onClick={() => setContainerSubTab('details')}
                  >
                    Детали
                  </button>
                  <button 
                    className={`${styles.subTab} ${containerSubTab === 'resources' ? styles.active : ''}`}
                    onClick={() => setContainerSubTab('resources')}
                  >
                    Ресурсы
                  </button>
                  <button 
                    className={`${styles.subTab} ${containerSubTab === 'logs' ? styles.active : ''}`}
                    onClick={() => setContainerSubTab('logs')}
                  >
                    Логи
                  </button>
                </div>

                <div className={styles.containerSubContent}>
                  {containerSubTab === 'details' && (
                    <div className={styles.containerDetailsContent}>
                      <div className={styles.detailsGrid}>
                        <div className={styles.detailItem}>
                          <span className={styles.detailLabel}>Container ID:</span>
                          <span className={styles.detailValue}>{container.container_id}</span>
                        </div>
                        <div className={styles.detailItem}>
                          <span className={styles.detailLabel}>Создан:</span>
                          <span className={styles.detailValue}>{new Date(container.created_at).toLocaleString('ru-RU')}</span>
                        </div>
                        <div className={styles.detailItem}>
                          <span className={styles.detailLabel}>IP адрес:</span>
                          <span className={styles.detailValue}>{container.ip_address || 'N/A'}</span>
                        </div>
                        <div className={styles.detailItem}>
                          <span className={styles.detailLabel}>MAC адрес:</span>
                          <span className={styles.detailValue}>{container.mac_address || 'N/A'}</span>
                        </div>
                        <div className={styles.detailItem}>
                          <span className={styles.detailLabel}>Сети:</span>
                          <span className={styles.detailValue}>{container.networks.join(', ') || 'N/A'}</span>
                        </div>
                        <div className={styles.detailItem}>
                          <span className={styles.detailLabel}>Тома:</span>
                          <span className={styles.detailValue}>{container.volumes.join(', ') || 'N/A'}</span>
                        </div>
                      </div>
                    </div>
                  )}

                  {containerSubTab === 'resources' && (
                    <div className={styles.containerResourcesContent}>
                      <div className={styles.resourcesGrid}>
                        <div className={styles.resourceChart}>
                          <h4>История CPU</h4>
                          {container.history.length > 0 ? (
                            <ResponsiveContainer width="100%" height={150}>
                              <LineChart data={container.history.slice(-20)}>
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="timestamp" tickFormatter={(time) => new Date(time).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })} />
                                <YAxis />
                                <Tooltip formatter={(value) => [`${Number(value).toFixed(1)}%`, 'CPU']} />
                                <Line type="monotone" dataKey="cpu_usage" stroke="#3b82f6" strokeWidth={2} />
                              </LineChart>
                            </ResponsiveContainer>
                          ) : (
                            <div className={styles.noData}>Нет данных</div>
                          )}
                        </div>
                        <div className={styles.resourceChart}>
                          <h4>История памяти</h4>
                          {container.history.length > 0 ? (
                            <ResponsiveContainer width="100%" height={150}>
                              <LineChart data={container.history.slice(-20)}>
                                <CartesianGrid strokeDasharray="3 3" />
                                <XAxis dataKey="timestamp" tickFormatter={(time) => new Date(time).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })} />
                                <YAxis />
                                <Tooltip formatter={(value) => [`${formatMemoryUsage(Number(value))}`, 'Memory']} />
                                <Line type="monotone" dataKey="memory_usage" stroke="#10b981" strokeWidth={2} />
                              </LineChart>
                            </ResponsiveContainer>
                          ) : (
                            <div className={styles.noData}>Нет данных</div>
                          )}
                        </div>
                      </div>
                    </div>
                  )}

                  {containerSubTab === 'logs' && (
                    <div className={styles.containerLogsContent}>
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
                        {container.logs
                          .filter(log => log.log_line.toLowerCase().includes(logSearch.toLowerCase()))
                          .map((log) => (
                            <div key={log.id} className={styles.logLine}>
                              <span className={styles.logTimestamp}>
                                {new Date(log.timestamp).toLocaleString('ru-RU')}
                              </span>
                              <span className={styles.logContent}>{log.log_line}</span>
                            </div>
                          ))}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  )

  const renderImagesTab = () => (
    <div className={styles.imagesContent}>
      <div className={styles.imagesList}>
        {data.images.map((image) => (
          <div key={image.id} className={styles.imageCard}>
            <div className={styles.imageHeader}>
              <Image className={styles.imageIcon} />
              <div className={styles.imageInfo}>
                <h3 className={styles.imageTags}>{image.tags.join(', ')}</h3>
                <span className={styles.imageId}>{image.image_id}</span>
              </div>
            </div>
            <div className={styles.imageDetails}>
              <div className={styles.imageDetail}>
                <span className={styles.imageDetailLabel}>Размер:</span>
                <span className={styles.imageDetailValue}>{formatBytes(image.size)}</span>
              </div>
              <div className={styles.imageDetail}>
                <span className={styles.imageDetailLabel}>Создан:</span>
                <span className={styles.imageDetailValue}>{new Date(image.created).toLocaleString('ru-RU')}</span>
              </div>
              <div className={styles.imageDetail}>
                <span className={styles.imageDetailLabel}>Архитектура:</span>
                <span className={styles.imageDetailValue}>{image.architecture}</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )

  const renderVolumesTab = () => (
    <div className={styles.volumesContent}>
      <div className={styles.volumesList}>
        {data.volumes.map((volume) => (
          <div key={volume.id} className={styles.volumeCard}>
            <div className={styles.volumeHeader}>
              <HardDrive className={styles.volumeIcon} />
              <h3 className={styles.volumeName}>{volume.name}</h3>
            </div>
            <div className={styles.volumeDetails}>
              <div className={styles.volumeDetail}>
                <span className={styles.volumeDetailLabel}>Драйвер:</span>
                <span className={styles.volumeDetailValue}>{volume.driver}</span>
              </div>
              <div className={styles.volumeDetail}>
                <span className={styles.volumeDetailLabel}>Mountpoint:</span>
                <span className={styles.volumeDetailValue}>{volume.mountpoint}</span>
              </div>
              <div className={styles.volumeDetail}>
                <span className={styles.volumeDetailLabel}>Создан:</span>
                <span className={styles.volumeDetailValue}>{new Date(volume.created).toLocaleString('ru-RU')}</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )

  const renderNetworksTab = () => (
    <div className={styles.networksContent}>
      <div className={styles.networksList}>
        {data.networks.map((network) => (
          <div key={network.id} className={styles.networkCard}>
            <div className={styles.networkHeader}>
              <Network className={styles.networkIcon} />
              <h3 className={styles.networkName}>{network.name}</h3>
            </div>
            <div className={styles.networkDetails}>
              <div className={styles.networkDetail}>
                <span className={styles.networkDetailLabel}>Драйвер:</span>
                <span className={styles.networkDetailValue}>{network.driver}</span>
              </div>
              <div className={styles.networkDetail}>
                <span className={styles.networkDetailLabel}>Область:</span>
                <span className={styles.networkDetailValue}>{network.scope}</span>
              </div>
              <div className={styles.networkDetail}>
                <span className={styles.networkDetailLabel}>Подсеть:</span>
                <span className={styles.networkDetailValue}>{network.subnet || 'N/A'}</span>
              </div>
              <div className={styles.networkDetail}>
                <span className={styles.networkDetailLabel}>Шлюз:</span>
                <span className={styles.networkDetailValue}>{network.gateway || 'N/A'}</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )

  const renderTabContent = () => {
    switch (activeTab) {
      case 'overview':
        return renderOverviewTab()
      case 'containers':
        return renderContainersTab()
      case 'images':
        return renderImagesTab()
      case 'volumes':
        return renderVolumesTab()
      case 'networks':
        return renderNetworksTab()
      default:
        return renderOverviewTab()
    }
  }

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
          {tabs.map((tab) => {
            const Icon = tab.icon
            return (
              <button
                key={tab.id}
                className={`${styles.tab} ${activeTab === tab.id ? styles.active : ''}`}
                onClick={() => setActiveTab(tab.id as TabType)}
              >
                <Icon className={styles.tabIcon} />
                {tab.label}
              </button>
            )
          })}
        </div>

        <div className={styles.tabContent}>
          {renderTabContent()}
        </div>
      </div>
    </div>
  )
}