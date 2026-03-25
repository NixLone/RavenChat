import { FormEvent, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { request } from '../api/client'
import { useAuth } from '../state/auth'

export default function LoginPage() {
  const nav = useNavigate()
  const { login } = useAuth()
  const [form, setForm] = useState({ login: 'alice', password: 'password' })
  const [error, setError] = useState('')

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    try {
      const data = await request<{ token: string }>('/api/v1/auth/login', { method: 'POST', body: JSON.stringify(form) })
      login(data.token)
      nav('/app/chats')
    } catch (err) {
      setError(String(err))
    }
  }

  return (
    <div style={{ maxWidth: 360, margin: '80px auto' }}>
      <h2>Corporate Messenger Login</h2>
      <form onSubmit={submit} style={{ display: 'grid', gap: 8 }}>
        <input value={form.login} onChange={e => setForm({ ...form, login: e.target.value })} placeholder="username/email/phone" />
        <input value={form.password} onChange={e => setForm({ ...form, password: e.target.value })} type="password" placeholder="password" />
        <button type="submit">Sign in</button>
      </form>
      <small>Seed users password: <b>password</b></small>
      {error && <p style={{ color: 'red' }}>{error}</p>}
    </div>
  )
}
