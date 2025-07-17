import { useState, useEffect } from 'react'
import { Bell, Bot, AlertTriangle, Cpu, HardDrive, Server, Container } from 'lucide-react'
import { notificationsApi, type NotificationSettings } from '../services/api'

export default function Notifications() {
  const [settings, setSettings] = useState<NotificationSettings>({
    telegram_bot_token: '',
    telegram_chat_id: '',
    notifications: {
      agent_offline: {
        enabled: false,
        message: '🚨 Агент {AGENT_NAME} не отвечает!'
      },
      container_stopped: {
        enabled: false,
        message: '⚠️ Контейнер {CONTAINER_NAME} остановился на агенте {AGENT_NAME}'
      },
      cpu_threshold: {
        enabled: false,
        threshold: 80,
        message: '🔥 Высокое использование CPU: {AGENT_NAME} - {CPU_USAGE}%'
      },
      ram_threshold: {
        enabled: false,
        threshold: 80,
        message: '💾 Высокое использование RAM: {AGENT_NAME} - {RAM_USAGE}%'
      }
    }
  })

  const [isLoading, setIsLoading] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    loadSettings()
  }, [])

  const loadSettings = async () => {
    try {
      setIsLoading(true)
      const response = await notificationsApi.getSettings()
      if (response.data) {
        setSettings(response.data)
      }
    } catch (error) {
      console.error('Ошибка загрузки настроек:', error)
      // При ошибке оставляем дефолтные настройки
    } finally {
      setIsLoading(false)
    }
  }

  const saveSettings = async () => {
    try {
      setIsLoading(true)
      await notificationsApi.updateSettings(settings)
      setMessage('Настройки успешно сохранены')
      setTimeout(() => setMessage(''), 3000)
    } catch (error) {
      setMessage('Ошибка сохранения настроек')
      setTimeout(() => setMessage(''), 3000)
    } finally {
      setIsLoading(false)
    }
  }

  const testNotification = async () => {
    try {
      await notificationsApi.sendTest()
      setMessage('Тестовое уведомление отправлено')
      setTimeout(() => setMessage(''), 3000)
    } catch (error) {
      setMessage('Ошибка отправки тестового уведомления')
      setTimeout(() => setMessage(''), 3000)
    }
  }

  const updateNotification = (type: keyof NotificationSettings['notifications'], field: string, value: any) => {
    setSettings(prev => {
      // Проверяем, что notifications существует
      if (!prev.notifications) {
        return prev
      }
      
      return {
        ...prev,
        notifications: {
          ...prev.notifications,
          [type]: {
            ...prev.notifications[type],
            [field]: value
          }
        }
      }
    })
  }

  return (
    <div style={{ padding: '2rem' }}>
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2rem', fontWeight: 'bold', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <Bell style={{ color: '#3b82f6' }} />
          Уведомления
        </h1>
        <p style={{ color: '#6b7280' }}>
          Настройка уведомлений через Telegram бота
        </p>
      </div>

      {message && (
        <div style={{
          padding: '1rem',
          marginBottom: '1rem',
          borderRadius: '0.5rem',
          backgroundColor: message.includes('Ошибка') ? '#fef2f2' : '#f0fdf4',
          color: message.includes('Ошибка') ? '#dc2626' : '#16a34a',
          border: `1px solid ${message.includes('Ошибка') ? '#fecaca' : '#bbf7d0'}`
        }}>
          {message}
        </div>
      )}

      {/* Telegram Bot Settings */}
      <div style={{ backgroundColor: 'white', padding: '2rem', borderRadius: '0.5rem', border: '1px solid #e5e7eb', marginBottom: '1rem' }}>
        <h2 style={{ marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <Bot size={20} />
          Telegram Bot
        </h2>
        <div style={{ marginBottom: '1rem' }}>
          <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
            Токен бота:
          </label>
          <input
            type="text"
            value={settings.telegram_bot_token}
            onChange={(e) => setSettings(prev => ({ ...prev, telegram_bot_token: e.target.value }))}
            placeholder="Введите токен Telegram бота"
            style={{
              width: '100%',
              padding: '0.75rem',
              border: '1px solid #d1d5db',
              borderRadius: '0.375rem',
              fontSize: '0.875rem'
            }}
          />
        </div>
        <div style={{ marginBottom: '1rem' }}>
          <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
            Chat ID:
          </label>
          <input
            type="text"
            value={settings.telegram_chat_id}
            onChange={(e) => setSettings(prev => ({ ...prev, telegram_chat_id: e.target.value }))}
            placeholder="Введите Chat ID (например: 123456789)"
            style={{
              width: '100%',
              padding: '0.75rem',
              border: '1px solid #d1d5db',
              borderRadius: '0.375rem',
              fontSize: '0.875rem'
            }}
          />
          <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.5rem' }}>
            Chat ID можно получить, отправив сообщение боту и проверив: https://api.telegram.org/botYOUR_TOKEN/getUpdates
          </p>
        </div>
        <button
          onClick={testNotification}
          disabled={isLoading || !settings.telegram_bot_token || !settings.telegram_chat_id}
          style={{
            padding: '0.5rem 1rem',
            backgroundColor: (settings.telegram_bot_token && settings.telegram_chat_id) ? '#3b82f6' : '#9ca3af',
            color: 'white',
            border: 'none',
            borderRadius: '0.375rem',
            cursor: (settings.telegram_bot_token && settings.telegram_chat_id) ? 'pointer' : 'not-allowed',
            marginRight: '0.5rem'
          }}
        >
          Отправить тестовое уведомление
        </button>
      </div>

      {/* Notification Types */}
      <div style={{ backgroundColor: 'white', padding: '2rem', borderRadius: '0.5rem', border: '1px solid #e5e7eb', marginBottom: '1rem' }}>
        <h2 style={{ marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <AlertTriangle size={20} />
          Типы уведомлений
        </h2>

        {/* Agent Offline */}
        <div style={{ marginBottom: '2rem', padding: '1rem', border: '1px solid #e5e7eb', borderRadius: '0.375rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem' }}>
            <Server size={16} />
            <h3 style={{ margin: 0, fontWeight: '600' }}>Агент не отвечает</h3>
            <label style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <input
                type="checkbox"
                checked={settings.notifications?.agent_offline?.enabled ?? false}
                onChange={(e) => updateNotification('agent_offline', 'enabled', e.target.checked)}
              />
              Включено
            </label>
          </div>
          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
              Текст уведомления:
            </label>
            <textarea
              value={settings.notifications?.agent_offline?.message || ''}
              onChange={(e) => updateNotification('agent_offline', 'message', e.target.value)}
              style={{
                width: '100%',
                padding: '0.75rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                minHeight: '80px',
                resize: 'vertical'
              }}
              placeholder="Введите текст уведомления"
            />
            <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.5rem' }}>
              Доступные переменные: {'{AGENT_NAME}'}
            </p>
          </div>
        </div>

        {/* Container Stopped */}
        <div style={{ marginBottom: '2rem', padding: '1rem', border: '1px solid #e5e7eb', borderRadius: '0.375rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem' }}>
            <Container size={16} />
            <h3 style={{ margin: 0, fontWeight: '600' }}>Контейнер остановился</h3>
            <label style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <input
                type="checkbox"
                checked={settings.notifications?.container_stopped?.enabled ?? false}
                onChange={(e) => updateNotification('container_stopped', 'enabled', e.target.checked)}
              />
              Включено
            </label>
          </div>
          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
              Текст уведомления:
            </label>
            <textarea
              value={settings.notifications?.container_stopped?.message || ''}
              onChange={(e) => updateNotification('container_stopped', 'message', e.target.value)}
              style={{
                width: '100%',
                padding: '0.75rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                minHeight: '80px',
                resize: 'vertical'
              }}
              placeholder="Введите текст уведомления"
            />
            <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.5rem' }}>
              Доступные переменные: {'{CONTAINER_NAME}'}, {'{AGENT_NAME}'}
            </p>
          </div>
        </div>

        {/* CPU Threshold */}
        <div style={{ marginBottom: '2rem', padding: '1rem', border: '1px solid #e5e7eb', borderRadius: '0.375rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem' }}>
            <Cpu size={16} />
            <h3 style={{ margin: 0, fontWeight: '600' }}>CPU &gt; N %</h3>
            <label style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <input
                type="checkbox"
                checked={settings.notifications?.cpu_threshold?.enabled ?? false}
                onChange={(e) => updateNotification('cpu_threshold', 'enabled', e.target.checked)}
              />
              Включено
            </label>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '1rem', marginBottom: '1rem' }}>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
                Порог CPU (%):
              </label>
              <input
                type="number"
                value={settings.notifications?.cpu_threshold?.threshold ?? 80}
                onChange={(e) => updateNotification('cpu_threshold', 'threshold', parseInt(e.target.value) || 0)}
                min="1"
                max="100"
                style={{
                  width: '100%',
                  padding: '0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '0.375rem',
                  fontSize: '0.875rem'
                }}
              />
            </div>
          </div>
          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
              Текст уведомления:
            </label>
            <textarea
              value={settings.notifications?.cpu_threshold?.message || ''}
              onChange={(e) => updateNotification('cpu_threshold', 'message', e.target.value)}
              style={{
                width: '100%',
                padding: '0.75rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                minHeight: '80px',
                resize: 'vertical'
              }}
              placeholder="Введите текст уведомления"
            />
            <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.5rem' }}>
              Доступные переменные: {'{AGENT_NAME}'}, {'{CPU_USAGE}'}
            </p>
          </div>
        </div>

        {/* RAM Threshold */}
        <div style={{ marginBottom: '2rem', padding: '1rem', border: '1px solid #e5e7eb', borderRadius: '0.375rem' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem' }}>
            <HardDrive size={16} />
            <h3 style={{ margin: 0, fontWeight: '600' }}>RAM &gt; M %</h3>
            <label style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <input
                type="checkbox"
                checked={settings.notifications?.ram_threshold?.enabled ?? false}
                onChange={(e) => updateNotification('ram_threshold', 'enabled', e.target.checked)}
              />
              Включено
            </label>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 2fr', gap: '1rem', marginBottom: '1rem' }}>
            <div>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
                Порог RAM (%):
              </label>
              <input
                type="number"
                value={settings.notifications?.ram_threshold?.threshold ?? 80}
                onChange={(e) => updateNotification('ram_threshold', 'threshold', parseInt(e.target.value) || 0)}
                min="1"
                max="100"
                style={{
                  width: '100%',
                  padding: '0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '0.375rem',
                  fontSize: '0.875rem'
                }}
              />
            </div>
          </div>
          <div>
            <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
              Текст уведомления:
            </label>
            <textarea
              value={settings.notifications?.ram_threshold?.message || ''}
              onChange={(e) => updateNotification('ram_threshold', 'message', e.target.value)}
              style={{
                width: '100%',
                padding: '0.75rem',
                border: '1px solid #d1d5db',
                borderRadius: '0.375rem',
                fontSize: '0.875rem',
                minHeight: '80px',
                resize: 'vertical'
              }}
              placeholder="Введите текст уведомления"
            />
            <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.5rem' }}>
              Доступные переменные: {'{AGENT_NAME}'}, {'{RAM_USAGE}'}
            </p>
          </div>
        </div>
      </div>

      {/* Save Button */}
      <div style={{ textAlign: 'right' }}>
        <button
          onClick={saveSettings}
          disabled={isLoading}
          style={{
            padding: '0.75rem 1.5rem',
            backgroundColor: isLoading ? '#9ca3af' : '#3b82f6',
            color: 'white',
            border: 'none',
            borderRadius: '0.375rem',
            cursor: isLoading ? 'not-allowed' : 'pointer',
            fontSize: '0.875rem',
            fontWeight: '500'
          }}
        >
          {isLoading ? 'Сохранение...' : 'Сохранить настройки'}
        </button>
      </div>
    </div>
  )
} 