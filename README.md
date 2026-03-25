# RavenChat MVP (Web-First Secure Corporate Messenger)

Production-minded MVP foundation for a modular-monolith secure messenger for ~200 employees.

## Implemented in this first vertical slice
- Go backend (chi + pgx + Redis + MinIO/S3 abstraction + WebSocket)
- React + TypeScript web app (Vite)
- PostgreSQL schema/migrations and seed data
- Local auth adapter with abstraction for external IdP replacement
- User/profile foundation with roles
- Private/group chat foundation with real-time updates via WebSocket
- Task board foundation (To Do / In Progress / Done)
- File upload foundation to MinIO + metadata in PostgreSQL
- Audit event foundation + simple audit viewer
- Docker Compose local environment

## Repository layout
- `apps/api` - Go API modular monolith
- `apps/web` - React web app
- `deploy/docker` - local development compose stack
- `docs` - architecture and module notes

## Quick start
```bash
cd deploy/docker
docker compose up --build
```

After startup:
- Web: http://localhost:5173
- API: http://localhost:8080/health
- MinIO console: http://localhost:9001 (minioadmin / minioadmin)

## Seed/demo users
All seed users have password `password`:
- `alice` (system_admin)
- `bob` (user)
- `carol` (security_auditor)

## MVP module boundaries (backend)
Implemented packages:
- `auth` (abstraction + local provider + JWT)
- `users`
- `invites` (schema/interface TODO)
- `chats`
- `groups`
- `boards`
- `tasks` (via boards package in first slice)
- `files`
- `audit`
- `presence`
- `notifications` (stub)
- `search` (stub)

## Security notes for this iteration
- Designed with encryption boundary by storing `messages.body_ciphertext` field naming and service-layer boundaries.
- Runtime plaintext processing is allowed in this first slice.
- Full cryptographic key management and at-rest encryption implementation is intentionally deferred.

## Deferred intentionally
- External IdP (e.g., Keycloak) adapter implementation (interfaces are prepared)
- Invite workflows endpoints (schema/flow next)
- Message edit/delete enforcement by role/time in handlers
- Notification infrastructure and delivery channels
- Advanced presence/typing events
- PM full-text search implementation (PostgreSQL strategy kept for next pass)
