# Teamy Backend

## Prerequisites

- [Go v1.17](https://golang.org/dl)

## Getting Started

1. Start Server

   ```bash
   go run main.go
   ```

2. Visit `http://localhost:8080/random`

## Deployment Environments

- [testing](https://teamy-backend.testing.teamyapp.com)
- [staging](https://teamy-backend.staging.teamyapp.com)
- [production](https://teamy-backend.teamyapp.com)

## DB

### Generate SQL to create new DB

```bash
go run cli/* db new -n [dbName]
```

Eg.

```bash
go run cli/* db new -n testing
```

### Seed

```bash
go run cli/* db migrate --steps [steps]
```

Eg.

```bash
go run cli/* db migrate --steps 1
```

### Migrate

```bash
ggo run cli/* db seed
```
