# BrewFlow

Coffee shop POS and inventory management system.

## Backend

**Stack**
- Go 1.26+
- Gin
- Bun ORM
- PostgreSQL 16
- JWT auth with token blacklist
- golang-migrate

**Principles**
- Handlers → Services → Repositories → Database
- JWT tokens with `jti` claims for logout/blacklist
- Transactions for multi-step writes
- Role-based access control

## Frontend

Next.js + TypeScript + Zod + Drizzle

## Prerequisites

- Go 1.26+
- Docker & Docker Compose
- PostgreSQL (via Docker)

## Getting Started

```bash
# Start PostgreSQL
task docker-up

# Run migrations
task migrate-up

# Start dev server
task dev
```

API available at `http://localhost:8080/health`

## Project Structure

```
internal/
├── app/            # Composition root, DI container
├── config/         # Configuration loading
├── database/       # Database connection
├── middleware/     # HTTP middleware (auth, logging)
├── server/         # HTTP server, routes, handlers
└── services/       # Business logic
    ├── auth/       # JWT + token blacklist
    ├── user/       # User domain
    ├── account/    # Account linking
    └── invitation/ # Invitation onboarding
```

## Available Tasks

```bash
task setup              # Install dev tools
task dev                # Start server with live reload
task docker-up          # Start PostgreSQL
task docker-down        # Stop PostgreSQL
task migrate-up         # Apply migrations
task migrate-down       # Rollback last migration
task migrate-create NAME=create_products_table  # New migration
```

## API

### Public
- `POST /api/login`
- `POST /api/logout`

### Protected (Owner/Manager)
- `POST /api/invitations/send`
- `POST /api/invitations/:id/cancel`
- `GET /api/invitations/`
- `POST /api/categories`
- `PUT /api/categories/:id`
- `DELETE /api/categories/:id`
- `POST /api/products`
- `PUT /api/products/:id`
- `DELETE /api/products/:id`
- `GET /api/products/`

### Protected (Staff+)
- `POST /api/orders`
- `GET /api/orders/`
- `GET /api/orders/:id`

## Contributing

- Keep business logic in services, not handlers
- Use Bun models in repositories
- Return DTOs, not DB models
- All migrations must be reversible

## License

MIT
