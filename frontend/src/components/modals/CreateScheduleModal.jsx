import { useState } from 'react'
import { createSchedule } from '../../api/client'
import { useAuth } from '../../context/AuthContext'
import { Loader, Alert } from '../ui'
import Modal from './Modal'
import styles from './CreateScheduleModal.module.css'

const DAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
const TIME_24H_RE = /^([01]\d|2[0-3]):[0-5]\d$/

export default function CreateScheduleModal({ room, onClose, onCreated }) {
  const { token } = useAuth()
  const [days, setDays] = useState([1, 2, 3, 4, 5])
  const [startTime, setStartTime] = useState('09:00')
  const [endTime, setEndTime] = useState('18:00')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  function toggleDay(d) {
    setDays((p) => (p.includes(d) ? p.filter((x) => x !== d) : [...p, d].sort((a, b) => a - b)))
  }

  function isValidTime(value) {
    return TIME_24H_RE.test(value)
  }

  function normalizeTimeInput(value) {
    const digits = value.replace(/\D/g, '').slice(0, 4)
    if (digits.length <= 2) return digits
    return `${digits.slice(0, 2)}:${digits.slice(2)}`
  }

  async function submit() {
    if (!days.length) { setError('Select at least one day'); return }
    if (!isValidTime(startTime) || !isValidTime(endTime)) {
      setError('Use HH:MM in 24-hour format')
      return
    }
    setLoading(true); setError('')
    try {
      const data = await createSchedule(token, room.id, { daysOfWeek: days, startTime, endTime })
      onCreated(data.schedule)
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      title="Set schedule"
      subtitle={`Configure availability for "${room.name}". Slots are 30 minutes. Cannot be changed later.`}
      onClose={onClose}
      actions={
        <>
          <button className="btn" onClick={onClose}>Cancel</button>
          <button className="btn btn-primary" onClick={submit} disabled={loading}>
            {loading ? <Loader /> : 'Save schedule'}
          </button>
        </>
      }
    >
      {error && <Alert type="error">{error}</Alert>}
      <div className="form-group">
        <label>Available days</label>
        <div className={styles.daysRow}>
          {DAYS.map((d, i) => (
            <button
              key={i}
              className={`${styles.dayBtn} ${days.includes(i + 1) ? styles.on : ''}`}
              onClick={() => toggleDay(i + 1)}
              type="button"
            >
              {d}
            </button>
          ))}
        </div>
      </div>
      <div className="form-row">
        <div className="form-group">
          <label>Opens at (UTC)</label>
          <input
            type="text"
            inputMode="numeric"
            placeholder="09:00"
            maxLength={5}
            value={startTime}
            onChange={(e) => setStartTime(normalizeTimeInput(e.target.value))}
          />
        </div>
        <div className="form-group">
          <label>Closes at (UTC)</label>
          <input
            type="text"
            inputMode="numeric"
            placeholder="18:00"
            maxLength={5}
            value={endTime}
            onChange={(e) => setEndTime(normalizeTimeInput(e.target.value))}
          />
        </div>
      </div>
    </Modal>
  )
}
