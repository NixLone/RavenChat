import { Link, Outlet } from 'react-router-dom'
import { useAuth } from '../state/auth'

export default function Layout() {
  const { logout } = useAuth()
  return (
    <div style={{ display: 'grid', gridTemplateColumns: '220px 1fr', minHeight: '100vh', fontFamily: 'Inter, sans-serif' }}>
      <aside style={{ borderRight: '1px solid #ddd', padding: 16 }}>
        <h3>RavenChat</h3>
        <nav style={{ display: 'grid', gap: 8 }}>
          <Link to="/app/chats">Chats</Link>
          <Link to="/app/boards">Task Board</Link>
          <Link to="/app/audit">Audit</Link>
        </nav>
        <button onClick={logout} style={{ marginTop: 20 }}>Logout</button>
      </aside>
      <main style={{ padding: 16 }}>
        <Outlet />
      </main>
    </div>
  )
}
