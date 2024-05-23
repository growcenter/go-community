# About The Project

Project by GROW IT Team for Church Community Dashboard

---

## Tech Stack

- Written in go 1.21.4 or latest

### Tools

- Docker
- Docker Compose

### HTTP Framework

- https://github.com/labstack/echo

### Driver Packages

- GORM https://github.com/go-gorm/gorm

### Testing Packages

### Additional Packages

- Libraries for configuration parsing https://github.com/spf13/viper
- Validator https://github.com/go-playground/validator
- SQL ORM https://github.com/go-gorm/gorm
- Swagger Doc https://github.com/swaggo/swag
- UUID https://github.com/google/uuid
- Linter https://github.com/golangci/golangci-lint
- List of go frameworks & libraries https://github.com/avelino/awesome-go

---

## HOW TO RUN

clone the project inside GitHub repository `https://github.com/growcenter`

```bash
git clone
```

To run this service, you need to add configuration file

```bash
cp config/config.local.example.yaml config/config.local.yaml
```

This service already uses `go.mod`. `make tidy` will simply get all dependencies.

### Run Service

1. Run `make docker-start`
2. Run `make database-up`
3. Run `make run-api`

### Generate Swagger - SOON

### Unit Test - SOON

the scope of mandatory unit testing is:

```
internal/deliveries
internal/repositories
internal/services
```

naming convention test table

```bash
success case - define success case
error case - define error case
```

#### Unit Test - SOON

Run `make test`

#### Unit Test Show Coverage - SOON

Run `make test-cover-display`

### Run Before Create PR - SOON

Run `make prepare-release`

### Run Linter - SOON

1. Install requirement `make lint-prepare`
2. Run linter `make lint`
3. Result in `lint.xml`

---
