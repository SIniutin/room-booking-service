import { createContext, useContext, useState } from 'react'

const AuthContext = createContext(null)

function readRoleFromToken(token) {
  try {
    const [, payload] = token.split('.')
    const json = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')))
    return json.role || null
  } catch {
    return null
  }
}

export function AuthProvider({ children }) {
  const [token, setToken] = useState(() => sessionStorage.getItem('rb_token'))
  const [role, setRole]   = useState(() => sessionStorage.getItem('rb_role'))

  function login(t, r = null) {
    const resolvedRole = r || readRoleFromToken(t)
    if (!resolvedRole) {
      throw new Error('Could not determine user role from token')
    }

    sessionStorage.setItem('rb_token', t)
    sessionStorage.setItem('rb_role', resolvedRole)
    setToken(t)
    setRole(resolvedRole)
  }

  function logout() {
    sessionStorage.removeItem('rb_token')
    sessionStorage.removeItem('rb_role')
    setToken(null)
    setRole(null)
  }

  return (
    <AuthContext.Provider value={{ token, role, login, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  return useContext(AuthContext)
}
