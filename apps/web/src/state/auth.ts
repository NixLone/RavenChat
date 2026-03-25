import { createContext, useContext, useMemo, useState } from 'react'

type AuthCtx = {
  token: string | null
  login: (token: string) => void
  logout: () => void
}

const Ctx = createContext<AuthCtx | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'))
  const value = useMemo(() => ({
    token,
    login: (t: string) => { localStorage.setItem('token', t); setToken(t) },
    logout: () => { localStorage.removeItem('token'); setToken(null) },
  }), [token])
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}

export function useAuth() {
  const ctx = useContext(Ctx)
  if (!ctx) throw new Error('Auth provider missing')
  return ctx
}
