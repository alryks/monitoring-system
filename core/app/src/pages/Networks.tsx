import { useState } from 'react'
import { Network, Database } from 'lucide-react'

export default function Networks() {
  const [activeTab, setActiveTab] = useState<'networks' | 'volumes'>('networks')

  return (
    <div style={{ padding: '2rem' }}>
      <div style={{ marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '2rem', fontWeight: 'bold', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
          <Network style={{ color: '#3b82f6' }} />
          Сети и Тома
        </h1>
        
        <div style={{ display: 'flex', gap: '1rem', marginBottom: '2rem' }}>
          <button
            onClick={() => setActiveTab('networks')}
            style={{
              padding: '0.5rem 1rem',
              border: '1px solid #d1d5db',
              backgroundColor: activeTab === 'networks' ? '#3b82f6' : 'white',
              color: activeTab === 'networks' ? 'white' : '#374151',
              borderRadius: '0.375rem',
              cursor: 'pointer'
            }}
          >
            Сети
          </button>
          <button
            onClick={() => setActiveTab('volumes')}
            style={{
              padding: '0.5rem 1rem',
              border: '1px solid #d1d5db',
              backgroundColor: activeTab === 'volumes' ? '#3b82f6' : 'white',
              color: activeTab === 'volumes' ? 'white' : '#374151',
              borderRadius: '0.375rem',
              cursor: 'pointer'
            }}
          >
            Тома
          </button>
        </div>
      </div>

      {activeTab === 'networks' && (
        <div style={{ backgroundColor: 'white', padding: '2rem', borderRadius: '0.5rem', border: '1px solid #e5e7eb' }}>
          <h2 style={{ marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Network size={20} />
            Docker Сети
          </h2>
          <p style={{ color: '#6b7280' }}>
            Здесь будет отображаться список всех Docker сетей со всех агентов.
          </p>
        </div>
      )}

      {activeTab === 'volumes' && (
        <div style={{ backgroundColor: 'white', padding: '2rem', borderRadius: '0.5rem', border: '1px solid #e5e7eb' }}>
          <h2 style={{ marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Database size={20} />
            Docker Тома
          </h2>
          <p style={{ color: '#6b7280' }}>
            Здесь будет отображаться список всех Docker томов со всех агентов.
          </p>
        </div>
      )}
    </div>
  )
} 