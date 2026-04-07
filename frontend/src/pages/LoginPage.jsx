import { useState } from 'react'
import { dummyLogin, login as loginRequest, register as registerRequest } from '../api/client'
import { useAuth } from '../context/AuthContext'
import { Loader, Alert } from '../components/ui'
import styles from './LoginPage.module.css'

const modes = [
  { value: 'login', label: 'Sign in' },
  { value: 'register', label: 'Register' },
  { value: 'demo', label: 'Demo access' },
]

export default function LoginPage() {
  const { login } = useAuth()
  const [mode, setMode] = useState('login')
  const [role, setRole] = useState('user')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  async function handleSubmit(event) {
    event.preventDefault()
    setLoading(true)
    setError('')
    setSuccess('')

    try {
      if (mode === 'demo') {
        const data = await dummyLogin(role)
        login(data.token, role)
        return
      }

      if (mode === 'login') {
        const data = await loginRequest({ email, password })
        login(data.token)
        return
      }

      await registerRequest({ email, password, role })
      const data = await loginRequest({ email, password })
      login(data.token)
      setSuccess('Account created')
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.wrap}>
      <div className={styles.card}>
        <div className={styles.cardHeader}>
          <h1 className={styles.title}>Authorization</h1>
          <p className={styles.subtitle}>Sign in, register, or enter with demo access.</p>

          <div className={styles.modeTabs}>
            {modes.map((item) => (
              <button
                key={item.value}
                type="button"
                className={`${styles.modeTab} ${mode === item.value ? styles.modeTabActive : ''}`}
                onClick={() => {
                  setMode(item.value)
                  setError('')
                  setSuccess('')
                }}
              >
                {item.label}
              </button>
            ))}
          </div>
        </div>

        <form onSubmit={handleSubmit} className={styles.form}>
          {(mode === 'register' || mode === 'demo') && (
            <div className="form-group">
              <label htmlFor="role">Role</label>
              <div className={styles.rolePicker}>
                {[
                  { value: 'admin', title: 'Administrator', desc: 'Manage rooms and inspect all bookings' },
                  { value: 'user', title: 'User', desc: 'Book rooms and manage your own reservations' },
                ].map((item) => (
                  <button
                    key={item.value}
                    type="button"
                    className={`${styles.roleOpt} ${role === item.value ? styles.selected : ''}`}
                    onClick={() => setRole(item.value)}
                  >
                    <strong className={styles.roleTitle}>{item.title}</strong>
                    <span className={styles.roleDesc}>{item.desc}</span>
                  </button>
                ))}
              </div>
            </div>
          )}

          {mode !== 'demo' && (
            <>
              <div className="form-group">
                <label htmlFor="email">Email</label>
                <input
                  id="email"
                  type="email"
                  placeholder="user@example.com"
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  autoComplete="email"
                  required
                />
              </div>

              <div className="form-group">
                <label htmlFor="password">Password</label>
                <input
                  id="password"
                  type="password"
                  placeholder="Enter password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  autoComplete={mode === 'register' ? 'new-password' : 'current-password'}
                  required
                />
              </div>
            </>
          )}

          {error && <Alert type="error">{error}</Alert>}
          {success && <Alert type="success">{success}</Alert>}

          <button className={`btn btn-primary ${styles.submitBtn}`} type="submit" disabled={loading}>
            {loading ? <Loader /> : mode === 'demo' ? 'Enter demo workspace' : mode === 'register' ? 'Create account' : 'Sign in'}
          </button>
        </form>

        <p className={styles.hint}>
          {mode === 'demo'
            ? 'Demo mode issues a token without email and password.'
            : 'Regular auth uses the backend login and registration endpoints.'}
        </p>
      </div>
    </div>
  )
}
