import { useState, useEffect } from 'react'
import { getAllBookings } from '../../api/client'
import { useAuth } from '../../context/AuthContext'
import { Loader, Empty, StatusPill, Pager } from '../ui'
import { formatDateTime } from '../../utils/format'

const PAGE_SIZE = 20

export default function BookingsTab() {
  const { token } = useAuth()
  const [bookings, setBookings] = useState([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)

  useEffect(() => {
    setLoading(true)
    getAllBookings(token, page, PAGE_SIZE)
      .then((d) => {
        setBookings(d.bookings || [])
        setTotal(d.pagination?.total || 0)
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [token, page])

  const totalPages = Math.ceil(total / PAGE_SIZE)

  return (
    <div>
      <h1 className="section-title">All Bookings</h1>
      <p className="section-sub">{total} booking{total !== 1 ? 's' : ''} total</p>

      {loading ? (
        <div className="empty"><Loader /></div>
      ) : bookings.length === 0 ? (
        <Empty icon="📋">No bookings found.</Empty>
      ) : (
        <>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Booking ID</th>
                  <th>Slot ID</th>
                  <th>User ID</th>
                  <th>Status</th>
                  <th>Conference</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody>
                {bookings.map((b) => (
                  <tr key={b.id}>
                    <td style={{ fontFamily: 'DM Mono, monospace', fontSize: 11 }}>
                      {b.id.slice(0, 8)}…
                    </td>
                    <td style={{ fontFamily: 'DM Mono, monospace', fontSize: 11 }}>
                      {b.slotId.slice(0, 8)}…
                    </td>
                    <td style={{ fontFamily: 'DM Mono, monospace', fontSize: 11 }}>
                      {b.userId.slice(0, 8)}…
                    </td>
                    <td><StatusPill status={b.status} /></td>
                    <td>
                      {b.conferenceLink
                        ? <a className="link" href={b.conferenceLink} target="_blank" rel="noreferrer">Join ↗</a>
                        : <span style={{ color: 'var(--text3)' }}>—</span>
                      }
                    </td>
                    <td style={{ color: 'var(--text2)', fontSize: 12 }}>
                      {formatDateTime(b.createdAt)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <Pager page={page} totalPages={totalPages} onChange={setPage} />
        </>
      )}
    </div>
  )
}
