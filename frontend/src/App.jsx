import { useState } from 'react'
import { AuthProvider, useAuth } from './context/AuthContext'
import LoginPage from './pages/LoginPage'
import Topbar from './components/Topbar'
import RoomsTab from './components/admin/RoomsTab'
import BookingsTab from './components/admin/BookingsTab'
import BookRoomTab from './components/user/BookRoomTab'
import MyBookingsTab from './components/user/MyBookingsTab'
import styles from './App.module.css'

const ADMIN_TABS = ['Rooms', 'All Bookings']
const USER_TABS  = ['Book a Room', 'My Bookings']

function Shell() {
  const { token, role } = useAuth()
  const [tab, setTab] = useState(0)

  if (!token) return <LoginPage />

  const tabs = role === 'admin' ? ADMIN_TABS : USER_TABS

  return (
    <div className={styles.app}>
      <Topbar tabs={tabs} activeTab={tab} onTabChange={setTab} />
      <main className={styles.main}>
        {role === 'admin' && tab === 0 && <RoomsTab />}
        {role === 'admin' && tab === 1 && <BookingsTab />}
        {role === 'user'  && tab === 0 && <BookRoomTab />}
        {role === 'user'  && tab === 1 && <MyBookingsTab />}
      </main>
    </div>
  )
}

export default function App() {
  return (
    <AuthProvider>
      <Shell />
    </AuthProvider>
  )
}
