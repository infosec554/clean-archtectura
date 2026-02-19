# clean-archtectura

Go Clean Architecture Template

## Stack
- Echo v4
- PostgreSQL
- Redis
- JWT Auth
- Swagger
- Docker

## Structure
```
├── app/          # main.go
├── config/       # configuration
├── domain/       # entities & interfaces
├── internal/
│   ├── repository/   # DB layer
│   └── rest/         # HTTP handlers
├── migrations/   # SQL migrations
├── pkg/          # shared utilities
└── service/      # business logic
```
