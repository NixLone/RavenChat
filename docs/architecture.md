# Architecture Decisions (MVP)

## Style
- Modular monolith backend in Go.
- Separate React frontend.
- REST + WebSocket transport.
- PostgreSQL as source of truth.
- Redis for ephemeral state.
- S3-compatible object storage abstraction (MinIO locally).

## Extension points
- `auth.Provider` allows replacing local auth with Keycloak/OIDC adapter.
- `notifications.Service` and `search.Service` are intentionally stubs.
- `chats.Service` is prepared for encryption and audit hooks.

## Data model summary
Core tables included:
- users, user_profiles
- chats, chat_members, messages
- groups, group_members
- boards, board_members, board_columns, tasks
- files
- audit_events
