const BASE = 'http://localhost:8080'

async function request(path, options = {}, token) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(BASE + path, { ...options, headers })
  const data = await res.json()
  if (!res.ok) throw new Error(data?.error?.message || 'Request failed')
  return data
}

// ── Auth ──────────────────────────────────────────────────────────────────────

export const dummyLogin = (role) =>
  request('/dummyLogin', { method: 'POST', body: JSON.stringify({ role }) })

export const login = (body) =>
  request('/login', { method: 'POST', body: JSON.stringify(body) })

export const register = (body) =>
  request('/register', { method: 'POST', body: JSON.stringify(body) })

// ── Rooms ─────────────────────────────────────────────────────────────────────

export const getRooms = (token) =>
  request('/rooms/list', {}, token)

export const createRoom = (token, body) =>
  request('/rooms/create', { method: 'POST', body: JSON.stringify(body) }, token)

// ── Schedules ─────────────────────────────────────────────────────────────────

export const createSchedule = (token, roomId, body) =>
  request(`/rooms/${roomId}/schedule/create`, { method: 'POST', body: JSON.stringify(body) }, token)

// ── Slots ─────────────────────────────────────────────────────────────────────

export const getSlots = (token, roomId, date) =>
  request(`/rooms/${roomId}/slots/list?date=${date}`, {}, token)

// ── Bookings ──────────────────────────────────────────────────────────────────

export const createBooking = (token, body) =>
  request('/bookings/create', { method: 'POST', body: JSON.stringify(body) }, token)

export const getMyBookings = (token) =>
  request('/bookings/my', {}, token)

export const getAllBookings = (token, page = 1, pageSize = 20) =>
  request(`/bookings/list?page=${page}&pageSize=${pageSize}`, {}, token)

export const cancelBooking = (token, bookingId) =>
  request(`/bookings/${bookingId}/cancel`, { method: 'POST' }, token)
