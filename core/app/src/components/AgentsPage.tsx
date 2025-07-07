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
  description: string
}

interface CreateNodeResponse {
  id: number
  name: string
  description: string
  api_key: string
  created_at: string
}

function AgentsPage() {
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [createLoading, setCreateLoading] = useState(false)
  const [newNode, setNewNode] = useState({ name: '', description: '' })
  const [createdNode, setCreatedNode] = useState<CreateNodeResponse | null>(null)

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

  const createNode = async () => {
    if (!newNode.name.trim()) {
      setError('Название узла обязательно')
      return
    }

    setCreateLoading(true)
    setError(null)

    try {
      const response = await axios.post<CreateNodeResponse>('/api/agents/create', {
        name: newNode.name.trim(),
        description: newNode.description.trim() || `Узел ${newNode.name.trim()}`
      })
      
      setCreatedNode(response.data)
      setNewNode({ name: '', description: '' })
      setShowCreateForm(false)
      fetchAgents()
    } catch (err) {
      if (axios.isAxiosError(err) && err.response?.status === 409) {
        setError('Узел с таким названием уже существует')
      } else {
        setError(err instanceof Error ? err.message : 'Ошибка создания узла')
      }
    } finally {
      setCreateLoading(false)
    }
  }

  useEffect(() => {
    fetchAgents()
    
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
        <h1>Узлы системы</h1>
        <p>Управление узлами мониторинга</p>
        <div className="header-actions">
          <button onClick={fetchAgents} disabled={loading} className="refresh-btn">
            {loading ? 'Обновление...' : 'Обновить'}
          </button>
          <button 
            onClick={() => setShowCreateForm(true)} 
            className="create-btn"
            disabled={showCreateForm}
          >
            Создать узел
          </button>
        </div>
      </div>

      {error && (
        <div className="error">
          <p>Ошибка: {error}</p>
          <button onClick={() => setError(null)} className="close-error">×</button>
        </div>
      )}

      {/* Форма создания узла */}
      {showCreateForm && (
        <div className="create-form">
          <h3>Создание нового узла</h3>
          <div className="form-group">
            <label>Название узла:</label>
            <input
              type="text"
              value={newNode.name}
              onChange={(e) => setNewNode({ ...newNode, name: e.target.value })}
              placeholder="Например: server-1"
              disabled={createLoading}
            />
          </div>
          <div className="form-group">
            <label>Описание (опционально):</label>
            <input
              type="text"
              value={newNode.description}
              onChange={(e) => setNewNode({ ...newNode, description: e.target.value })}
              placeholder="Описание назначения узла"
              disabled={createLoading}
            />
          </div>
          <div className="form-actions">
            <button 
              onClick={createNode} 
              disabled={createLoading || !newNode.name.trim()}
              className="create-btn"
            >
              {createLoading ? 'Создание...' : 'Создать'}
            </button>
            <button 
              onClick={() => {
                setShowCreateForm(false)
                setNewNode({ name: '', description: '' })
              }}
              disabled={createLoading}
              className="cancel-btn"
            >
              Отмена
            </button>
          </div>
        </div>
      )}

      {/* Показ созданного узла с учетными данными */}
      {createdNode && (
        <div className="node-created">
          <h3>✅ Узел успешно создан!</h3>
          <div className="credentials">
            <h4>Данные для настройки агента:</h4>
            <div className="cred-item">
              <label>ID узла:</label>
              <code>{createdNode.id}</code>
            </div>
            <div className="cred-item">
              <label>API Ключ:</label>
              <code className="api-key">{createdNode.api_key}</code>
            </div>
            <div className="instructions">
              <p><strong>Инструкция:</strong></p>
              <p>1. Скопируйте эти данные</p>
              <p>2. Создайте файл .env на сервере агента</p>
              <p>3. Добавьте в .env:</p>
              <pre>{`AGENT_ID=${createdNode.id}
API_KEY=${createdNode.api_key}
CORE_API_URL=http://ваш-core-сервер/api`}</pre>
            </div>
          </div>
          <button onClick={() => setCreatedNode(null)} className="close-btn">
            Закрыть
          </button>
        </div>
      )}

      {loading && agents.length === 0 ? (
        <div className="loading">
          <p>Загрузка узлов...</p>
        </div>
      ) : (
        <div className="agents-grid">
          {agents.length === 0 ? (
            <div className="no-agents">
              <p>Узлы не найдены</p>
              <p>Создайте первый узел с помощью кнопки "Создать узел"</p>
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
                  
                  {agent.description && (
                    <div className="detail-row">
                      <span className="label">Описание:</span>
                      <span className="value">{agent.description}</span>
                    </div>
                  )}
                  
                  <div className="detail-row">
                    <span className="label">Последняя активность:</span>
                    <span className="value">
                      {getTimeSince(agent.last_seen_at)}
                    </span>
                  </div>
                  
                  <div className="detail-row">
                    <span className="label">Создан:</span>
                    <span className="value">{formatDate(agent.created_at)}</span>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      )}
      
      <div className="page-footer">
        <p>Узлы с последней активностью более 30 секунд назад считаются оффлайн</p>
        <p>Обновление каждые 5 секунд</p>
      </div>
    </div>
  )
}

export default AgentsPage