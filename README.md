# BrewFlow

Coffee shop POS and inventory management system. Backend-first MVP built with clean architecture.

## Tech Stack

- **Language:** Go 1.26+
- **Framework:** Gin
- **ORM:** Bun (`github.com/uptrace/bun`)
- **Database:** PostgreSQL 17
- **Auth:** JWT (HS256) with `jti` claims + token blacklist
- **Migrations:** golang-migrate
- **Dev:** Air (live reload), Task
- **Frontend (planned):** Next.js + TypeScript + Zod + Drizzle

## Architecture

- **Clean architecture:** Handlers → Services → Repositories → Database
- **No business logic in handlers.** Handlers parse request, call service, return DTO.
- **Sentinel errors** per service package, mapped to HTTP status centrally.
- **All money values** stored as BIGINT (cents). No floats.
- **Manual DI** via `internal/app/container.go`.

## Current Scope (MVP — In Progress)

| Domain | Features |
|--------|----------|
| **Auth** | JWT login/logout, token blacklist, role-based access (`owner`, `manager`, `staff`) |
| **Users** | List users |
| **Staff Onboarding** | Send/accept invitation, set password, cancel, list |
| **Categories** | Create, rename, activate/deactivate, list |
| **Products** | Create (auto-SKU), find by ID/SKU, list, price in cents, category link |
| **Inventory** | Stock level per product (model/repo/service ready, HTTP handler pending) |
| **Inventory Movements** | Track `RECEIVED`, `SOLD`, `RETURNED`, `DAMAGED`, `ADJUSTED` (model/repo/service ready, handler pending) |
| **Orders** | Create with line items, list, detail, modify pending items, status lifecycle (`pending` → `completed`/`cancelled`) |
| **Payments** | Record payment (cash/card/mobile), change calculation for cash |
| **Discounts** | Create/list discounts, apply percentage or fixed discount to pending order |

## Roadmap

| Phase | Focus | Status |
|-------|-------|--------|
| **1** | Orders core (create, list, detail, status, item management) | 🔄 In progress |
| **2** | Inventory HTTP handlers | ⏳ Planned |
| **3** | Payments (pay completed order, change calculation) | ⏳ Planned |
| **4** | Discounts (create/list/apply) | ⏳ Planned |
| **5** | Customers CRM | 🔜 Future |
| **6** | Tables / dine-in floor | 🔜 Future |
| **7** | Shifts / cash drawer | 🔜 Future |
| **8** | Receipts (PDF) | 🔜 Future |
| **9** | Sales reports & analytics | 🔜 Future |
| **10** | Multi-location support | 🔜 Future |

## Key Design Decisions

- **Tax:** Backend-calculated from `TAX_RATE` env var (default 0.08), rounded half-up to nearest cent.
- **Payments:** Exact amount required for card/mobile. Cash allows overpayment; change calculated automatically.
- **Inventory:** Deducted on order `completed`, restored on `cancelled`. Atomic transactions.
- **Idempotency:** `POST /api/protected/orders` supports optional `Idempotency-Key` header.

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
├── app/                # Composition root, DI container
├── config/             # Configuration loading
├── database/           # Database connection
├── middleware/         # Auth, error handling, logging
├── server/             # HTTP server, routes
└── services/           # Business logic
    ├── auth/           # JWT + token blacklist
    ├── user/           # User domain
    ├── account/        # Account linking
    ├── invitations/    # Staff onboarding
    ├── categories/     # Product categories
    ├── products/       # Product catalog
    ├── inventory/      # Stock levels
    ├── inventory_movements/ # Stock movements
    ├── orders/         # Orders + order items
    ├── payments/       # Order payments
    └── discounts/      # Discounts management
shared/                 # Pagination, logger, constants
db/migrations/          # golang-migrate SQL files
```

## API Endpoints

### Public
- `POST /api/login`
- `POST /api/logout`
- `POST /api/accept-invitation`
- `POST /api/set-password`

### Protected (Owner/Manager)
- `POST /api/protected/invitations/send`
- `POST /api/protected/invitations/:id/cancel`
- `GET /api/protected/invitations/`
- `POST /api/protected/categories/add`
- `POST /api/protected/categories/:id/update`
- `POST /api/protected/categories/:id/status`
- `GET /api/protected/categories/`
- `POST /api/protected/products/add`
- `GET /api/protected/products/`
- `GET /api/protected/products/:id`
- `GET /api/protected/products/sku/:sku`
- `POST /api/protected/discounts`
- `GET /api/protected/discounts`

### Protected (Staff+)
- `POST /api/protected/orders`
- `GET /api/protected/orders`
- `GET /api/protected/orders/:id`
- `POST /api/protected/orders/:id/status`
- `POST /api/protected/orders/:id/items`
- `DELETE /api/protected/orders/:id/items/:itemId`
- `PATCH /api/protected/orders/:id/items/:itemId`
- `POST /api/protected/orders/:id/pay`
- `POST /api/protected/orders/:id/discount`
- `GET /api/protected/inventory`
- `GET /api/protected/inventory/:productId`
- `POST /api/protected/inventory-movements`
- `GET /api/protected/inventory-movements`

## Contributing

- Keep business logic in services, not handlers
- Return DTOs, not DB models
- All migrations must be reversible
- Run `task migrate-up` after adding migrations

## License

MIT
