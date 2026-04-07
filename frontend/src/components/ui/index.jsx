// Shared low-level UI components

export function Loader() {
  return <span className="loader" />
}

export function Alert({ type = 'error', children }) {
  return <div className={`alert alert-${type}`}>{children}</div>
}

export function Empty({ icon = '📭', children }) {
  return (
    <div className="empty">
      <span className="icon">{icon}</span>
      {children}
    </div>
  )
}

export function StatusPill({ status }) {
  return <span className={`status-pill status-${status}`}>{status}</span>
}

export function Pager({ page, totalPages, onChange }) {
  if (totalPages <= 1) return null
  return (
    <div className="pager">
      <button className="btn btn-sm" onClick={() => onChange(page - 1)} disabled={page === 1}>
        ← Prev
      </button>
      <span>Page {page} of {totalPages}</span>
      <button className="btn btn-sm" onClick={() => onChange(page + 1)} disabled={page === totalPages}>
        Next →
      </button>
    </div>
  )
}

export function StepLabel({ n, children }) {
  return (
    <div style={{
      fontSize: 11,
      fontWeight: 600,
      color: 'var(--text3)',
      marginBottom: 12,
      textTransform: 'uppercase',
      letterSpacing: '0.8px',
    }}>
      Step {n} — {children}
    </div>
  )
}
