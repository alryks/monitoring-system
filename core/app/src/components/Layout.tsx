import { Link, useLocation } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { 
  Monitor,
  LogOut, 
  User,
  Activity,
  Server,
  Container,
  Package,
  Network,
  Settings,
} from 'lucide-react'
import styles from './Layout.module.css'

interface LayoutProps {
  children: React.ReactNode
}

export default function Layout({ children }: LayoutProps) {
  const { user, logout } = useAuth()
  const location = useLocation()

  const navigation = [
    { name: 'Дашборд', href: '/dashboard', icon: Monitor },
    { name: 'Агенты', href: '/agents', icon: Server },
    { name: 'Контейнеры', href: '/containers', icon: Container },
    { name: 'Образы', href: '/images', icon: Package },
    { name: 'Сети и Тома', href: '/networks', icon: Network },
    { name: 'Настройки', href: '/settings', icon: Settings, adminOnly: true },
  ]

  const filteredNavigation = navigation.filter(item => {
    if (item.adminOnly && user?.role !== 'admin') {
      return false
    }
    return true
  })

  return (
    <div className={styles.layout}>
      {/* Sidebar */}
      <div className={styles.sidebar}>
        <div className={styles.sidebarHeader}>
          <div className={styles.logo}>
            <Activity className={styles.logoIcon} />
            <span>Мониторинг</span>
          </div>
        </div>

        {/* Navigation */}
        <nav className={styles.nav}>
          <ul className={styles.navList}>
            {filteredNavigation.map((item) => {
              const Icon = item.icon
              const isActive = location.pathname === item.href
              
              return (
                <li key={item.name} className={styles.navItem}>
                  <Link
                    to={item.href}
                    className={`${styles.navLink} ${isActive ? styles.active : ''}`}
                  >
                    <Icon className={styles.navIcon} />
                    {item.name}
                  </Link>
                </li>
              )
            })}
          </ul>
        </nav>

        {/* User info */}
        <div className={styles.userSection}>
          <div className={styles.userInfo}>
            <div className={styles.userIcon}>
              <User />
            </div>
            <div>
              <div className={styles.userName}>{user?.username}</div>
              <div className={styles.userRole}>{user?.role}</div>
            </div>
          </div>
          <button
            onClick={logout}
            className={styles.logoutButton}
            title="Выйти"
          >
            <LogOut className={styles.logoutIcon} />
            Выйти
          </button>
        </div>
      </div>

      {/* Main content */}
      <div className={styles.main}>
        <main className={styles.content}>
          {children}
        </main>
      </div>
    </div>
  )
} 