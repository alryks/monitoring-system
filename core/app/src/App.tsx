import { useState, useEffect } from 'react'
import axios from 'axios'
import './App.css'

interface PingResponse {
  message: string
  timestamp: string
  version: string
}

function App() {
  const [pingData, setPingData] = useState<PingResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const testConnection = async () => {
    setLoading(true)
    setError(null)
    
    try {
      const response = await axios.get<PingResponse>('/api/ping')
      setPingData(response.data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Произошла ошибка')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    testConnection()
  }, [])

  return (
    <div className="App">
      <header className="App-header">
        <h1>Система мониторинга</h1>
        <p>Высоконагруженная система управления и балансировки сервисов</p>
        
        <div className="connection-test">
          <h2>Проверка подключения к API</h2>
          
          <button onClick={testConnection} disabled={loading}>
            {loading ? 'Проверка...' : 'Проверить подключение'}
          </button>
          
          {error && (
            <div className="error">
              <p>Ошибка подключения: {error}</p>
            </div>
          )}
          
          {pingData && (
            <div className="success">
              <p>✅ Подключение успешно!</p>
              <p>Ответ: {pingData.message}</p>
              <p>Время: {pingData.timestamp}</p>
              <p>Версия: {pingData.version}</p>
            </div>
          )}
        </div>
      </header>
    </div>
  )
}

export default App 