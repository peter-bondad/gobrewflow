# BrewFlow

A modern coffee shop management system focused on operational efficiency, inventory accuracy, and a seamless ordering experience.

## Tech Stack

### Backend

- **Runtime:** Go 1.26.5
- **Framework:** Gin
- **ORM:** Bun
- **Database:** PostgreSQL 16
- **Authentication:** go-jwt
- **Migrations:** golang-migrate
- **Task Runner:** go-task
- **Live Reload:** air

### Admin Dashboard

- React
- TypeScript
- Tailwind CSS v4
- shadcn/ui
- Base UI

### Mobile (Staff Ordering)

- React Native
- Expo
- TypeScript
- NativeWind
- TanStack Query
- React Hook Form
- Zod

## Prerequisites

- Go 1.26+
- Docker & Docker Compose
- Homebrew (macOS)
- PostgreSQL (via Docker)

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/your-username/gobrewflow.git
cd gobrewflow
```

### 2. Install development tools

```bash
task setup
```

This installs:

- `golang-migrate` CLI
- `air` live reload tool

### 3. Start PostgreSQL

```bash
task docker-up
```

### 4. Run migrations

```bash
task migrate-up
```

### 5. Start the development server

```bash
task dev
```

The API will be available at `http://localhost:8080`

### 6. Verify health

```bash
curl http://localhost:8080/health
```

## Project Structure

```
gobrewflow/
├── cmd/api/                    # Application entrypoint
├── db/migrations/              # Database migrations
├── internal/
│   ├── app/                    # Application composition root
│   │   └── container.go        # Dependency injection container
│   ├── config/                 # Configuration loading & validation
│   ├── database/               # Database connection setup
│   ├── middleware/             # HTTP middleware (auth, logging)
│   ├── server/                 # HTTP server, routes, handlers
│   └── services/               # Business logic layer
│       ├── user/               # User domain
│       ├── account/            # OAuth/credentials account linking
│       └── invitation/         # User invitations
└── shared/                     # Shared utilities (logger)
```

## Available Tasks

```bash
task setup              # Install dev tools (migrate, air)
task dev                # Start development server with live reload
task docker-up          # Start PostgreSQL in background
task docker-down        # Stop PostgreSQL and remove volumes
task migrate-up         # Apply pending migrations
task migrate-down       # Rollback last migration
task migrate-create NAME=create_orders_table  # Scaffold new migration
task migrate-force VERSION=1  # Force migration version (emergency only)
```

## Architecture

### Clean Architecture

The project follows Clean Architecture principles with clear separation of concerns:

- **Handlers** — HTTP layer, request/response mapping, validation
- **Services** — Business logic, orchestration, authorization
- **Repositories** — Data access, SQL queries, Bun models
- **Models** — Bun ORM models, database schema definitions

### Dependency Injection

Dependencies flow inward through constructor injection:

```
main.go
  └── App
       └── Server
            └── Routes
                 └── Handlers
                      └── Services
                           └── Repositories
                                └── Database
```

### Authentication Flow

1. Client sends credentials or OAuth callback
2. Handler validates input
3. Service creates/retrieves account
4. Session is created and stored in database
5. Session token is returned to client
6. Protected routes validate session via middleware

### Data Model

```
users
  └── accounts (1:N)          # OAuth providers / password hash
  └── sessions (1:N)          # Active sessions
  └── invitations (1:N)       # Invitation records
```

## Roadmap

### Phase 1 — Core Platform 🚧 In Progress

- [x] Authentication & Session Management
- [x] Role Permissions (`owner`, `manager`, `staff`)
- [x] Dashboard Overview & Inventory Alerts
- [ ] Recent Orders (real backend feed)
- [ ] Top Selling Products (real backend feed)
- [x] Products CRUD (Admin)
- [ ] Product Menu Endpoint (Mobile)
- [x] Inventory Management
- [x] Orders (Admin + Mobile)
- [x] Suppliers (Backend Schema)
- [ ] User Invitations

### Next: React Native Staff Ordering App

- Token-based authentication
- Browse/search products
- Create orders
- View/update order status

### Future

- POS System
- Analytics & Reporting
- Employee Management
- Multi-Branch Support
- Customer Loyalty
- Online Payments

## Design Principles

- Simplicity over complexity
- Type safety (Go + TypeScript)
- SOLID principles
- Clean Architecture
- Feature-first organization
- Reusable components
- Consistent UI/UX
- Performance & Scalability
- Auditability

## Contributing

Contributions are welcome. Please ensure:

1. All migrations are reversible (`.up.sql` + `.down.sql`)
2. Business logic lives in services, not handlers
3. Database queries use Bun models in repositories
4. API responses use DTOs, not database models
5. Errors are wrapped with context

## License

MIT
