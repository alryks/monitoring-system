import { useState, useEffect } from 'react'
import { 
  Plus, 
  Trash2, 
  Eye,
  CheckCircle, 
  XCircle, 
  Copy,
  Check,
  X,
  Server,
  RefreshCw
} from 'lucide-react'
import { agentsApi, type Agent } from '../services/api'
import styles from './Agents.module.css'

export default function Agents() {
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showTokenModal, setShowTokenModal] = useState(false)
  const [selectedAgent, setSelectedAgent] = useState<Agent | null>(null)
  const [newAgentName, setNewAgentName] = useState('')
  const [copiedToken, setCopiedToken] = useState(false)

  const fetchAgents = async () => {
    try {
      setLoading(true)
      const response = await agentsApi.getAll()
      // Ensure we always have an array, even if the API returns null
      setAgents(Array.isArray(response.data) ? response.data : [])
      setError('')
    } catch (error: any) {
      setError('Ошибка загрузки агентов')
      console.error('Agents fetch error:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchAgents()
    const interval = setInterval(fetchAgents, 30000)
    return () => clearInterval(interval)
  }, [])

  const handleCreateAgent = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newAgentName.trim()) return

    try {
      const response = await agentsApi.create(newAgentName.trim())
      // Ensure we have a valid array before spreading
      const currentAgents = Array.isArray(agents) ? agents : []
      setAgents([response.data, ...currentAgents])
      setNewAgentName('')
      setShowCreateModal(false)
      setSelectedAgent(response.data)
      setShowTokenModal(true)
    } catch (error: any) {
      setError('Ошибка создания агента')
    }
  }

  const handleDeleteAgent = async (agent: Agent) => {
    if (!confirm(`Удалить агента "${agent.name}"?`)) return

    try {
      await agentsApi.delete(agent.id)
      // Ensure we have a valid array before filtering
      const currentAgents = Array.isArray(agents) ? agents : []
      setAgents(currentAgents.filter(a => a.id !== agent.id))
    } catch (error: any) {
      setError('Ошибка удаления агента')
    }
  }

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopiedToken(true)
      setTimeout(() => setCopiedToken(false), 2000)
    } catch (error) {
      console.error('Failed to copy to clipboard:', error)
    }
  }

  const handleViewToken = (agent: Agent) => {
    setSelectedAgent(agent)
    setShowTokenModal(true)
  }

  if (loading && (!Array.isArray(agents) || agents.length === 0)) {
    return (
      <div className={styles.loading}>
        <RefreshCw className={`${styles.refreshIcon} ${styles.spinning}`} />
        <span className={styles.loadingText}>Загрузка...</span>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1 className={styles.title}>Агенты</h1>
        <button
          onClick={() => setShowCreateModal(true)}
          className={styles.addButton}
        >
          <Plus className={styles.addIcon} />
          Добавить агента
        </button>
      </div>

      {error && (
        <div className={styles.error}>
          <span className={styles.errorText}>{error}</span>
        </div>
      )}

      {!Array.isArray(agents) || agents.length === 0 ? (
        <div className={styles.empty}>
          <Server size={48} />
          <p>Нет агентов</p>
          <p>Создайте первого агента для начала мониторинга</p>
        </div>
      ) : (
        <div className={styles.agentsList}>
          {(Array.isArray(agents) ? agents : []).map((agent) => (
            <div key={agent.id} className={styles.agentCard}>
              <div className={styles.agentHeader}>
                <div className={styles.agentInfo}>
                  <div className={styles.agentName}>{agent.name}</div>
                  <div className={styles.agentStatus}>
                    {agent.status === 'online' ? (
                      <CheckCircle className={`${styles.statusIcon} ${styles.online}`} />
                    ) : (
                      <XCircle className={`${styles.statusIcon} ${styles.offline}`} />
                    )}
                    <span>{agent.status === 'online' ? 'Онлайн' : 'Оффлайн'}</span>
                  </div>
                </div>
                <div className={styles.agentActions}>
                  <button
                    onClick={() => handleViewToken(agent)}
                    className={`${styles.actionButton} ${styles.view}`}
                    title="Показать токен"
                  >
                    <Eye className={styles.actionIcon} />
                  </button>
                  <button
                    onClick={() => handleDeleteAgent(agent)}
                    className={`${styles.actionButton} ${styles.delete}`}
                    title="Удалить агента"
                  >
                    <Trash2 className={styles.actionIcon} />
                  </button>
                </div>
              </div>
              
              <div className={styles.agentDetails}>
                <div className={styles.detailItem}>
                  <span className={styles.detailLabel}>IP адрес</span>
                  <span className={styles.detailValue}>{agent.public_ip || 'Неизвестен'}</span>
                </div>
                <div className={styles.detailItem}>
                  <span className={styles.detailLabel}>Последний пинг</span>
                  <span className={styles.detailValue}>
                    {agent.last_ping ? 
                      new Date(agent.last_ping).toLocaleString('ru-RU') : 
                      'Нет пингов'
                    }
                  </span>
                </div>
                <div className={styles.detailItem}>
                  <span className={styles.detailLabel}>Создан</span>
                  <span className={styles.detailValue}>
                    {new Date(agent.created).toLocaleDateString('ru-RU')}
                  </span>
                </div>
                <div className={styles.detailItem}>
                  <span className={styles.detailLabel}>Активен</span>
                  <span className={styles.detailValue}>
                    {agent.is_active ? 'Да' : 'Нет'}
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Agent Modal */}
      {showCreateModal && (
        <div className={styles.modal}>
          <div className={styles.modalContent}>
            <div className={styles.modalHeader}>
              <h2 className={styles.modalTitle}>Создать агента</h2>
              <button
                onClick={() => setShowCreateModal(false)}
                className={styles.closeButton}
              >
                <X className={styles.closeIcon} />
              </button>
            </div>
            
            <form onSubmit={handleCreateAgent} className={styles.form}>
              <div className={styles.field}>
                <label htmlFor="agentName" className={styles.label}>
                  Название агента
                </label>
                <input
                  id="agentName"
                  type="text"
                  value={newAgentName}
                  onChange={(e) => setNewAgentName(e.target.value)}
                  className={styles.input}
                  placeholder="Введите название агента"
                  required
                />
              </div>
              
              <div className={styles.formActions}>
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className={`${styles.button} ${styles.secondary}`}
                >
                  Отмена
                </button>
                <button
                  type="submit"
                  className={`${styles.button} ${styles.primary}`}
                  disabled={!newAgentName.trim()}
                >
                  Создать
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Token Modal */}
      {showTokenModal && selectedAgent && (
        <div className={styles.modal}>
          <div className={styles.modalContent}>
            <div className={styles.modalHeader}>
              <h2 className={styles.modalTitle}>Токен агента</h2>
              <button
                onClick={() => setShowTokenModal(false)}
                className={styles.closeButton}
              >
                <X className={styles.closeIcon} />
              </button>
            </div>
            
            <div>
              <p style={{ marginBottom: '1rem', color: '#64748b' }}>
                Используйте этот токен для настройки агента "{selectedAgent.name}"
              </p>
              
              <div className={styles.agentToken}>
                <div className={styles.tokenLabel}>Bearer Token:</div>
                <div className={styles.tokenContainer}>
                  <input
                    type="text"
                    value={selectedAgent.token}
                    readOnly
                    className={styles.tokenValue}
                  />
                  <button
                    onClick={() => copyToClipboard(selectedAgent.token)}
                    className={`${styles.copyButton} ${copiedToken ? styles.copied : ''}`}
                    title="Копировать токен"
                  >
                    {copiedToken ? <Check className={styles.copyIcon} /> : <Copy className={styles.copyIcon} />}
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
} 