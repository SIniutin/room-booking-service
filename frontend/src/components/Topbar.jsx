import { useAuth } from '../context/AuthContext'
import styles from './Topbar.module.css'

export default function Topbar({ tabs, activeTab, onTabChange }) {
  const { role, logout } = useAuth()

  return (
    <header className={styles.topbar}>
      <div className={styles.brand}>
        <div className={styles.brandIcon}>R</div>
        <span>Room Booking</span>
      </div>

      <nav className={styles.tabs}>
        {tabs.map((t, i) => (
          <button
            key={i}
            className={`${styles.tab} ${activeTab === i ? styles.active : ''}`}
            onClick={() => onTabChange(i)}
          >
            {t}
          </button>
        ))}
      </nav>

      <div className={styles.right}>
        <span className={`badge badge-${role}`}>
          {role === 'admin' ? 'Administrator' : 'User'}
        </span>
        <button className="btn btn-sm" onClick={logout}>Sign out</button>
      </div>
    </header>
  )
}
