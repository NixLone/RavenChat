import { useEffect, useState } from 'react'
import { request } from '../api/client'
import { AuditEvent } from '../types'

export default function AuditPage() {
  const [events, setEvents] = useState<AuditEvent[]>([])
  useEffect(() => { request<AuditEvent[]>('/api/v1/audit/events').then(setEvents) }, [])
  return (
    <div>
      <h3>Audit Events</h3>
      <table width="100%">
        <thead><tr><th>When</th><th>Event</th><th>Target</th><th>Payload</th></tr></thead>
        <tbody>
          {events.map(e => (
            <tr key={e.id}>
              <td>{new Date(e.created_at).toLocaleString()}</td>
              <td>{e.event_type}</td>
              <td>{e.target_type}</td>
              <td><pre>{JSON.stringify(e.payload)}</pre></td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
