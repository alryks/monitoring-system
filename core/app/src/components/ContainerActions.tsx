import { useState } from 'react'
import { Play, Square, RotateCcw, Trash2 } from 'lucide-react'
import { actionsApi } from '../services/api'
import styles from './ContainerActions.module.css'

interface ContainerActionsProps {
  containerId: string
  agentId: string
  status: string
  onActionComplete: () => void
}

export default function ContainerActions({ 
  containerId, 
  agentId, 
  status, 
  onActionComplete 
}: ContainerActionsProps) {
  const [loading, setLoading] = useState<string | null>(null)

  const handleAction = async (actionType: string) => {
    setLoading(actionType)
    try {
      let payload: any = {}

      if (actionType === 'start_container') {
        // Для запуска существующего контейнера передаем container_id
        payload = {
          container_id: containerId
        }
      } else {
        // Для других действий также передаем container_id
        payload = {
          container_id: containerId
        }
      }

      await actionsApi.create({
        agent_id: agentId,
        type: actionType,
        payload
      })

      onActionComplete()
    } catch (error) {
      console.error(`Error executing ${actionType}:`, error)
      alert(`Ошибка при выполнении действия: ${actionType}`)
    } finally {
      setLoading(null)
    }
  }

  const isRunning = status.toLowerCase().includes('running') || status.toLowerCase().includes('up')

  return (
    <div className={styles.container}>
      {!isRunning && (
        <button
          onClick={() => handleAction('start_container')}
          disabled={loading === 'start_container'}
          className={`${styles.actionButton} ${styles.startButton}`}
          title="Запустить контейнер"
        >
          <Play className={styles.icon} />
          {loading === 'start_container' ? 'Запуск...' : 'Запустить'}
        </button>
      )}

      {isRunning && (
        <>
          <button
            onClick={() => handleAction('stop_container')}
            disabled={loading === 'stop_container'}
            className={`${styles.actionButton} ${styles.stopButton}`}
            title="Остановить контейнер"
          >
            <Square className={styles.icon} />
            {loading === 'stop_container' ? 'Остановка...' : 'Остановить'}
          </button>

          <button
            onClick={() => handleAction('restart_container')}
            disabled={loading === 'restart_container'}
            className={`${styles.actionButton} ${styles.restartButton}`}
            title="Перезапустить контейнер"
          >
            <RotateCcw className={styles.icon} />
            {loading === 'restart_container' ? 'Перезапуск...' : 'Перезапустить'}
          </button>
        </>
      )}

      <button
        onClick={() => handleAction('remove_container')}
        disabled={loading === 'remove_container'}
        className={`${styles.actionButton} ${styles.removeButton}`}
        title="Удалить контейнер"
      >
        <Trash2 className={styles.icon} />
        {loading === 'remove_container' ? 'Удаление...' : 'Удалить'}
      </button>
    </div>
  )
} 