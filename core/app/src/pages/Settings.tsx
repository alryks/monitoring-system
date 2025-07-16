import { Settings as SettingsIcon, Shield, Bell } from 'lucide-react'

export default function Settings() {
  return (
    <div style={{ padding: '2rem' }}>
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2rem', fontWeight: 'bold', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <SettingsIcon style={{ color: '#3b82f6' }} />
          Настройки
        </h1>
        <p style={{ color: '#6b7280' }}>
          Управление пользователями и настройками системы
        </p>
      </div>

        <div style={{ backgroundColor: 'white', padding: '2rem', borderRadius: '0.5rem', border: '1px solid #e5e7eb' }}>
          <h2 style={{ marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Shield size={20} />
            Безопасность
          </h2>
          <p style={{ color: '#6b7280', marginBottom: '1rem' }}>
            Настройки безопасности
          </p>
          <button style={{
            padding: '0.5rem 1rem',
            backgroundColor: '#6b7280',
            color: 'white',
            border: 'none',
            borderRadius: '0.375rem',
            cursor: 'pointer'
          }}>
            Настройки безопасности
          </button>
        </div>

        <div style={{ backgroundColor: 'white', padding: '2rem', borderRadius: '0.5rem', border: '1px solid #e5e7eb', marginTop: '1rem' }}>
          <h2 style={{ marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Bell size={20} />
            Уведомления
          </h2>
          <p style={{ color: '#6b7280', marginBottom: '1rem' }}>
            Настройка уведомлений и оповещений
          </p>
          <button style={{
            padding: '0.5rem 1rem',
            backgroundColor: '#6b7280',
            color: 'white',
            border: 'none',
            borderRadius: '0.375rem',
            cursor: 'pointer'
          }}>
            Настройки уведомлений
          </button>
        </div>
    </div>
  )
} 