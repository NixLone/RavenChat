import { useEffect, useMemo, useState } from 'react'
import { request, wsUrl } from '../api/client'
import { Chat, Group, Message } from '../types'

export default function ChatsPage() {
  const [chats, setChats] = useState<Chat[]>([])
  const [groups, setGroups] = useState<Group[]>([])
  const [activeChat, setActiveChat] = useState<string>('')
  const [messages, setMessages] = useState<Message[]>([])
  const [body, setBody] = useState('')

  useEffect(() => {
    request<Chat[]>('/api/v1/chats').then(c => { setChats(c); if (c[0]) setActiveChat(c[0].id) })
    request<Group[]>('/api/v1/groups').then(setGroups)
  }, [])

  useEffect(() => {
    if (!activeChat) return
    request<Message[]>(`/api/v1/chats/${activeChat}/messages`).then(setMessages)
    const ws = new WebSocket(wsUrl(activeChat))
    ws.onmessage = e => {
      const parsed = JSON.parse(e.data)
      if (parsed?.type === 'message.created') setMessages(prev => [...prev, parsed.message])
    }
    return () => ws.close()
  }, [activeChat])

  const send = async () => {
    if (!body.trim()) return
    const m = await request<Message>(`/api/v1/chats/${activeChat}/messages`, { method: 'POST', body: JSON.stringify({ body }) })
    setMessages(prev => [...prev, m])
    setBody('')
  }

  const active = useMemo(() => chats.find(c => c.id === activeChat), [chats, activeChat])

  return (
    <div style={{ display: 'grid', gridTemplateColumns: '260px 1fr', gap: 12 }}>
      <section>
        <h3>Chats</h3>
        {chats.map(c => <button key={c.id} onClick={() => setActiveChat(c.id)} style={{ display: 'block', width: '100%', marginBottom: 6 }}>{c.name} ({c.kind})</button>)}
        <h4>Groups</h4>
        {groups.map(g => <div key={g.id}>{g.name}</div>)}
      </section>
      <section>
        <h3>{active?.name || 'Select chat'}</h3>
        <div style={{ border: '1px solid #ccc', minHeight: 380, padding: 8 }}>
          {messages.map(m => <p key={m.id}><b>{m.sender_id.slice(0, 6)}:</b> {m.body}</p>)}
        </div>
        <div style={{ display: 'flex', marginTop: 8, gap: 8 }}>
          <input value={body} onChange={e => setBody(e.target.value)} style={{ flex: 1 }} />
          <button onClick={send}>Send</button>
        </div>
      </section>
    </div>
  )
}
