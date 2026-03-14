# site

this repo includes a..

- Go API, task handling, and sqlite3 database in [/api](api)
- Svelte + TypeScript frontend in [/web](web)

development tracking currently takes place within a private Linear workspace, but help is welcome! contact me at [spiritov_v@pm.me](mailto:spiritov_v@pm.me)

- api docs are provided at `/docs` (see [.env.local.example](api/env/.env.local.example?plain=1#L17))
- [asynq](https://github.com/hibiken/asynq/) with Redis is used for distributed task handling
- [migrate cli](https://github.com/golang-migrate/migrate) is used to manage sql migrations
- [sqlc](https://github.com/sqlc-dev/sqlc) is used to generate Go code from sql
- [openapi-typescript](https://github.com/openapi-ts/openapi-typescript) is used to generate types from the api's schema
- UI is originally derived from [this repo](https://github.com/spiritov/ui)

## quickstart with empty database
1. populate `.env.local` in both `/api` and `/web` using `.env.local.example` in each as a template
2. create empty sqlite3 db, Redis db, run api
```sh
cd api
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate
migrate -source file://db/migrations -database sqlite3://db/jump.db up
docker run -d -p 6379:6379 redis:latest
go run .
```
3. run web
```sh
cd web
npm i
npm run dev
```

> [!NOTE]
> most API endpoints require a session cookie, and many require an elevated role stored in the player table. for a local database, set your role to "dev"

## migrations

changes requiring a database schema change use migrations, see existing migrations to get an idea of what they should look like

```sh
cd api
migrate create -ext sql -dir db/migrations change-summary
# move up to next migration version
migrate -source file://db/migrations -database sqlite3://db/jump.db up 1
```
