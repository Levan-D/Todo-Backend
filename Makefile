.PHONY:
.SILENT:
.DEFAULT_GOAL := info

buildInfo = -ldflags "-X 'todo-backend/pkg/config.BuildVersion=`git tag --sort=-version:refname | head -n 1`' -X 'todo-backend/pkg/config.BuildTime=${shell date -u}'"

info:
	echo "Todo Platform"

run/app:
	swag init -g ./cmd/app/main.go -o ./docs/app
	go run cmd/app/main.go

build:
	swag init -g ./cmd/app/main.go -o ./docs/app
	packr2 build ${buildInfo} -o ../../.bin/todo-app cmd/app/main.go

build/app:
	swag init -g ./cmd/app/main.go -o ./docs/app
	packr2 build ${buildInfo} -o ../../.bin/todo-app cmd/app/main.go

build/linux:
	swag init -g ./cmd/app/main.go -o ./docs/app
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 packr2 build ${buildInfo} -o ../../.bin/todo-app cmd/app/main.go

swagger:
	swag init -g ./cmd/app/main.go -o ./docs/app

migrate/up:
	go run cmd/app/main.go migrate up

migrate/down:
	go run cmd/app/main.go migrate down
