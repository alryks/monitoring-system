import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { 
  Container, 
  RotateCcw, 
  Search,
  RefreshCw,
  CheckCircle,
  XCircle,
  AlertCircle,
  Server,
  ChevronDown,
  ChevronUp,
  Eye,
  Cpu,
  MemoryStick,
} from 'lucide-react'
import {
  containersApi,
  agentsApi,
  type Container as ContainerType,
  type Agent,
  formatCPUUsage,
  formatBytes,
  formatMemoryUsageMB
} from '../services/api'
import ContainerActions from '../components/ContainerActions'
import styles from './Containers.module.css'

export default function Containers() {
  const navigate = useNavigate()
  const [containers, setContainers] = useState<ContainerType[]>([])
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState('')
  const [searchTerm, setSearchTerm] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [agentFilter, setAgentFilter] = useState<string>('all')
  const [expandedContainer, setExpandedContainer] = useState<string | null>(null)

  const fetchData = async (isRefresh = false) => {
    try {
      if (isRefresh) {
        setRefreshing(true)
      } else {
        setLoading(true)
      }
      
      const [containersResponse, agentsResponse] = await Promise.all([
        containersApi.getAll({ 
          agent_id: agentFilter !== 'all' ? agentFilter : undefined,
          status: statusFilter !== 'all' ? statusFilter : undefined,
          search: searchTerm || undefined
        }),
        agentsApi.getAll()
      ])
      
      // Проверяем, что containersResponse.data является массивом
      const containersData = Array.isArray(containersResponse.data) ? containersResponse.data : []
      const agentsData = Array.isArray(agentsResponse.data) ? agentsResponse.data : []
      
      setContainers(containersData)
      setAgents(agentsData)
      setError('')
    } catch (err: any) {
      setError('Ошибка загрузки контейнеров')
      console.error('Containers fetch error:', err)
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [])

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      fetchData(true)
    }, 500)

    return () => clearTimeout(timeoutId)
  }, [searchTerm, statusFilter, agentFilter])

  const handleContainerToggle = (containerId: string) => {
    setExpandedContainer(expandedContainer === containerId ? null : containerId)
  }

  const handleAgentClick = (agentId: string) => {
    navigate(`/agents/${agentId}`)
  }

  const getStatusIcon = (status: string) => {
    const statusLower = status.toLowerCase()
    if (statusLower.includes('up') || statusLower.includes('running')) {
      return <CheckCircle className={`${styles.statusIcon} ${styles.running}`} />
    } else if (statusLower.includes('exited') || statusLower.includes('stopped')) {
      return <XCircle className={`${styles.statusIcon} ${styles.stopped}`} />
    } else if (statusLower.includes('paused')) {
      return <AlertCircle className={`${styles.statusIcon} ${styles.paused}`} />
    } else if (statusLower.includes('restarting')) {
      return <RotateCcw className={`${styles.statusIcon} ${styles.restarting}`} />
    } else {
      return <AlertCircle className={`${styles.statusIcon} ${styles.unknown}`} />
    }
  }

  const getStatusText = (status: string) => {
    const statusLower = status.toLowerCase()
    if (statusLower.includes('up') || statusLower.includes('running')) {
      return 'Запущен'
    } else if (statusLower.includes('exited')) {
      return 'Завершен'
    } else if (statusLower.includes('stopped')) {
      return 'Остановлен'
    } else if (statusLower.includes('paused')) {
      return 'Приостановлен'
    } else if (statusLower.includes('restarting')) {
      return 'Перезапускается'
    } else {
      return status // Возвращаем оригинальный статус, если не распознали
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  const statusOptions = [
    { value: 'all', label: 'Все статусы' },
    { value: 'running', label: 'Запущенные' },
    { value: 'stopped', label: 'Остановленные' },
    { value: 'exited', label: 'Завершенные' },
    { value: 'paused', label: 'Приостановленные' },
    { value: 'restarting', label: 'Перезапускающиеся' }
  ]

  if (loading && !containers.length) {
    return (
      <div className={styles.loading}>
        <RefreshCw className={`${styles.refreshIcon} ${styles.spinning}`} />
        <span>Загрузка контейнеров...</span>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <div className={styles.titleSection}>
          <h1 className={styles.title}>
            <Container className={styles.titleIcon} />
            Контейнеры
          </h1>
          <span className={styles.subtitle}>
            Всего: {containers.length}
          </span>
        </div>
        
        <div className={styles.actions}>
          <button 
            className={styles.refreshButton}
            onClick={() => fetchData(true)}
            disabled={refreshing}
          >
            <RefreshCw className={`${styles.refreshIcon} ${refreshing ? styles.spinning : ''}`} />
            Обновить
          </button>
        </div>
      </div>

      {error && (
        <div className={styles.error}>
          <span>{error}</span>
        </div>
      )}

      <div className={styles.filters}>
        <div className={styles.searchBox}>
          <Search className={styles.searchIcon} />
          <input
            type="text"
            placeholder="Поиск по имени или образу..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className={styles.searchInput}
          />
        </div>

        <div className={styles.filterGroup}>
          <div className={styles.filterItem}>
            <label className={styles.filterLabel}>Статус:</label>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className={styles.filterSelect}
            >
              {statusOptions.map(option => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>

          <div className={styles.filterItem}>
            <label className={styles.filterLabel}>Агент:</label>
            <select
              value={agentFilter}
              onChange={(e) => setAgentFilter(e.target.value)}
              className={styles.filterSelect}
            >
              <option value="all">Все агенты</option>
              {agents.map(agent => (
                <option key={agent.id} value={agent.id}>
                  {agent.name}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      <div className={styles.containersList}>
        {containers.length === 0 ? (
          <div className={styles.noData}>
            <Container className={styles.noDataIcon} />
            <span>Контейнеры не найдены</span>
          </div>
        ) : (
          containers.map((container) => (
            <div key={container.id} className={styles.containerCard}>
              <div 
                className={styles.containerHeader}
                onClick={() => handleContainerToggle(container.id)}
              >
                <div className={styles.containerInfo}>
                  <div className={styles.statusSection}>
                    {getStatusIcon(container.status)}
                    <div className={styles.containerBasicInfo}>
                      <h3 className={styles.containerName}>{container.name}</h3>
                      <span className={styles.containerStatus}>
                        {getStatusText(container.status)}
                      </span>
                    </div>
                  </div>
                  
                  <div className={styles.containerDetails}>
                    <div className={styles.detailItem}>
                      <span className={styles.detailLabel}>Образ:</span>
                      <span className={styles.detailValue}>{container.image_id}</span>
                    </div>
                    <div className={styles.detailItem}>
                      <span className={styles.detailLabel}>Агент:</span>
                      <button 
                        className={styles.agentLink}
                        onClick={(e) => {
                          e.stopPropagation()
                          if (container.agent_id) {
                            handleAgentClick(container.agent_id)
                          }
                        }}
                      >
                        <Server className={styles.agentIcon} />
                        {container.agent_name || 'Неизвестен'}
                      </button>
                    </div>
                    <div className={styles.detailItem}>
                      <span className={styles.detailLabel}>Создан:</span>
                      <span className={styles.detailValue}>{formatDate(container.created_at)}</span>
                    </div>
                  </div>
                </div>

                <div className={styles.containerMetrics}>
                  <div className={styles.metricItem}>
                    <Cpu className={styles.metricIcon} />
                    <span className={styles.metricValue}>
                      {formatCPUUsage(container.cpu_usage_percent)}
                    </span>
                  </div>
                  <div className={styles.metricItem}>
                    <MemoryStick className={styles.metricIcon} />
                    <span className={styles.metricValue}>
                      {formatMemoryUsageMB(container.memory_usage_mb)}
                    </span>
                  </div>
                  <div className={styles.metricItem}>
                    <RotateCcw className={styles.metricIcon} />
                    <span className={styles.metricValue}>
                      {container.restart_count}
                    </span>
                  </div>
                </div>

                <div className={styles.containerToggle}>
                  {expandedContainer === container.id ? <ChevronUp /> : <ChevronDown />}
                </div>
              </div>

              {expandedContainer === container.id && (
                <div className={styles.containerExpanded}>
                  <div className={styles.expandedContent}>
                    <div className={styles.expandedSection}>
                      <h4 className={styles.expandedTitle}>Подробная информация</h4>
                      <div className={styles.expandedGrid}>
                        <div className={styles.expandedItem}>
                          <span className={styles.expandedLabel}>Container ID:</span>
                          <span className={styles.expandedValue}>{container.container_id}</span>
                        </div>
                        <div className={styles.expandedItem}>
                          <span className={styles.expandedLabel}>IP адрес:</span>
                          <span className={styles.expandedValue}>{container.ip_address || 'N/A'}</span>
                        </div>
                        <div className={styles.expandedItem}>
                          <span className={styles.expandedLabel}>MAC адрес:</span>
                          <span className={styles.expandedValue}>{container.mac_address || 'N/A'}</span>
                        </div>
                        <div className={styles.expandedItem}>
                          <span className={styles.expandedLabel}>Сеть отправлено:</span>
                          <span className={styles.expandedValue}>
                            {container.network_sent_bytes ? formatBytes(container.network_sent_bytes) : 'N/A'}
                          </span>
                        </div>
                        <div className={styles.expandedItem}>
                          <span className={styles.expandedLabel}>Сеть получено:</span>
                          <span className={styles.expandedValue}>
                            {container.network_received_bytes ? formatBytes(container.network_received_bytes) : 'N/A'}
                          </span>
                        </div>
                        <div className={styles.expandedItem}>
                          <span className={styles.expandedLabel}>Количество перезапусков:</span>
                          <span className={styles.expandedValue}>{container.restart_count}</span>
                        </div>
                      </div>
                    </div>

                    <div className={styles.expandedActions}>
                      <button 
                        className={styles.detailButton}
                        onClick={(e) => {
                          e.stopPropagation()
                          navigate(`/containers/${container.id}`)
                        }}
                      >
                        <Eye className={styles.buttonIcon} />
                        Подробнее
                      </button>
                      <button 
                        className={styles.agentButton}
                        onClick={(e) => {
                          e.stopPropagation()
                          if (container.agent_id) {
                            handleAgentClick(container.agent_id)
                          }
                        }}
                      >
                        <Server className={styles.buttonIcon} />
                        Перейти к агенту
                      </button>
                    </div>

                    {container.agent_id && (
                      <div className={styles.containerActionsSection}>
                        <h4 className={styles.expandedTitle}>Действия</h4>
                        <ContainerActions
                          containerId={container.container_id}
                          agentId={container.agent_id}
                          status={container.status}
                          onActionComplete={fetchData}
                        />
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          ))
        )}
      </div>

      {containers.length > 0 && (
        <div className={styles.summary}>
          <div className={styles.summaryStats}>
            <div className={styles.statItem}>
              <CheckCircle className={styles.statIcon} />
              <span className={styles.statLabel}>Запущено:</span>
              <span className={styles.statValue}>
                {containers.filter(c => c.status.toLowerCase().includes("up") || c.status.toLowerCase().includes("running")).length}
              </span>
            </div>
            <div className={styles.statItem}>
              <XCircle className={styles.statIcon} />
              <span className={styles.statLabel}>Остановлено:</span>
              <span className={styles.statValue}>
                {containers.filter(c => c.status.toLowerCase().includes("stopped") || c.status.toLowerCase().includes("exited")).length}
              </span>
            </div>
            <div className={styles.statItem}>
              <AlertCircle className={styles.statIcon} />
              <span className={styles.statLabel}>Другие:</span>
              <span className={styles.statValue}>
                {containers.filter(c => !c.status.toLowerCase().includes("up") && !c.status.toLowerCase().includes("running") && !c.status.toLowerCase().includes("stopped") && !c.status.toLowerCase().includes("exited")).length}
              </span>
            </div>
          </div>
        </div>
      )}
    </div>
  )
} 