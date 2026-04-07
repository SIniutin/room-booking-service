import { useState, useEffect } from 'react'
import { getRooms, getSlots, createBooking } from '../../api/client'
import { useAuth } from '../../context/AuthContext'
import { Loader, Alert, Empty, StepLabel } from '../ui'
import { formatDate, formatTime } from '../../utils/format'
import styles from './BookRoomTab.module.css'

function todayISO() {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

export default function BookRoomTab() {
  const { token } = useAuth()

  const [rooms, setRooms] = useState([])
  const [selectedRoom, setSelectedRoom] = useState(null)
  const [date, setDate] = useState(todayISO())
  const [slots, setSlots] = useState([])
  const [selectedSlot, setSelectedSlot] = useState(null)
  const [withConf, setWithConf] = useState(false)

  const [loadingRooms, setLoadingRooms] = useState(true)
  const [loadingSlots, setLoadingSlots] = useState(false)
  const [booking, setBooking] = useState(false)

  const [success, setSuccess] = useState('')
  const [confLink, setConfLink] = useState(null)
  const [error, setError] = useState('')

  useEffect(() => {
    getRooms(token)
      .then((d) => setRooms(d.rooms || []))
      .catch(() => {})
      .finally(() => setLoadingRooms(false))
  }, [token])

  useEffect(() => {
    if (!selectedRoom || !date) return
    setSlots([]); setSelectedSlot(null); setSuccess(''); setError(''); setConfLink(null)
    setLoadingSlots(true)
    getSlots(token, selectedRoom.id, date)
      .then((d) => setSlots(d.slots || []))
      .catch(() => {})
      .finally(() => setLoadingSlots(false))
  }, [selectedRoom, date, token])

  async function handleBook() {
    if (!selectedSlot) return
    setBooking(true); setError(''); setSuccess(''); setConfLink(null)
    try {
      const data = await createBooking(token, { slotId: selectedSlot.id, createConferenceLink: withConf })
      setSuccess('Booking confirmed!')
      setConfLink(data.booking?.conferenceLink || null)
      setSelectedSlot(null)
      // refresh slots
      getSlots(token, selectedRoom.id, date)
        .then((d) => setSlots(d.slots || []))
        .catch(() => {})
    } catch (e) {
      setError(e.message)
    } finally {
      setBooking(false)
    }
  }

  return (
    <div>
      <h1 className="section-title">Book a Room</h1>
      <p className="section-sub">Select a room, pick a date, and choose an available slot</p>

      {success && (
        <Alert type="success">
          ✓ {success}
          {confLink && (
            <div style={{ marginTop: 6 }}>
              Conference link: <a className="link" href={confLink} target="_blank" rel="noreferrer">{confLink}</a>
            </div>
          )}
        </Alert>
      )}
      {error && <Alert type="error">{error}</Alert>}

      {/* Step 1 — Room */}
      <div className="card" style={{ marginBottom: 12 }}>
        <StepLabel n={1}>Select a room</StepLabel>
        {loadingRooms ? (
          <Loader />
        ) : rooms.length === 0 ? (
          <Empty icon="🏢">No rooms available</Empty>
        ) : (
          <div className="grid-3">
            {rooms.map((r) => (
              <div
                key={r.id}
                className={`${styles.roomOption} ${selectedRoom?.id === r.id ? styles.selected : ''}`}
                onClick={() => { setSelectedRoom(r); setSuccess(''); setError('') }}
              >
                <strong>{r.name}</strong>
                {r.capacity && <span className={styles.cap}>👥 {r.capacity}</span>}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Step 2 — Date */}
      {selectedRoom && (
        <div className="card" style={{ marginBottom: 12 }}>
          <StepLabel n={2}>Pick a date</StepLabel>
          <input
            type="date"
            value={date}
            min={todayISO()}
            onChange={(e) => setDate(e.target.value)}
            className={styles.dateInput}
          />
        </div>
      )}

      {/* Step 3 — Slot */}
      {selectedRoom && date && (
        <div className="card" style={{ marginBottom: 12 }}>
          <StepLabel n={3}>Choose a time slot</StepLabel>
          {loadingSlots ? (
            <div style={{ padding: '16px 0' }}><Loader /></div>
          ) : slots.length === 0 ? (
            <Empty icon="🕐">No available slots for this date</Empty>
          ) : (
            <div className={styles.slotGrid}>
              {slots.map((s) => (
                <button
                  key={s.id}
                  className={`${styles.slot} ${selectedSlot?.id === s.id ? styles.slotSelected : ''}`}
                  onClick={() => setSelectedSlot(s)}
                >
                  {formatTime(s.start)}–{formatTime(s.end)}
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Step 4 — Confirm */}
      {selectedSlot && (
        <div className="card">
          <StepLabel n={4}>Confirm booking</StepLabel>
          <p style={{ fontSize: 14, marginBottom: 14, color: 'var(--text)' }}>
            <strong>{selectedRoom.name}</strong>
            {' · '}
            {formatDate(date)}
            {' · '}
            {formatTime(selectedSlot.start)}–{formatTime(selectedSlot.end)}
          </p>
          <label className={styles.confToggle}>
            <input
              type="checkbox"
              checked={withConf}
              onChange={(e) => setWithConf(e.target.checked)}
              style={{ accentColor: 'var(--accent)', width: 16, height: 16 }}
            />
            Generate conference link
          </label>
          <button className="btn btn-primary" onClick={handleBook} disabled={booking}>
            {booking ? <Loader /> : 'Confirm booking →'}
          </button>
        </div>
      )}
    </div>
  )
}
