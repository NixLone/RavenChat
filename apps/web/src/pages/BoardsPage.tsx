import { useEffect, useMemo, useState } from 'react'
import { request } from '../api/client'
import { BoardColumn, Task } from '../types'

const BOARD_ID = '55555555-5555-5555-5555-555555555555'

export default function BoardsPage() {
  const [columns, setColumns] = useState<BoardColumn[]>([])
  const [tasks, setTasks] = useState<Task[]>([])
  const [newTitle, setNewTitle] = useState('')

  const load = () => request<{ columns: BoardColumn[]; tasks: Task[] }>(`/api/v1/boards/${BOARD_ID}`).then(data => { setColumns(data.columns); setTasks(data.tasks) })
  useEffect(() => { load() }, [])

  const createTask = async () => {
    if (!columns[0] || !newTitle.trim()) return
    await request(`/api/v1/boards/${BOARD_ID}/tasks`, { method: 'POST', body: JSON.stringify({ title: newTitle, description: '', column_id: columns[0].id, status: 'todo' }) })
    setNewTitle('')
    load()
  }

  const moveTask = async (taskId: string, columnId: string, status: string) => {
    await request(`/api/v1/tasks/${taskId}/move`, { method: 'PATCH', body: JSON.stringify({ column_id: columnId, status }) })
    load()
  }

  const tasksByColumn = useMemo(() => {
    const map: Record<string, Task[]> = {}
    for (const c of columns) map[c.id] = []
    for (const t of tasks) if (map[t.column_id]) map[t.column_id].push(t)
    return map
  }, [columns, tasks])

  return (
    <div>
      <h3>Task Board</h3>
      <div style={{ display: 'flex', gap: 8, marginBottom: 12 }}>
        <input value={newTitle} onChange={e => setNewTitle(e.target.value)} placeholder='New task title' />
        <button onClick={createTask}>Create task</button>
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12 }}>
        {columns.map(c => (
          <div key={c.id} style={{ border: '1px solid #ccc', minHeight: 220, padding: 8 }}>
            <h4>{c.name}</h4>
            {(tasksByColumn[c.id] || []).map(t => (
              <div key={t.id} style={{ border: '1px solid #eee', padding: 8, marginBottom: 6 }}>
                <strong>{t.title}</strong>
                <div style={{ display: 'flex', gap: 4, marginTop: 6 }}>
                  {columns.filter(col => col.id !== c.id).map(target => (
                    <button key={target.id} onClick={() => moveTask(t.id, target.id, target.name.toLowerCase().replace(' ', '_'))}>→ {target.name}</button>
                  ))}
                </div>
              </div>
            ))}
          </div>
        ))}
      </div>
    </div>
  )
}
