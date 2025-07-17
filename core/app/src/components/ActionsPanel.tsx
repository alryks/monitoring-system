import { useState } from 'react'
import { 
  Play, 
  Square, 
  Trash2, 
  RotateCcw, 
  FileText, 
  Settings,
  CheckCircle,
  XCircle,
  Clock
} from 'lucide-react'
import { actionsApi, type Action, type CreateActionRequest } from '../services/api'
import styles from './ActionsPanel.module.css'

interface ActionsPanelProps {
  agentId: string
  onActionCreated?: () => void
}

export default function ActionsPanel({ agentId, onActionCreated }: ActionsPanelProps) {
  const [loading, setLoading] = useState(false)
  const [actions, setActions] = useState<Action[]>([])
  const [showForm, setShowForm] = useState(false)
  const [actionType, setActionType] = useState('')
  const [formData, setFormData] = useState<Record<string, any>>({})

  const actionTypes = [
    { 
      type: 'start_container', 
      label: 'Запустить контейнер', 
      icon: Play,
      fields: [
        { name: 'image', label: 'Образ', type: 'text', required: true },
        { name: 'name', label: 'Имя контейнера', type: 'text', required: false },
        { name: 'ports', label: 'Порты (JSON)', type: 'textarea', required: false },
        { name: 'environment', label: 'Переменные окружения (JSON)', type: 'textarea', required: false },
        { name: 'command', label: 'Команда (JSON массив)', type: 'textarea', required: false },
      ]
    },
    { 
      type: 'stop_container', 
      label: 'Остановить контейнер', 
      icon: Square,
      fields: [
        { name: 'container_id', label: 'ID контейнера', type: 'text', required: true },
        { name: 'timeout', label: 'Таймаут (сек)', type: 'number', required: false },
      ]
    },
    { 
      type: 'remove_container', 
      label: 'Удалить контейнер', 
      icon: Trash2,
      fields: [
        { name: 'container_id', label: 'ID контейнера', type: 'text', required: true },
        { name: 'force', label: 'Принудительно', type: 'checkbox', required: false },
      ]
    },
    { 
      type: 'remove_image', 
      label: 'Удалить образ', 
      icon: Trash2,
      fields: [
        { name: 'image_id', label: 'ID образа', type: 'text', required: true },
        { name: 'force', label: 'Принудительно', type: 'checkbox', required: false },
      ]
    },
    { 
      type: 'restart_nginx', 
      label: 'Перезапустить nginx', 
      icon: RotateCcw,
      fields: []
    },
    { 
      type: 'write_file', 
      label: 'Записать файл', 
      icon: FileText,
      fields: [
        { name: 'path', label: 'Путь к файлу', type: 'text', required: true },
        { name: 'content', label: 'Содержимое', type: 'textarea', required: true },
      ]
    },
  ]

  const loadActions = async () => {
    try {
      const response = await actionsApi.list({ agent_id: agentId })
      setActions(response.data.actions)
    } catch (error) {
      console.error('Error loading actions:', error)
    }
  }

  const createAction = async (data: CreateActionRequest) => {
    setLoading(true)
    try {
      await actionsApi.create(data)
      setShowForm(false)
      setActionType('')
      setFormData({})
      onActionCreated?.()
      loadActions()
    } catch (error) {
      console.error('Error creating action:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    
    const selectedAction = actionTypes.find(a => a.type === actionType)
    if (!selectedAction) return

    // Валидация обязательных полей
    const requiredFields = selectedAction.fields.filter(f => f.required)
    for (const field of requiredFields) {
      if (!formData[field.name]) {
        alert(`Поле "${field.label}" обязательно`)
        return
      }
    }

    // Парсим JSON поля
    const payload = { ...formData }
    for (const field of selectedAction.fields) {
      if (field.type === 'textarea' && payload[field.name]) {
        try {
          payload[field.name] = JSON.parse(payload[field.name])
        } catch (error) {
          alert(`Поле "${field.label}" должно быть валидным JSON`)
          return
        }
      }
    }

    createAction({
      agent_id: agentId,
      type: actionType,
      payload
    })
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className={styles.statusIcon} />
      case 'failed':
        return <XCircle className={styles.statusIcon} />
      default:
        return <Clock className={styles.statusIcon} />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return styles.completed
      case 'failed':
        return styles.failed
      default:
        return styles.pending
    }
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h3>Действия</h3>
        <button 
          className={styles.addButton}
          onClick={() => setShowForm(!showForm)}
        >
          <Settings className={styles.addIcon} />
          Добавить действие
        </button>
      </div>

      {showForm && (
        <div className={styles.form}>
          <h4>Создать действие</h4>
          <form onSubmit={handleSubmit}>
            <div className={styles.formGroup}>
              <label>Тип действия:</label>
              <select 
                value={actionType} 
                onChange={(e) => setActionType(e.target.value)}
                required
              >
                <option value="">Выберите тип действия</option>
                {actionTypes.map(action => (
                  <option key={action.type} value={action.type}>
                    {action.label}
                  </option>
                ))}
              </select>
            </div>

            {actionType && (
              <div className={styles.fields}>
                {actionTypes.find(a => a.type === actionType)?.fields.map(field => (
                  <div key={field.name} className={styles.formGroup}>
                    <label>{field.label}:</label>
                    {field.type === 'textarea' ? (
                      <textarea
                        value={formData[field.name] || ''}
                        onChange={(e) => setFormData({
                          ...formData,
                          [field.name]: e.target.value
                        })}
                        required={field.required}
                        placeholder={field.type === 'textarea' ? '{"key": "value"}' : ''}
                      />
                    ) : field.type === 'checkbox' ? (
                      <input
                        type="checkbox"
                        checked={formData[field.name] || false}
                        onChange={(e) => setFormData({
                          ...formData,
                          [field.name]: e.target.checked
                        })}
                      />
                    ) : (
                      <input
                        type={field.type}
                        value={formData[field.name] || ''}
                        onChange={(e) => setFormData({
                          ...formData,
                          [field.name]: e.target.value
                        })}
                        required={field.required}
                      />
                    )}
                  </div>
                ))}
              </div>
            )}

            <div className={styles.formActions}>
              <button 
                type="submit" 
                disabled={loading}
                className={styles.submitButton}
              >
                {loading ? 'Создание...' : 'Создать'}
              </button>
              <button 
                type="button" 
                onClick={() => setShowForm(false)}
                className={styles.cancelButton}
              >
                Отмена
              </button>
            </div>
          </form>
        </div>
      )}

      <div className={styles.actionsList}>
        <h4>История действий</h4>
        {actions.map(action => (
          <div key={action.id} className={`${styles.actionItem} ${getStatusColor(action.status)}`}>
            <div className={styles.actionHeader}>
              <div className={styles.actionInfo}>
                {getStatusIcon(action.status)}
                <span className={styles.actionType}>
                  {actionTypes.find(a => a.type === action.type)?.label || action.type}
                </span>
                <span className={styles.actionDate}>
                  {new Date(action.created).toLocaleString('ru-RU')}
                </span>
              </div>
              <span className={styles.actionStatus}>{action.status}</span>
            </div>
            {action.response && (
              <div className={styles.actionResponse}>
                <strong>Ответ:</strong> {action.response}
              </div>
            )}
            {action.error && (
              <div className={styles.actionError}>
                <strong>Ошибка:</strong> {action.error}
              </div>
            )}
          </div>
        ))}
        {actions.length === 0 && (
          <div className={styles.emptyState}>
            Нет выполненных действий
          </div>
        )}
      </div>
    </div>
  )
}
