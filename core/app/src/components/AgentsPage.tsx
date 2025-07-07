import { useState, useEffect } from 'react'
import axios from 'axios'

interface Agent {
  id: number
  name: string
  url: string
  last_seen_at: string
  created_at: string
  updated_at: string
  status: string
}

function AgentsPage() {
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchAgents = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const response = await axios.get<Agent[]>('/api/agents')
      setAgents(response.data || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Произошла ошибка')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchAgents()
    
    // Auto-refresh every 5 seconds
    const interval = setInterval(fetchAgents, 5000)
    return () => clearInterval(interval)
  }, [])

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online':
        return '#4caf50'
      case 'offline':
        return '#f44336'
      default:
        return '#ff9800'
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('ru-RU')
  }

  const getTimeSince = (dateString: string) => {
    const now = new Date()
    const date = new Date(dateString)
    const diffMs = now.getTime() - date.getTime()
    const diffSeconds = Math.floor(diffMs / 1000)
    
    if (diffSeconds < 60) {
      return `${diffSeconds} сек назад`
    }
    
    const diffMinutes = Math.floor(diffSeconds / 60)
    if (diffMinutes < 60) {
      return `${diffMinutes} мин назад`
    }
    
    const diffHours = Math.floor(diffMinutes / 60)
    if (diffHours < 24) {
      return `${diffHours} ч назад`
    }
    
    const diffDays = Math.floor(diffHours / 24)
    return `${diffDays} дн назад`
  }

  return (
    <div className="agents-page">
      <div className="page-header">
        <h1>Агенты системы</h1>
        <p>Список всех зарегистрированных агентов и их статус</p>
        <button onClick={fetchAgents} disabled={loading} className="refresh-btn">
          {loading ? 'Обновление...' : 'Обновить'}
        </button>
      </div>

      {error && (
        <div className="error">
          <p>Ошибка загрузки: {error}</p>
        </div>
      )}

      {loading && agents.length === 0 ? (
        <div className="loading">
          <p>Загрузка агентов...</p>
        </div>
      ) : (
        <div className="agents-grid">
          {agents.length === 0 ? (
            <div className="no-agents">
              <p>Агенты не найдены</p>
              <p>Запустите агент на удаленном сервере для его автоматической регистрации</p>
            </div>
          ) : (
            agents.map((agent) => (
              <div key={agent.id} className="agent-card">
                <div className="agent-header">
                  <h3>{agent.name}</h3>
                  <div 
                    className="status-indicator"
                    style={{ backgroundColor: getStatusColor(agent.status) }}
                  >
                    {agent.status}
                  </div>
                </div>
                
                <div className="agent-details">
                  <div className="detail-row">
                    <span className="label">ID:</span>
                    <span className="value">{agent.id}</span>
                  </div>
                  
                  <div className="detail-row">
                    <span className="label">URL:</span>
                    <span className="value">{agent.url}</span>
                  </div>
                  
                  <div className="detail-row">
                    <span className="label">Последняя активность:</span>
                    <span className="value">
                      {getTimeSince(agent.last_seen_at)}
                    </span>
                  </div>
                  
                  <div className="detail-row">
                    <span className="label">Зарегистрирован:</span>
                    <span className="value">{formatDate(agent.created_at)}</span>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      )}
      
      <div className="page-footer">
        <p>Агенты с последней активностью более 30 секунд назад считаются оффлайн</p>
        <p>Обновление каждые 5 секунд</p>
      </div>
    </div>
  )
}

export default AgentsPage