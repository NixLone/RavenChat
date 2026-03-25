export type User = { id: string; username: string; email: string; phone: string; role: string }
export type Chat = { id: string; kind: 'direct' | 'group'; name: string }
export type Message = { id: string; chat_id: string; sender_id: string; body: string; created_at: string }
export type Group = { id: string; name: string }
export type BoardColumn = { id: string; name: string; sort_order: number }
export type Task = {
  id: string
  board_id: string
  column_id: string
  title: string
  description: string
  assignee_id?: string
  creator_id: string
  due_date?: string
  status: string
}
export type AuditEvent = { id: string; event_type: string; target_type: string; created_at: string; payload: Record<string, unknown> }
