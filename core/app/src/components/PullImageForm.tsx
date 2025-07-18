import { useState } from 'react'
import { Download, X } from 'lucide-react'
import { actionsApi } from '../services/api'
import styles from './PullImageForm.module.css'

interface PullImageFormProps {
  agentId: string
  onImagePulled: () => void
}

export default function PullImageForm({ agentId, onImagePulled }: PullImageFormProps) {
  const [showForm, setShowForm] = useState(false)
  const [loading, setLoading] = useState(false)
  const [imageName, setImageName] = useState('')
  const [imageTag, setImageTag] = useState('latest')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!imageName.trim()) return

    setLoading(true)
    try {
      await actionsApi.create({
        agent_id: agentId,
        type: 'pull_image',
        payload: {
          image: imageName.trim(),
          tag: imageTag.trim() || 'latest'
        }
      })

      setImageName('')
      setImageTag('latest')
      setShowForm(false)
      onImagePulled()
    } catch (error) {
      console.error('Error pulling image:', error)
      alert('Ошибка при загрузке образа')
    } finally {
      setLoading(false)
    }
  }

  if (!showForm) {
    return (
      <button
        onClick={() => setShowForm(true)}
        className={styles.pullButton}
        title="Загрузить новый образ"
      >
        <Download className={styles.icon} />
        Загрузить образ
      </button>
    )
  }

  return (
    <div className={styles.formContainer}>
      <div className={styles.formHeader}>
        <h3>Загрузить новый образ</h3>
        <button
          onClick={() => setShowForm(false)}
          className={styles.closeButton}
          title="Закрыть"
        >
          <X className={styles.closeIcon} />
        </button>
      </div>

      <form onSubmit={handleSubmit} className={styles.form}>
        <div className={styles.formGroup}>
          <label htmlFor="imageName">Имя образа:</label>
          <input
            id="imageName"
            type="text"
            value={imageName}
            onChange={(e) => setImageName(e.target.value)}
            placeholder="например: nginx, ubuntu, node"
            required
            className={styles.input}
          />
        </div>

        <div className={styles.formGroup}>
          <label htmlFor="imageTag">Тег:</label>
          <input
            id="imageTag"
            type="text"
            value={imageTag}
            onChange={(e) => setImageTag(e.target.value)}
            placeholder="latest"
            className={styles.input}
          />
        </div>

        <div className={styles.formActions}>
          <button
            type="submit"
            disabled={loading || !imageName.trim()}
            className={styles.submitButton}
          >
            <Download className={styles.icon} />
            {loading ? 'Загрузка...' : 'Загрузить'}
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