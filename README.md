i will use this to run data processing data pipeline on a 2025 fligh data dataset

the overall purpose of this project is being able to serve some kind of aviation rewind type of thing using some nest js api with a simple performant svelte frontend

local database:

```
docker compose --env-file envs/.local.env -f docker/docker-compose.postgres.yml up -d

docker compose -f docker/docker-compose.postgres.yml down

sudo rm -rf docker/pgdata
```

## migrations

files must follow the pattern `000001_name.up.sql` / `000001_name.down.sql`

if a migration fails dirty:
```
migrate -path ./migrations -database "postgres://user:pass@localhost:5432/adsb_analytics?sslmode=disable" force <version>

migrate -path ./migrations -database "postgres://user:pass@localhost:5432/adsb_analytics?sslmode=disable" down 1

migrate -path ./migrations -database "$MIGRATE_DB" version
no migration means the down worked 

```

`down N` rolls back N steps, not down to version N.