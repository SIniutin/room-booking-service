import { useState, useEffect, useCallback } from 'react'
import { getMyBookings, cancelBooking } from '../../api/client'
import { useAuth } from '../../context/AuthContext'
import { Loader, Alert, Empty, StatusPill } from '../ui'
import { formatDateTime } from '../../utils/format'
import styles from './MyBookingsTab.module.css'

export default function MyBookingsTab() {
  const { token } = useAuth()
  const [bookings, setBookings] = useState([])
  const [loading, setLoading] = useState(true)
  const [cancelling, setCancelling] = useState(null)
  const [success, setSuccess] = useState('')
  const [error, setError] = useState('')

  const load = useCallback(() => {
    setLoading(true)
    getMyBookings(token)
      .then((d) => setBookings(d.bookings || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [token])

  useEffect(() => { load() }, [load])

  async function handleCancel(id) {
    setCancelling(id); setSuccess(''); setError('')
    try {
      await cancelBooking(token, id)
      setSuccess('Booking cancelled successfully')
      load()
    } catch (e) {
      setError(e.message)
    } finally {
      setCancelling(null)
    }
  }

  return (
    <div>
      <h1 className="section-title">My Bookings</h1>
      <p className="section-sub">Upcoming bookings — past bookings are not shown</p>

      {success && <Alert type="success">{success}</Alert>}
      {error   && <Alert type="error">{error}</Alert>}

      {loading ? (
        <div className="empty"><Loader /></div>
      ) : bookings.length === 0 ? (
        <Empty icon="📅">No upcoming bookings</Empty>
      ) : (
        <div>
          {bookings.map((b) => (
            <div key={b.id} className={`card ${styles.row}`}>
              <div className={styles.info}>
                <div className={styles.idLine}>
                  Booking{' '}
                  <code className={styles.code}>{b.id.slice(0, 8)}…</code>
                </div>
                <div className={styles.secondaryLine}>
                  Slot <code className={styles.code}>{b.slotId.slice(0, 8)}…</code>
                </div>
                <div className={styles.meta}>
                  <StatusPill status={b.status} />
                  {b.conferenceLink && (
                    <a className="link" href={b.conferenceLink} target="_blank" rel="noreferrer">
                      Join call ↗
                    </a>
                  )}
                  <span className={styles.date}>Created {formatDateTime(b.createdAt)}</span>
                </div>
              </div>
              {b.status === 'active' && (
                <button
                  className="btn btn-danger btn-sm"
                  onClick={() => handleCancel(b.id)}
                  disabled={cancelling === b.id}
                >
                  {cancelling === b.id ? <Loader /> : 'Cancel'}
                </button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
