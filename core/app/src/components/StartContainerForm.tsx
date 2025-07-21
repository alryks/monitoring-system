import { useState } from 'react'
import { Play, X } from 'lucide-react'
import { actionsApi } from '../services/api'
import styles from './StartContainerForm.module.css'

interface StartContainerFormProps {
  agentId: string
  imageName: string
  onContainerStarted: () => void
}

export default function StartContainerForm({ 
  agentId, 
  imageName, 
  onContainerStarted 
}: StartContainerFormProps) {
  const [showForm, setShowForm] = useState(false)
  const [loading, setLoading] = useState(false)
  const [containerName, setContainerName] = useState('')
  const [port, setPort] = useState('')
  const [environment, setEnvironment] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!containerName.trim()) return

    setLoading(true)
    try {
      const payload: any = {
        image: imageName,
        name: containerName.trim()
      }

      // Добавляем порт, если указан
      if (port.trim()) {
        payload.ports = {
          [`${port}/tcp`]: port
        }
      }

      // Добавляем переменные окружения, если указаны
      if (environment.trim()) {
        try {
          const envVars = JSON.parse(environment)
          payload.environment = envVars
        } catch (error) {
          // Если JSON невалидный, игнорируем переменные окружения
          console.warn('Invalid environment variables JSON:', environment)
        }
      }

      await actionsApi.create({
        agent_id: agentId,
        type: 'start_container',
        payload
      })

      setContainerName('')
      setPort('')
      setEnvironment('')
      setShowForm(false)
      onContainerStarted()
    } catch (error) {
      console.error('Error starting container:', error)
      alert('Ошибка при запуске контейнера')
    } finally {
      setLoading(false)
    }
  }

  if (!showForm) {
    return (
      <button
        onClick={() => setShowForm(true)}
        className={styles.startButton}
        title="Запустить контейнер из этого образа"
      >
        <Play className={styles.icon} />
        Запустить
      </button>
    )
  }

  return (
    <div className={styles.formContainer}>
      <div className={styles.formHeader}>
        <h3>Запустить контейнер</h3>
        <button
          onClick={() => setShowForm(false)}
          className={styles.closeButton}
          title="Закрыть"
        >
          <X className={styles.closeIcon} />
        </button>
      </div>

      <div className={styles.imageInfo}>
        <strong>Образ:</strong> {imageName}
      </div>

      <form onSubmit={handleSubmit} className={styles.form}>
        <div className={styles.formGroup}>
          <label htmlFor="containerName">Имя контейнера:</label>
          <input
            id="containerName"
            type="text"
            value={containerName}
            onChange={(e) => setContainerName(e.target.value)}
            placeholder="my-container"
            required
            className={styles.input}
          />
        </div>

        <div className={styles.formGroup}>
          <label htmlFor="port">Порт (опционально):</label>
          <input
            id="port"
            type="text"
            value={port}
            onChange={(e) => setPort(e.target.value)}
            placeholder="8080"
            className={styles.input}
          />
        </div>

        <div className={styles.formGroup}>
          <label htmlFor="environment">Переменные окружения (JSON, опционально):</label>
          <textarea
            id="environment"
            value={environment}
            onChange={(e) => setEnvironment(e.target.value)}
            placeholder='{"NODE_ENV": "production", "PORT": "3000"}'
            className={styles.textarea}
            rows={3}
          />
        </div>

        <div className={styles.formActions}>
          <button
            type="submit"
            disabled={loading || !containerName.trim()}
            className={styles.submitButton}
          >
            <Play className={styles.icon} />
            {loading ? 'Запуск...' : 'Запустить'}
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
  )
} 