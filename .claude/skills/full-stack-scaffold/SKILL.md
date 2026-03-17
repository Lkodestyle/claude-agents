---
name: full-stack-scaffold
description: Use when scaffolding a new full-stack project from scratch, creating a web or mobile application with API and database. Orchestrates architecture, web/mobile, databases, programming, and docker agents for complete project setup. MUST BE USED when user mentions new project, scaffold, starter, boilerplate, MVP setup, or full-stack application creation.
---

# Full-Stack Scaffold — Complete Project Setup

Scaffold a complete full-stack project with architecture, database, API, frontend, containerization, and testing infrastructure.

## When to Use

- Starting a new full-stack project from scratch
- Creating an MVP with API + frontend + database
- Setting up a project with proper structure and best practices
- Scaffolding a monorepo or multi-service project

## Agents Activated

| Agent | Role |
|-------|------|
| `architecture` | System design, component relationships, data flow |
| `web` / `mobile` | Frontend and backend framework setup |
| `databases` | Schema design, migrations, ORM setup |
| `programming` | Code patterns, testing setup, code quality |
| `docker` | Containerization, Docker Compose for local dev |

## Workflow

### Step 1: Requirements Gathering

Using the `architecture` agent:

1. Understand the project:
   - What problem does it solve?
   - Who are the users?
   - What are the core features for MVP?
2. Ask the user for:
   - **Platform**: Web / Mobile / Both
   - **Frontend**: React, Next.js, Vue, Nuxt, React Native, Flutter
   - **Backend**: Node.js (Express/Fastify), Python (FastAPI/Django), Go
   - **Database**: PostgreSQL, MySQL, MongoDB, SQLite
   - **Auth**: JWT, OAuth, session-based, or third-party (Auth0, Clerk)
   - **Project structure**: Monorepo or separate repos

### Step 2: Architecture Design

Using the `architecture` agent:

1. Define system components and their relationships
2. Design API structure (REST or GraphQL)
3. Define data flow between frontend, API, and database
4. Generate architecture diagram (Mermaid):
   ```mermaid
   graph TD
     Client --> API
     API --> Database
     API --> Cache
     API --> ExternalServices
   ```

### Step 3: Database Schema

Using the `databases` agent:

1. Design the data model based on requirements:
   - Entity identification and relationships
   - Primary keys, foreign keys, indexes
   - Timestamps (created_at, updated_at)
2. Generate initial migration files
3. Set up ORM configuration:
   - **Node.js**: Prisma / TypeORM / Drizzle
   - **Python**: SQLAlchemy / Django ORM
   - **Go**: GORM / sqlc
4. Create seed data for development

### Step 4: Backend API

Using the `web` and `programming` agents:

1. Scaffold API project:
   - Project structure with clean architecture (routes, controllers, services, repositories)
   - Environment configuration (.env.example)
   - Error handling middleware
   - Request validation
   - Logging setup
2. Implement core endpoints:
   - Health check (`GET /health`)
   - Auth endpoints (if applicable)
   - CRUD for main entities
3. Set up API documentation (Swagger/OpenAPI)

### Step 5: Frontend Application

Using the `web` or `mobile` agent:

1. Scaffold frontend project:
   - Component structure (pages, components, hooks, utils)
   - Routing configuration
   - State management setup
   - API client configuration
2. Create base components:
   - Layout (header, sidebar, main content)
   - Auth pages (login, register) if applicable
   - Main entity pages (list, detail, create, edit)
3. Configure styling (Tailwind CSS / CSS Modules / styled-components)

### Step 6: Testing Infrastructure

Using the `programming` agent:

1. **Backend tests**:
   - Unit test setup (Jest / Pytest / Go testing)
   - API integration test examples
   - Test database configuration
2. **Frontend tests**:
   - Component test setup (Testing Library)
   - E2E test setup (Playwright / Cypress) — config only
3. Generate test examples for core functionality

### Step 7: Containerization

Using the `docker` agent:

1. Create `Dockerfile` for backend (multi-stage, optimized)
2. Create `Dockerfile` for frontend (if separate)
3. Create `docker-compose.yml` with:
   - App service(s)
   - Database service (PostgreSQL/MySQL/MongoDB)
   - Cache service (Redis) if needed
   - Volumes for data persistence
   - Environment variables
4. Create `Makefile` or scripts for common operations

### Step 8: Developer Experience

1. Configure code quality tools:
   - Linter (ESLint / Ruff / golangci-lint)
   - Formatter (Prettier / Black / gofmt)
   - Pre-commit hooks (husky + lint-staged)
2. Create comprehensive `.gitignore`
3. Generate README.md with:
   - Project description
   - Quick start (docker-compose up)
   - Development setup
   - API documentation link
   - Project structure explanation

## Output Structure

```
project/
├── docker-compose.yml
├── Makefile
├── README.md
├── .gitignore
├── .env.example
├── backend/
│   ├── Dockerfile
│   ├── src/
│   │   ├── routes/
│   │   ├── controllers/
│   │   ├── services/
│   │   ├── repositories/
│   │   ├── models/
│   │   ├── middleware/
│   │   ├── config/
│   │   └── utils/
│   ├── tests/
│   ├── migrations/
│   └── package.json (or requirements.txt, go.mod)
├── frontend/
│   ├── Dockerfile
│   ├── src/
│   │   ├── pages/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── services/
│   │   ├── utils/
│   │   └── styles/
│   ├── tests/
│   └── package.json
└── docs/
    └── architecture.md
```

## Key Principles

- **Convention over configuration**: Follow framework defaults unless there's a good reason not to
- **12-factor app**: Environment-based config, stateless processes, port binding
- **Start small**: Only scaffold what's needed for MVP, avoid premature complexity
- **Docker-first dev**: `docker-compose up` should be the only command to start developing
- **Tests from day one**: Testing infrastructure set up, even if coverage is minimal initially
- **Clean architecture**: Separation of concerns from the start, easy to extend later
