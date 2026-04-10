# TaskFlow

A complete full-stack take-home implementation for the Full Stack Engineer role.

Tech stack:

- Backend: Go 1.23, Gin, PostgreSQL, Goose migrations, JWT auth, bcrypt
- Frontend: React + TypeScript + Vite + React Router
- Infra: Docker Compose (PostgreSQL + API + Frontend)

## 1) Overview

TaskFlow is a minimal but production-minded project/task management system.

Implemented features:

- Register and login with JWT
- Project CRUD with ownership checks
- Task CRUD with status/priority/assignee/due-date
- Project-level task filtering by status and assignee
- Protected frontend routes and persisted auth state
- Loading, error, and empty states in all core screens
- Optimistic task status updates in the UI
- PostgreSQL migrations (up/down) and seed data
- One-command local startup with Docker Compose

## 2) Architecture Decisions

Backend design:

- Gin for a lightweight HTTP layer with explicit middleware.
- Raw SQL with pgx pool for clarity and control over relational logic.
- Goose migrations for deterministic schema evolution (no auto-migrate magic).
- Explicit permission checks to keep 401 and 403 behavior clean.

Data model decision:

- Added task.creator_id beyond the required fields.
- Why: needed to satisfy delete permission requirement (project owner OR task creator).

Frontend design:

- Custom component system (no UI library) to keep code straightforward and reviewable.
- AuthContext for token/user persistence in localStorage.
- React Router protected route wrapper for auth-gated views.
- Reusable TaskModal for both create/edit flows.

Tradeoffs:

- No drag-and-drop and no websocket sync in this version.
- Pagination exists in backend list endpoints but frontend keeps simple first-page flows.
- No test suite in this submission to keep scope balanced; code is structured to add integration tests quickly.

## 3) Running Locally

Prerequisite:

- Docker + Docker Compose plugin

Commands:

```bash
git clone <your-repo-url>
cd taskflow-tejas-pachgade
cp .env.example .env
docker compose up --build
```

App URLs:

- Frontend: http://localhost:3000
- API: http://localhost:8080
- Postgres: localhost:5432

## 4) Running Migrations

Migrations run automatically on API startup.

How it works:

- API boot calls Goose Up from backend/migrations
- If tables already exist, Goose skips previously applied migrations

Manual migration notes (if needed later):

- This project uses Goose-compatible SQL files with Up/Down blocks.
- You can run Goose manually from inside the API container if desired.

## 5) Test Credentials

Seeded user:

- Email: test@example.com
- Password: password123

Also seeded:

- One project
- Three tasks (todo, in_progress, done)

## 6) API Reference

Auth:

- POST /auth/register
- POST /auth/login

Users:

- GET /users

Projects:

- GET /projects?page=&limit=
- POST /projects
- GET /projects/:id
- PATCH /projects/:id
- DELETE /projects/:id
- GET /projects/:id/stats

Tasks:

- GET /projects/:id/tasks?status=&assignee=&page=&limit=
- POST /projects/:id/tasks
- PATCH /tasks/:id
- DELETE /tasks/:id

Error response shape:

```json
{ "error": "validation failed", "fields": { "email": "is required" } }
```

Standard auth/permission errors:

- 401: { "error": "unauthorized" }
- 403: { "error": "forbidden" }
- 404: { "error": "not found" }
