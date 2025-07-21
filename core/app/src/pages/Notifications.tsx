import { useState, useEffect } from 'react'
import { Bell, Bot, AlertTriangle, Cpu, HardDrive, Server, Container, Mail } from 'lucide-react'
import { notificationsApi, type NotificationSettings } from '../services/api'

export default function Notifications() {
  const [settings, setSettings] = useState<NotificationSettings>({
    telegram_bot_token: '',
    telegram_chat_id: '',
    email_settings: {
      enabled: false,
      smtp_host: '',
      smtp_port: 587,
      username: '',
      password: '',
      from_email: '',
      from_name: 'Система мониторинга',
      to_emails: '',
      use_tls: true,
      use_start_tls: true
    },
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
          Настройка уведомлений через Telegram бота и email
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
          disabled={isLoading || (!settings.telegram_bot_token && !settings.email_settings.enabled)}
          style={{
            padding: '0.5rem 1rem',
            backgroundColor: (settings.telegram_bot_token || settings.email_settings.enabled) ? '#3b82f6' : '#9ca3af',
            color: 'white',
            border: 'none',
            borderRadius: '0.375rem',
            cursor: (settings.telegram_bot_token || settings.email_settings.enabled) ? 'pointer' : 'not-allowed',
            marginRight: '0.5rem'
          }}
        >
          Отправить тестовое уведомление
        </button>
      </div>

      {/* Email Settings */}
      <div style={{ backgroundColor: 'white', padding: '2rem', borderRadius: '0.5rem', border: '1px solid #e5e7eb', marginBottom: '1rem' }}>
        <h2 style={{ marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <Mail size={20} />
          Email уведомления
        </h2>
        
        <div style={{ marginBottom: '1rem' }}>
          <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem' }}>
            <input
              type="checkbox"
              checked={settings.email_settings.enabled}
              onChange={(e) => setSettings(prev => ({
                ...prev,
                email_settings: { ...prev.email_settings, enabled: e.target.checked }
              }))}
            />
            Включить email уведомления
          </label>
        </div>

        {settings.email_settings.enabled && (
          <>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
                  SMTP сервер:
                </label>
                <input
                  type="text"
                  value={settings.email_settings.smtp_host}
                  onChange={(e) => setSettings(prev => ({
                    ...prev,
                    email_settings: { ...prev.email_settings, smtp_host: e.target.value }
                  }))}
                  placeholder="smtp.gmail.com"
                  style={{
                    width: '100%',
                    padding: '0.75rem',
                    border: '1px solid #d1d5db',
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem'
                  }}
                />
              </div>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
                  Порт:
                </label>
                <input
                  type="number"
                  value={settings.email_settings.smtp_port}
                  onChange={(e) => setSettings(prev => ({
                    ...prev,
                    email_settings: { ...prev.email_settings, smtp_port: parseInt(e.target.value) || 587 }
                  }))}
                  placeholder="587"
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

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
                  Имя пользователя:
                </label>
                <input
                  type="text"
                  value={settings.email_settings.username}
                  onChange={(e) => setSettings(prev => ({
                    ...prev,
                    email_settings: { ...prev.email_settings, username: e.target.value }
                  }))}
                  placeholder="your-email@gmail.com"
                  style={{
                    width: '100%',
                    padding: '0.75rem',
                    border: '1px solid #d1d5db',
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem'
                  }}
                />
              </div>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
                  Пароль:
                </label>
                <input
                  type="password"
                  value={settings.email_settings.password}
                  onChange={(e) => setSettings(prev => ({
                    ...prev,
                    email_settings: { ...prev.email_settings, password: e.target.value }
                  }))}
                  placeholder="Ваш пароль или токен приложения"
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

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem', marginBottom: '1rem' }}>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
                  Email отправителя:
                </label>
                <input
                  type="email"
                  value={settings.email_settings.from_email}
                  onChange={(e) => setSettings(prev => ({
                    ...prev,
                    email_settings: { ...prev.email_settings, from_email: e.target.value }
                  }))}
                  placeholder="noreply@yourdomain.com"
                  style={{
                    width: '100%',
                    padding: '0.75rem',
                    border: '1px solid #d1d5db',
                    borderRadius: '0.375rem',
                    fontSize: '0.875rem'
                  }}
                />
              </div>
              <div>
                <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
                  Имя отправителя:
                </label>
                <input
                  type="text"
                  value={settings.email_settings.from_name}
                  onChange={(e) => setSettings(prev => ({
                    ...prev,
                    email_settings: { ...prev.email_settings, from_name: e.target.value }
                  }))}
                  placeholder="Система мониторинга"
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

            <div style={{ marginBottom: '1rem' }}>
              <label style={{ display: 'block', marginBottom: '0.5rem', fontWeight: '500' }}>
                Email получателей:
              </label>
              <input
                type="text"
                value={settings.email_settings.to_emails}
                onChange={(e) => setSettings(prev => ({
                  ...prev,
                  email_settings: { ...prev.email_settings, to_emails: e.target.value }
                }))}
                placeholder="admin@example.com, support@example.com"
                style={{
                  width: '100%',
                  padding: '0.75rem',
                  border: '1px solid #d1d5db',
                  borderRadius: '0.375rem',
                  fontSize: '0.875rem'
                }}
              />
              <p style={{ fontSize: '0.75rem', color: '#6b7280', marginTop: '0.5rem' }}>
                Укажите email адреса через запятую для отправки уведомлений нескольким получателям
              </p>
            </div>

            <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem' }}>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <input
                  type="checkbox"
                  checked={settings.email_settings.use_tls}
                  onChange={(e) => setSettings(prev => ({
                    ...prev,
                    email_settings: { ...prev.email_settings, use_tls: e.target.checked }
                  }))}
                />
                Использовать TLS
              </label>
              <label style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <input
                  type="checkbox"
                  checked={settings.email_settings.use_start_tls}
                  onChange={(e) => setSettings(prev => ({
                    ...prev,
                    email_settings: { ...prev.email_settings, use_start_tls: e.target.checked }
                  }))}
                />
                Использовать STARTTLS
              </label>
            </div>

            <div style={{ padding: '1rem', backgroundColor: '#f3f4f6', borderRadius: '0.375rem', marginBottom: '1rem' }}>
              <h4 style={{ margin: '0 0 0.5rem 0', fontSize: '0.875rem', fontWeight: '600' }}>Примеры настроек:</h4>
              <div style={{ fontSize: '0.75rem', color: '#6b7280' }}>
                <p><strong>Gmail:</strong> smtp.gmail.com:587, используйте токен приложения вместо пароля</p>
                <p><strong>Yandex:</strong> smtp.yandex.ru:465, включите TLS</p>
                <p><strong>Mail.ru:</strong> smtp.mail.ru:465, включите TLS</p>
                <p><strong>Outlook:</strong> smtp-mail.outlook.com:587, включите STARTTLS</p>
              </div>
            </div>
          </>
        )}
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