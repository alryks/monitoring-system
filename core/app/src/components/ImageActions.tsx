import { useState } from 'react'
import { Trash2 } from 'lucide-react'
import { actionsApi } from '../services/api'
import styles from './ImageActions.module.css'

interface ImageActionsProps {
  imageId: string
  agentId: string
  imageName: string
  onActionComplete: () => void
}

export default function ImageActions({ 
  imageId, 
  agentId, 
  imageName,
  onActionComplete 
}: ImageActionsProps) {
  const [loading, setLoading] = useState<string | null>(null)

  const handleAction = async (actionType: string) => {
    setLoading(actionType)
    try {
      let payload: any = {}

      if (actionType === 'remove_image') {
        payload = {
          image_id: imageId
        }
      } else if (actionType === 'pull_image') {
        // Для pull_image нужно извлечь имя и тег из imageName
        const [name, tag = 'latest'] = imageName.split(':')
        payload = {
          image: name,
          tag: tag
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

  return (
    <div className={styles.container}>
      <button
        onClick={() => handleAction('remove_image')}
        disabled={loading === 'remove_image'}
        className={`${styles.actionButton} ${styles.removeButton}`}
        title="Удалить образ"
      >
        <Trash2 className={styles.icon} />
        {loading === 'remove_image' ? 'Удаление...' : 'Удалить'}
      </button>
    </div>
  )
} 