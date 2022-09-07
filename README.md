# Todo Backend
Todo (Backend)

## Build

```
make run/app
```

```
make build/app
```

build for linux systems linux/amd64
```
make build/all
```

## Documentation
Regenerate swagger documentation

```
make swagger
```

## Migrations

```
sql-migrate new create_user_table
```

```
bin migrate up
bin migrate down

make migrate/up
make migrate/down
```

