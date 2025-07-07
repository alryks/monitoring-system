import { useState } from 'react'
import axios from 'axios'
import './App.css'
import AgentsPage from './components/AgentsPage'
import './components/AgentsPage.css'

interface PingResponse {
  message: string
  timestamp: string
  version: string
}

function App() {
  const [response, setResponse] = useState<PingResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [currentPage, setCurrentPage] = useState('home')

  const testConnection = async () => {
    setLoading(true)
    setError(null)
    setResponse(null)

    try {
      const result = await axios.get<PingResponse>('/api/ping')
      setResponse(result.data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Произошла ошибка')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="App">
      <header className="App-header">
        <h1>Система мониторинга</h1>
        <nav>
          <button onClick={() => setCurrentPage('home')}>Главная</button>
          <button onClick={() => setCurrentPage('agents')}>Агенты</button>
        </nav>
        
        {currentPage === 'home' && (
          <div className="connection-test">
            <h2>Тестирование соединения</h2>
            <p>Нажмите кнопку ниже, чтобы проверить соединение с API</p>
            <button 
              className="test-button"
              onClick={testConnection} 
              disabled={loading}
            >
              {loading ? 'Тестируем...' : 'Тестировать соединение'}
            </button>

            {error && (
              <div className="result error">
                <h3>Ошибка соединения:</h3>
                <p>{error}</p>
              </div>
            )}

            {response && (
              <div className="result success">
                <h3>Соединение установлено!</h3>
                <p><strong>Сообщение:</strong> {response.message}</p>
                <p><strong>Версия:</strong> {response.version}</p>
                <p><strong>Время:</strong> {new Date(response.timestamp).toLocaleString()}</p>
              </div>
            )}
          </div>
        )}
      </header>
      
      <main>
        {currentPage === 'home' && (
          <div>
            <h2>Добро пожаловать в систему мониторинга</h2>
            <p>Высоконагруженная система управления и балансировки сервисов</p>
            <p>Используйте навигацию выше для перехода к разным разделам системы</p>
          </div>
        )}
        
        {currentPage === 'agents' && <AgentsPage />}
      </main>
    </div>
  )
}

export default App