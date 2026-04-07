export function formatDateTime(dt) {
  if (!dt) return '—'
  return new Date(dt).toLocaleString('en-GB', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

export function formatTime(dt) {
  if (!dt) return ''
  return new Date(dt).toLocaleTimeString('en-GB', {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  })
}

export function formatDate(dt) {
  if (!dt) return ''
  return new Date(dt).toLocaleDateString('en-GB', {
    weekday: 'long',
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}
