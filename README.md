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

## Migrate DB

```bash
./script/migrate_db.sh [user] [password] [dbName] [up/down] [steps]
```

Eg.

```bash
./script/migrate_db.sh postgres password teamy up 1
```
