import { useState } from 'react'
import { createRoom } from '../../api/client'
import { useAuth } from '../../context/AuthContext'
import { Loader, Alert } from '../ui'
import Modal from './Modal'

export default function CreateRoomModal({ onClose, onCreated }) {
  const { token } = useAuth()
  const [form, setForm] = useState({ name: '', description: '', capacity: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const set = (k) => (e) => setForm((p) => ({ ...p, [k]: e.target.value }))

  async function submit() {
    if (!form.name.trim()) { setError('Room name is required'); return }
    setLoading(true); setError('')
    try {
      const body = {
        name: form.name.trim(),
        description: form.description.trim() || null,
        capacity: form.capacity ? parseInt(form.capacity, 10) : null,
      }
      const data = await createRoom(token, body)
      onCreated(data.room)
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Modal
      title="Create room"
      subtitle="Add a new conference room to the system"
      onClose={onClose}
      actions={
        <>
          <button className="btn" onClick={onClose}>Cancel</button>
          <button className="btn btn-primary" onClick={submit} disabled={loading}>
            {loading ? <Loader /> : 'Create room'}
          </button>
        </>
      }
    >
      {error && <Alert type="error">{error}</Alert>}
      <div className="form-group">
        <label>Room name *</label>
        <input value={form.name} onChange={set('name')} placeholder="e.g. Boardroom A" autoFocus />
      </div>
      <div className="form-group">
        <label>Description</label>
        <input value={form.description} onChange={set('description')} placeholder="Optional description" />
      </div>
      <div className="form-group">
        <label>Capacity</label>
        <input type="number" min="1" value={form.capacity} onChange={set('capacity')} placeholder="Number of seats" />
      </div>
    </Modal>
  )
}
