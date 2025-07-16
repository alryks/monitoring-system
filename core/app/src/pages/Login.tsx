import { useState } from 'react'
import { useAuth } from '../contexts/AuthContext'
import { Activity, AlertCircle } from 'lucide-react'
import styles from './Login.module.css'

export default function Login() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useAuth()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      await login(username, password)
    } catch (error: any) {
      setError(
        error.response?.data?.message || 
        error.message || 
        'Ошибка авторизации'
      )
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <div className={styles.header}>
          <Activity className={styles.icon} />
          <h2 className={styles.title}>
            Система мониторинга
          </h2>
          <p className={styles.subtitle}>
            Войдите в свою учетную запись
          </p>
        </div>

        <form className={styles.form} onSubmit={handleSubmit}>
          {error && (
            <div className={styles.error}>
              <AlertCircle className={styles.errorIcon} />
              <p className={styles.errorText}>{error}</p>
            </div>
          )}

          <div className={styles.field}>
            <label htmlFor="username" className={styles.label}>
              Имя пользователя
            </label>
            <input
              id="username"
              name="username"
              type="text"
              required
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className={styles.input}
              placeholder="admin"
            />
          </div>

          <div className={styles.field}>
            <label htmlFor="password" className={styles.label}>
              Пароль
            </label>
            <input
              id="password"
              name="password"
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className={styles.input}
              placeholder="admin"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className={styles.button}
          >
            {loading ? 'Вход...' : 'Войти'}
          </button>

          <p className={styles.hint}>
            По умолчанию: admin / admin
          </p>
        </form>
      </div>
    </div>
  )
} 