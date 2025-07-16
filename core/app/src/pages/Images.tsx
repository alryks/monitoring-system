import { useState, useEffect } from 'react'
import { Package, Search, RefreshCw, Server } from 'lucide-react'
import styles from './Images.module.css'
import { imagesApi, agentsApi } from '../services/api'
import type { ImageDetail, Agent } from '../services/api'

export default function Images() {
  const [images, setImages] = useState<ImageDetail[]>([])
  const [agents, setAgents] = useState<Agent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [searchTerm, setSearchTerm] = useState('')
  const [agentFilter, setAgentFilter] = useState<string>('all')

  const fetchData = async () => {
    try {
      setLoading(true)
      setError('')
      
      // Загружаем образы и агентов параллельно
      const [imagesResponse, agentsResponse] = await Promise.all([
        imagesApi.getAll(),
        agentsApi.getAll()
      ])
      
      setImages(imagesResponse.data as ImageDetail[])
      setAgents(agentsResponse.data)
    } catch (err) {
      console.error('Error fetching data:', err)
      setError('Ошибка загрузки данных')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [])

  const filteredImages = images.filter(image => {
    const matchesSearch = image.tags.some(tag => 
      tag.toLowerCase().includes(searchTerm.toLowerCase())
    ) || image.image_id.toLowerCase().includes(searchTerm.toLowerCase())
    
    const matchesAgent = agentFilter === 'all' || image.agent?.id === agentFilter
    
    return matchesSearch && matchesAgent
  })

  const formatSize = (bytes: number) => {
    if (bytes >= 1024 * 1024 * 1024) {
      return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`
    } else if (bytes >= 1024 * 1024) {
      return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
    } else if (bytes >= 1024) {
      return `${(bytes / 1024).toFixed(1)} KB`
    }
    return `${bytes} B`
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

  if (loading) {
    return (
      <div className={styles.loading}>
        <RefreshCw className={`${styles.refreshIcon} ${styles.spinning}`} />
        <span>Загрузка образов...</span>
      </div>
    )
  }

  if (error) {
    return (
      <div className={styles.error}>
        <span>{error}</span>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <div className={styles.titleSection}>
          <h1 className={styles.title}>
            <Package className={styles.titleIcon} />
            Docker Образы
          </h1>
          <span className={styles.subtitle}>
            Всего: {filteredImages.length}
          </span>
        </div>
        
        <div className={styles.actions}>
          <button 
            className={styles.refreshButton}
            onClick={() => fetchData()}
            disabled={loading}
          >
            <RefreshCw className={`${styles.refreshIcon} ${loading ? styles.spinning : ''}`} />
            Обновить
          </button>
        </div>
      </div>

      <div className={styles.filters}>
        <div className={styles.searchBox}>
          <Search className={styles.searchIcon} />
          <input
            type="text"
            placeholder="Поиск по тегу или ID образа..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className={styles.searchInput}
          />
        </div>

        <div className={styles.filterGroup}>
          <Server className={styles.filterIcon} />
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

      <div className={styles.imagesList}>
        {filteredImages.map((image) => (
          <div key={image.id} className={styles.imageCard}>
            <div className={styles.imageHeader}>
              <div className={styles.imageId}>
                <span className={styles.id}>{image.image_id.slice(0, 12)}</span>
              </div>
              <div className={styles.imageSize}>
                {formatSize(image.size)}
              </div>
            </div>

            <div className={styles.imageTags}>
              {image.tags.map((tag, index) => (
                <span key={index} className={styles.tag}>
                  {tag}
                </span>
              ))}
            </div>

            <div className={styles.imageInfo}>
              <div className={styles.infoRow}>
                <span className={styles.label}>Агент:</span>
                <span className={styles.value}>{image.agent?.name || 'Неизвестно'}</span>
              </div>
              <div className={styles.infoRow}>
                <span className={styles.label}>Создан:</span>
                <span className={styles.value}>{formatDate(image.created)}</span>
              </div>
            </div>
          </div>
        ))}
      </div>

      {filteredImages.length === 0 && (
        <div className={styles.empty}>
          <Package size={48} />
          <p>Образы не найдены</p>
          <p>Попробуйте изменить фильтры поиска</p>
        </div>
      )}
    </div>
  )
} 