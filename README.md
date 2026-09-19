# ads-b-trace-analyzer

## Migrations

SQL migrations live in `migrations/` as numbered `up`/`down` pairs. Both options below track the applied version in the same `schema_migrations` table, so they can be used interchangeably.

Database credentials are read from `envs/.local.env` (see `envs/.example.env` for the required variables). Run all commands from the repo root.

### Option 1: `-migration` flag

Applies all pending migrations using the connection settings in `envs/.local.env`:

```sh
go run . -migration
```

- Already up to date is not treated as an error.
- After migrating, the program continues with the normal analysis run (`-in traces` by default). If that input path doesn't exist it exits with an error, but the migrations have already been applied.

### Option 2: `migrate` CLI

The devcontainer ships the [golang-migrate](https://github.com/golang-migrate/migrate) CLI. Use it when you need more than applying migrations. From a bash shell in the devcontainer, load `envs/.local.env` and build the connection URL:

```sh
set -a; . envs/.local.env; set +a
export DATABASE_URL="postgres://$DB_USER:$DB_PASS@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
```

| Task | Command |
| --- | --- |
| Apply all pending migrations | `migrate -path migrations -database "$DATABASE_URL" up` |
| Roll back the last migration | `migrate -path migrations -database "$DATABASE_URL" down 1` |
| Show the current version | `migrate -path migrations -database "$DATABASE_URL" version` |
| Clear a dirty state after a failed migration | `migrate -path migrations -database "$DATABASE_URL" force <version>` |
| Create a new migration pair | `migrate create -ext sql -dir migrations -seq <name>` |

`force` only sets the recorded version; fix or revert the partially applied changes by hand first.
