import styles from './Modal.module.css'

export default function Modal({ title, subtitle, onClose, children, actions }) {
  return (
    <div className={styles.backdrop} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        {title && <h2 className={styles.title}>{title}</h2>}
        {subtitle && <p className={styles.subtitle}>{subtitle}</p>}
        <div className={styles.body}>{children}</div>
        {actions && <div className={styles.actions}>{actions}</div>}
      </div>
    </div>
  )
}
