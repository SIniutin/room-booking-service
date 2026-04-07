import { useState, useEffect } from 'react'
import { getRooms } from '../../api/client'
import { useAuth } from '../../context/AuthContext'
import { Loader, Alert, Empty } from '../ui'
import CreateRoomModal from '../modals/CreateRoomModal'
import CreateScheduleModal from '../modals/CreateScheduleModal'
import styles from './RoomsTab.module.css'

export default function RoomsTab() {
  const { token } = useAuth()
  const [rooms, setRooms] = useState([])
  const [loading, setLoading] = useState(true)
  const [showCreate, setShowCreate] = useState(false)
  const [scheduleRoom, setScheduleRoom] = useState(null)
  const [success, setSuccess] = useState('')

  useEffect(() => {
    setLoading(true)
    getRooms(token)
      .then((d) => setRooms(d.rooms || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [token])

  function handleRoomCreated(room) {
    setRooms((p) => [...p, room])
    setShowCreate(false)
    setSuccess('Room created successfully')
  }

  function handleScheduleCreated() {
    setScheduleRoom(null)
    setSuccess('Schedule saved successfully')
  }

  return (
    <div>
      <div className={styles.header}>
        <div>
          <h1 className="section-title" style={{ margin: 0 }}>Conference Rooms</h1>
          <p style={{ fontSize: 13, color: 'var(--text2)', marginTop: 4 }}>
            {rooms.length} room{rooms.length !== 1 ? 's' : ''} total
          </p>
        </div>
        <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
          + New room
        </button>
      </div>

      {success && <Alert type="success">{success}</Alert>}

      {loading ? (
        <div className="empty"><Loader /></div>
      ) : rooms.length === 0 ? (
        <Empty icon="🏢">No rooms yet. Create your first room.</Empty>
      ) : (
        <div className="grid-2">
          {rooms.map((r) => (
            <div key={r.id} className={styles.roomCard}>
              <h3 className={styles.roomName}>{r.name}</h3>
              {r.description && <p className={styles.roomDesc}>{r.description}</p>}
              <div className={styles.roomMeta}>
                {r.capacity && <span className="chip">👥 {r.capacity} seats</span>}
                <button
                  className={`chip ${styles.scheduleBtn}`}
                  onClick={() => setScheduleRoom(r)}
                >
                  📅 Set schedule
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {showCreate && (
        <CreateRoomModal onClose={() => setShowCreate(false)} onCreated={handleRoomCreated} />
      )}
      {scheduleRoom && (
        <CreateScheduleModal
          room={scheduleRoom}
          onClose={() => setScheduleRoom(null)}
          onCreated={handleScheduleCreated}
        />
      )}
    </div>
  )
}
